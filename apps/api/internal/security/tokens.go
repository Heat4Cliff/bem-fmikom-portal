package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	jwt.RegisteredClaims
}

type TokenService struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

func NewTokenService(secret []byte, issuer string, accessTTL time.Duration) *TokenService {
	return &TokenService{secret: append([]byte(nil), secret...), issuer: issuer, accessTTL: accessTTL}
}

func (service *TokenService) IssueAccess(userID int64, now time.Time) (string, time.Time, error) {
	var jtiBytes [16]byte
	if _, err := rand.Read(jtiBytes[:]); err != nil {
		return "", time.Time{}, fmt.Errorf("generate access token id")
	}
	expiresAt := now.Add(service.accessTTL)
	claims := AccessClaims{RegisteredClaims: jwt.RegisteredClaims{
		Issuer: service.issuer, Subject: strconv.FormatInt(userID, 10),
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expiresAt),
		ID: hex.EncodeToString(jtiBytes[:]),
	}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	raw, err := token.SignedString(service.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token")
	}
	return raw, expiresAt, nil
}

func (service *TokenService) ParseAccess(raw string, now time.Time) (int64, error) {
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return service.secret, nil
	}, jwt.WithIssuer(service.issuer), jwt.WithTimeFunc(func() time.Time { return now }))
	if err != nil || token == nil || !token.Valid {
		return 0, fmt.Errorf("invalid access token")
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID < 1 || claims.ID == "" {
		return 0, fmt.Errorf("invalid access token subject")
	}
	return userID, nil
}

func NewRefreshToken() (string, error) {
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", fmt.Errorf("generate refresh token")
	}
	return base64.RawURLEncoding.EncodeToString(token[:]), nil
}

func HashRefreshToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}

func NewSessionFamilyID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate refresh session family id")
	}
	return hex.EncodeToString(value[:]), nil
}

func ValidateRoleDivision(role domain.Role, divisionID *int64) error {
	switch role {
	case domain.RoleSuperAdmin, domain.RoleBPH:
		if divisionID != nil {
			return fmt.Errorf("%s must not be assigned to a division", role)
		}
	case domain.RoleKepala:
		if divisionID == nil || *divisionID < 1 {
			return fmt.Errorf("KEPALA_BIDANG must be assigned to exactly one division")
		}
	default:
		return fmt.Errorf("role is invalid")
	}
	return nil
}
