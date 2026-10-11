package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/security"
	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrValidation         = errors.New("validation error")
	ErrConflict           = errors.New("conflict")
)

type AuthInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResult struct {
	AccessToken  string
	ExpiresIn    int64
	RefreshToken string
	User         domain.User
}

type AuthService struct {
	db         *sql.DB
	users      *repository.Users
	sessions   *repository.Sessions
	activities *repository.ActivityLogs
	tokens     *security.TokenService
	cfg        config.Config
	logger     *slog.Logger
	dummyHash  string
	clock      func() time.Time
}

func NewAuthService(db *sql.DB, cfg config.Config, logger *slog.Logger) (*AuthService, error) {
	dummyHash, err := security.NewDummyPasswordHash(cfg.BcryptCost)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthService{
		db: db, users: repository.NewUsers(db), sessions: repository.NewSessions(db), activities: repository.NewActivityLogs(db),
		tokens: security.NewTokenService(cfg.JWTSecret, "bem-fmikom-api", cfg.AccessTokenTTL), cfg: cfg, logger: logger,
		dummyHash: dummyHash, clock: time.Now,
	}, nil
}

func (service *AuthService) Login(ctx context.Context, input AuthInput, ipAddress, userAgent string) (AuthResult, error) {
	email, err := NormalizeEmail(input.Email)
	if err != nil || input.Password == "" {
		return AuthResult{}, fmt.Errorf("%w: email and password are required and email must be valid", ErrValidation)
	}
	user, findErr := service.users.FindByEmail(ctx, email)
	if findErr != nil && !errors.Is(findErr, repository.ErrNotFound) {
		return AuthResult{}, findErr
	}
	passwordHash := service.dummyHash
	if findErr == nil {
		passwordHash = user.PasswordHash
	}
	passwordOK := security.VerifyPassword(passwordHash, input.Password)
	if findErr != nil || !passwordOK || !HasValidAuthorizationState(user) {
		service.logger.Warn("auth_login_denied", slog.String("source_ip", ipAddress))
		return AuthResult{}, ErrInvalidCredentials
	}

	now := service.clock().UTC()
	refreshToken, err := security.NewRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}
	familyID, err := security.NewSessionFamilyID()
	if err != nil {
		return AuthResult{}, err
	}
	accessToken, expiresAt, err := service.tokens.IssueAccess(user.ID, now)
	if err != nil {
		return AuthResult{}, err
	}
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return AuthResult{}, fmt.Errorf("begin login transaction")
	}
	defer func() { _ = tx.Rollback() }()
	currentUser, err := service.users.FindByIDTx(ctx, tx, user.ID)
	if err != nil || !HasValidAuthorizationState(currentUser) {
		return AuthResult{}, ErrInvalidCredentials
	}
	if err := service.users.SetLastLogin(ctx, tx, currentUser.ID, now); err != nil {
		return AuthResult{}, err
	}
	if err := service.sessions.Create(ctx, tx, repository.RefreshSession{UserID: currentUser.ID, FamilyID: familyID, ExpiresAt: now.Add(service.cfg.RefreshTokenTTL)}, security.HashRefreshToken(refreshToken), ipAddress, userAgent); err != nil {
		return AuthResult{}, err
	}
	actorID := currentUser.ID
	if err := service.activities.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &actorID, Action: "LOGIN", EntityType: "AUTH_SESSION", IPAddress: ipAddress, UserAgent: userAgent}); err != nil {
		return AuthResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return AuthResult{}, fmt.Errorf("commit login transaction")
	}
	currentUser.LastLoginAt = &now
	currentUser.Permissions = PermissionsForRole(currentUser.Role)
	return AuthResult{AccessToken: accessToken, ExpiresIn: int64(expiresAt.Sub(now).Seconds()), RefreshToken: refreshToken, User: currentUser}, nil
}

func (service *AuthService) Refresh(ctx context.Context, refreshToken, ipAddress, userAgent string) (AuthResult, error) {
	if !security.ValidateRefreshToken(refreshToken) {
		return AuthResult{}, ErrUnauthenticated
	}
	now := service.clock().UTC()
	newRawToken, err := security.NewRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return AuthResult{}, fmt.Errorf("begin refresh transaction")
	}
	defer func() { _ = tx.Rollback() }()

	oldSession, err := service.sessions.FindByHashForUpdate(ctx, tx, security.HashRefreshToken(refreshToken))
	if errors.Is(err, repository.ErrNotFound) {
		service.logger.Warn("auth_refresh_rejected", slog.String("reason", "unknown_token"), slog.String("source_ip", ipAddress))
		return AuthResult{}, ErrUnauthenticated
	}
	if err != nil {
		return AuthResult{}, err
	}
	if oldSession.RevokedAt.Valid {
		if err := service.sessions.RevokeFamily(ctx, tx, oldSession.FamilyID, now); err != nil {
			return AuthResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return AuthResult{}, fmt.Errorf("commit refresh reuse response")
		}
		service.logger.Error("auth_refresh_token_reuse", slog.Int64("user_id", oldSession.UserID), slog.String("source_ip", ipAddress))
		return AuthResult{}, ErrUnauthenticated
	}
	if !oldSession.ExpiresAt.After(now) {
		if err := service.sessions.RevokeFamily(ctx, tx, oldSession.FamilyID, now); err != nil {
			return AuthResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return AuthResult{}, fmt.Errorf("commit expired refresh response")
		}
		service.logger.Warn("auth_refresh_rejected", slog.String("reason", "expired"), slog.Int64("user_id", oldSession.UserID), slog.String("source_ip", ipAddress))
		return AuthResult{}, ErrUnauthenticated
	}
	user, err := service.users.FindByIDTx(ctx, tx, oldSession.UserID)
	if err != nil || !HasValidAuthorizationState(user) {
		if revokeErr := service.sessions.RevokeFamily(ctx, tx, oldSession.FamilyID, now); revokeErr != nil {
			return AuthResult{}, revokeErr
		}
		if err := tx.Commit(); err != nil {
			return AuthResult{}, fmt.Errorf("commit inactive refresh response")
		}
		service.logger.Warn("auth_refresh_rejected", slog.String("reason", "inactive_or_invalid_account"), slog.Int64("user_id", oldSession.UserID), slog.String("source_ip", ipAddress))
		return AuthResult{}, ErrUnauthenticated
	}
	if err := service.sessions.Revoke(ctx, tx, oldSession.ID, now); err != nil {
		return AuthResult{}, err
	}
	rotatedFrom := sql.NullInt64{Int64: oldSession.ID, Valid: true}
	if err := service.sessions.Create(ctx, tx, repository.RefreshSession{UserID: user.ID, FamilyID: oldSession.FamilyID, ExpiresAt: now.Add(service.cfg.RefreshTokenTTL), RotatedFrom: rotatedFrom}, security.HashRefreshToken(newRawToken), ipAddress, userAgent); err != nil {
		return AuthResult{}, err
	}
	accessToken, expiresAt, err := service.tokens.IssueAccess(user.ID, now)
	if err != nil {
		return AuthResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return AuthResult{}, fmt.Errorf("commit refresh rotation")
	}
	user.Permissions = PermissionsForRole(user.Role)
	return AuthResult{AccessToken: accessToken, ExpiresIn: int64(expiresAt.Sub(now).Seconds()), RefreshToken: newRawToken, User: user}, nil
}

func (service *AuthService) Logout(ctx context.Context, refreshToken string, authenticatedUserID *int64, ipAddress, userAgent string) error {
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin logout transaction")
	}
	defer func() { _ = tx.Rollback() }()
	actorID := int64(0)
	if refreshToken != "" {
		actorID, err = service.sessions.RevokeHash(ctx, tx, security.HashRefreshToken(refreshToken), service.clock().UTC())
		if err != nil {
			return err
		}
	}
	if actorID == 0 && authenticatedUserID != nil {
		actorID = *authenticatedUserID
	}
	if actorID > 0 {
		if err := service.activities.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &actorID, Action: "LOGOUT", EntityType: "AUTH_SESSION", IPAddress: ipAddress, UserAgent: userAgent}); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit logout transaction")
	}
	return nil
}

func (service *AuthService) CurrentUser(ctx context.Context, rawToken string) (domain.User, error) {
	userID, err := service.tokens.ParseAccess(rawToken, service.clock().UTC())
	if err != nil {
		return domain.User{}, ErrUnauthenticated
	}
	user, err := service.users.FindByID(ctx, userID)
	if err != nil || !HasValidAuthorizationState(user) {
		return domain.User{}, ErrUnauthenticated
	}
	user.Permissions = PermissionsForRole(user.Role)
	return user, nil
}

func (service *AuthService) CreateSuperAdmin(ctx context.Context, name, email, password string) (int64, error) {
	return CreateSuperAdmin(ctx, service.db, service.cfg.BcryptCost, name, email, password)
}

func CreateSuperAdmin(ctx context.Context, db *sql.DB, bcryptCost int, name, email, password string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 160 {
		return 0, fmt.Errorf("name must contain 1 to 160 characters")
	}
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil {
		return 0, err
	}
	hash, err := security.HashPassword(password, bcryptCost)
	if err != nil {
		return 0, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin admin creation transaction")
	}
	defer func() { _ = tx.Rollback() }()
	id, err := repository.NewUsers(db).Create(ctx, tx, domain.User{Name: name, Email: normalizedEmail, PasswordHash: hash, Role: domain.RoleSuperAdmin, Status: domain.StatusActive})
	if err != nil {
		var mysqlError *mysql.MySQLError
		if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
			return 0, ErrConflict
		}
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit admin creation transaction")
	}
	return id, nil
}

func NormalizeEmail(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" || len(normalized) > 254 {
		return "", fmt.Errorf("email is required and must be valid")
	}
	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized || !strings.Contains(normalized, "@") {
		return "", fmt.Errorf("email is required and must be valid")
	}
	return normalized, nil
}
