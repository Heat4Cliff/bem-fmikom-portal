package router

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/database"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/database/migrate"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/security"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func TestDatabaseIntegration(t *testing.T) {
	testURL := os.Getenv("TEST_DATABASE_URL")
	if testURL == "" {
		_, sourceFile, _, _ := runtime.Caller(0)
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", "..", "..", "..", ".."))
		_ = godotenv.Load(filepath.Join(projectRoot, ".env"))
		testURL = os.Getenv("TEST_DATABASE_URL")
	}
	if testURL == "" {
		t.Skip("set TEST_DATABASE_URL or copy .env.example to .env for a dedicated MySQL/MariaDB database ending in _test")
	}
	dsn, err := mysql.ParseDSN(testURL)
	if err != nil {
		t.Fatalf("invalid TEST_DATABASE_URL DSN: %v", err)
	}
	if !strings.HasSuffix(strings.ToLower(dsn.DBName), "_test") {
		t.Fatalf("refusing destructive integration test against database %q; database name must end in _test", dsn.DBName)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	db, err := database.OpenWithSettings(ctx, config.DatabaseSettings{URL: testURL, MaxOpenConns: 8, MaxIdleConns: 2, ConnMaxLifetime: 5 * time.Minute})
	if err != nil {
		t.Fatalf("connect integration database: %v", err)
	}
	defer db.Close()
	if err := migrate.DownToZero(ctx, db); err != nil {
		t.Fatalf("reset dedicated test database: %v", err)
	}
	if err := migrate.Up(ctx, db); err != nil {
		t.Fatalf("apply migrations from empty database: %v", err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if err := migrate.DownToZero(cleanupCtx, db); err != nil {
			t.Errorf("clean dedicated test database: %v", err)
		}
	}()
	var tableCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name IN ('periods','divisions','users','refresh_sessions','activity_logs')`, dsn.DBName).Scan(&tableCount); err != nil || tableCount != 5 {
		t.Fatalf("foundation schema table count = %d, err = %v", tableCount, err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback foundation migration: %v", err)
	}
	if err := migrate.Up(ctx, db); err != nil {
		t.Fatalf("reapply foundation migration after rollback: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO periods (name,start_date,end_date,status) VALUES ('2025/2026','2025-01-01','2026-12-31','ACTIVE')`); err != nil {
		t.Fatalf("insert active period: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO periods (name,start_date,end_date,status) VALUES ('2026/2027','2026-01-01','2027-12-31','ACTIVE')`); err == nil {
		t.Fatal("single-active-period database invariant accepted a second ACTIVE period")
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO periods (name,start_date,end_date,status) VALUES ('Archived','2024-01-01','2024-12-31','ARCHIVED')`); err != nil {
		t.Fatalf("insert archived period: %v", err)
	}
	var archivedPeriodID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM periods WHERE name='Archived'`).Scan(&archivedPeriodID); err != nil {
		t.Fatal(err)
	}
	divisionResult, err := db.ExecContext(ctx, `INSERT INTO divisions (period_id,name,status) VALUES (?,'Akademik','ACTIVE')`, archivedPeriodID)
	if err != nil {
		t.Fatalf("insert division: %v", err)
	}
	divisionID, err := divisionResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	divisionBResult, err := db.ExecContext(ctx, `INSERT INTO divisions (period_id,name,status) VALUES (?,'Teknologi','ACTIVE')`, archivedPeriodID)
	if err != nil {
		t.Fatal(err)
	}
	divisionBID, err := divisionBResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.CreateSuperAdmin(ctx, db, 4, "Portal Admin", " ADMIN@EXAMPLE.ORG ", "a secure integration passphrase"); err != nil {
		t.Fatalf("create initial admin: %v", err)
	}
	if _, err := service.CreateSuperAdmin(ctx, db, 4, "Duplicate Admin", "admin@example.org", "a secure integration passphrase"); !errors.Is(err, service.ErrConflict) {
		t.Fatalf("duplicate create-admin error = %v, want conflict", err)
	}
	kepalaHash, err := security.HashPassword("a secure kepala password", 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users (name,email,password_hash,role,division_id,status) VALUES ('Kepala Akademik','kepala@example.org',?,'KEPALA_BIDANG',?,'ACTIVE')`, kepalaHash, divisionID); err != nil {
		t.Fatal(err)
	}
	kepalaBHash, err := security.HashPassword("a secure kepala B password", 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users (name,email,password_hash,role,division_id,status) VALUES ('Kepala Teknologi','kepala-b@example.org',?,'KEPALA_BIDANG',?,'ACTIVE')`, kepalaBHash, divisionBID); err != nil {
		t.Fatal(err)
	}
	bphHash, err := security.HashPassword("a secure bph password", 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users (name,email,password_hash,role,status) VALUES ('BPH Operasional','bph@example.org',?,'BPH','ACTIVE')`, bphHash); err != nil {
		t.Fatal(err)
	}
	inactiveHash, err := security.HashPassword("an inactive test password", 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users (name,email,password_hash,role,status) VALUES ('Inactive','inactive@example.org',?,'BPH','INACTIVE')`, inactiveHash); err != nil {
		t.Fatal(err)
	}

	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		t.Fatal(err)
	}
	loggerBuffer := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(loggerBuffer, nil))
	cfg := config.Config{
		AppEnv: "test", FrontendOrigins: []string{"http://localhost:3000"}, JWTSecret: secret[:],
		AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: 7 * 24 * time.Hour,
		Cookie:     config.Cookie{Name: "bem_refresh", SameSite: "lax", Path: "/api/v1/auth"},
		BcryptCost: 4, LoginLimitCount: 100, LoginLimitWindow: time.Minute,
	}
	engine, err := NewWithDependencies(logger, cfg, db)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	login := postJSON(engine, "/api/v1/auth/login", `{"email":" Admin@Example.org ","password":"a secure integration passphrase"}`, nil)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", login.Code, login.Body.String())
	}
	var loginBody struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken string      `json:"access_token"`
			ExpiresIn   int64       `json:"expires_in"`
			User        domain.User `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(login.Body.Bytes(), &loginBody); err != nil || !loginBody.Success || loginBody.Data.AccessToken == "" {
		t.Fatalf("parse login payload: %v (%s)", err, login.Body.String())
	}
	if loginBody.Data.ExpiresIn != 900 || loginBody.Data.User.Role != domain.RoleSuperAdmin {
		t.Fatalf("login response has unexpected lifetime/role: %+v", loginBody.Data)
	}
	refreshCookie := responseCookie(t, login, "bem_refresh")
	if !refreshCookie.HttpOnly || refreshCookie.Secure || refreshCookie.Path != "/api/v1/auth" || refreshCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("development refresh cookie settings are incorrect: %+v", refreshCookie)
	}
	if strings.Contains(login.Body.String(), refreshCookie.Value) {
		t.Fatal("raw refresh token appeared in JSON response")
	}
	var storedHash string
	if err := db.QueryRowContext(ctx, `SELECT token_hash FROM refresh_sessions LIMIT 1`).Scan(&storedHash); err != nil {
		t.Fatal(err)
	}
	if storedHash != security.HashRefreshToken(refreshCookie.Value) || storedHash == refreshCookie.Value {
		t.Fatal("database did not store only the deterministic SHA-256 refresh token hash")
	}
	var adminID int64
	var adminPasswordHash, adminRole, adminStatus string
	var adminDivision sql.NullInt64
	var initialLastLogin sql.NullTime
	if err := db.QueryRowContext(ctx, `SELECT id, password_hash, role, status, division_id, last_login_at FROM users WHERE email='admin@example.org'`).Scan(&adminID, &adminPasswordHash, &adminRole, &adminStatus, &adminDivision, &initialLastLogin); err != nil {
		t.Fatal(err)
	}
	if adminPasswordHash == "a secure integration passphrase" || !security.VerifyPassword(adminPasswordHash, "a secure integration passphrase") {
		t.Fatal("bootstrap password was not stored as a verifiable password hash")
	}
	if adminRole != string(domain.RoleSuperAdmin) || adminStatus != string(domain.StatusActive) || adminDivision.Valid || !initialLastLogin.Valid {
		t.Fatalf("bootstrap admin state is incorrect: role=%s status=%s division=%v last_login=%v", adminRole, adminStatus, adminDivision, initialLastLogin)
	}
	var adminSessionCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_sessions WHERE user_id=?`, adminID).Scan(&adminSessionCount); err != nil || adminSessionCount != 1 {
		t.Fatalf("successful login refresh session count=%d err=%v", adminSessionCount, err)
	}
	authService := mustAuthService(t, db, cfg, logger)
	probe := gin.New()
	probe.GET("/super-admin", middleware.RequireAuth(authService), middleware.RequireRoles(domain.RoleSuperAdmin), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	probe.GET("/bph-operational", middleware.RequireAuth(authService), middleware.RequireRoles(domain.RoleBPH), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	if response := getWithToken(probe, "/super-admin", loginBody.Data.AccessToken); response.Code != http.StatusNoContent {
		t.Fatalf("SUPER_ADMIN was denied its protected route: %d %s", response.Code, response.Body.String())
	}
	if response := getWithToken(probe, "/bph-operational", loginBody.Data.AccessToken); response.Code != http.StatusForbidden {
		t.Fatalf("SUPER_ADMIN unexpectedly passed BPH-only route: %d", response.Code)
	}
	var loginActorID int64
	var loginAt time.Time
	if err := db.QueryRowContext(ctx, `SELECT actor_user_id, created_at FROM activity_logs WHERE action='LOGIN' ORDER BY id LIMIT 1`).Scan(&loginActorID, &loginAt); err != nil {
		t.Fatal(err)
	}
	if loginActorID != adminID || loginAt.Location() != time.UTC || loginAt.Before(time.Now().UTC().Add(-time.Minute)) || loginAt.After(time.Now().UTC().Add(time.Minute)) {
		t.Fatalf("LOGIN activity actor or UTC timestamp is incorrect: actor=%d created_at=%s", loginActorID, loginAt.Location())
	}

	bphLogin := postJSON(engine, "/api/v1/auth/login", `{"email":"bph@example.org","password":"a secure bph password"}`, nil)
	if bphLogin.Code != http.StatusOK {
		t.Fatalf("BPH login status = %d body=%s", bphLogin.Code, bphLogin.Body.String())
	}
	bphAccessToken := accessTokenFromRecorder(t, bphLogin)
	if response := getWithToken(probe, "/bph-operational", bphAccessToken); response.Code != http.StatusNoContent {
		t.Fatalf("BPH was denied cross-division operational route: %d", response.Code)
	}
	if response := getWithToken(probe, "/super-admin", bphAccessToken); response.Code != http.StatusForbidden {
		t.Fatalf("BPH passed SUPER_ADMIN-only route: %d", response.Code)
	}
	kepalaLogin := postJSON(engine, "/api/v1/auth/login", `{"email":"kepala@example.org","password":"a secure kepala password"}`, nil)
	if kepalaLogin.Code != http.StatusOK {
		t.Fatalf("KEPALA_BIDANG login status = %d body=%s", kepalaLogin.Code, kepalaLogin.Body.String())
	}
	kepalaAccessToken := accessTokenFromRecorder(t, kepalaLogin)
	kepalaMe := getWithToken(engine, "/api/v1/auth/me", kepalaAccessToken)
	if kepalaMe.Code != http.StatusOK || !strings.Contains(kepalaMe.Body.String(), `"division":{"id":`+fmt.Sprint(divisionID)) {
		t.Fatalf("auth/me did not return current KEPALA_BIDANG division: %d %s", kepalaMe.Code, kepalaMe.Body.String())
	}
	if response := getWithToken(probe, "/super-admin", kepalaAccessToken); response.Code != http.StatusForbidden {
		t.Fatalf("KEPALA_BIDANG passed SUPER_ADMIN-only route: %d", response.Code)
	}
	kepalaBLogin := postJSON(engine, "/api/v1/auth/login", `{"email":"kepala-b@example.org","password":"a secure kepala B password"}`, nil)
	if kepalaBLogin.Code != http.StatusOK {
		t.Fatalf("second KEPALA_BIDANG login status = %d body=%s", kepalaBLogin.Code, kepalaBLogin.Body.String())
	}
	kepalaAUser, err := repository.NewUsers(db).FindByEmail(ctx, "kepala@example.org")
	if err != nil {
		t.Fatal(err)
	}
	kepalaBUser, err := repository.NewUsers(db).FindByEmail(ctx, "kepala-b@example.org")
	if err != nil {
		t.Fatal(err)
	}
	bphUser, err := repository.NewUsers(db).FindByEmail(ctx, "bph@example.org")
	if err != nil {
		t.Fatal(err)
	}
	if !service.HasValidAuthorizationState(kepalaAUser) || !service.HasValidAuthorizationState(kepalaBUser) || !service.HasValidAuthorizationState(bphUser) {
		t.Fatal("database-backed role/division principals should have valid authorization state")
	}
	divisionCases := []struct {
		name     string
		role     domain.Role
		own      *int64
		resource int64
		allowed  bool
	}{
		{"SUPER_ADMIN cross division", domain.RoleSuperAdmin, nil, divisionBID, true},
		{"BPH cross division", bphUser.Role, bphUser.DivisionID, divisionBID, true},
		{"KEPALA_A own division", kepalaAUser.Role, kepalaAUser.DivisionID, divisionID, true},
		{"KEPALA_A other division", kepalaAUser.Role, kepalaAUser.DivisionID, divisionBID, false},
		{"KEPALA_B other division", kepalaBUser.Role, kepalaBUser.DivisionID, divisionID, false},
		{"KEPALA without division", domain.RoleKepala, nil, divisionID, false},
	}
	for _, test := range divisionCases {
		t.Run(test.name, func(t *testing.T) {
			if got := service.CanAccessDivision(test.role, test.own, test.resource); got != test.allowed {
				t.Fatalf("CanAccessDivision(%s, %v, %d)=%v, want %v", test.role, test.own, test.resource, got, test.allowed)
			}
		})
	}
	invalidPrincipalStates := []domain.User{
		{Role: domain.RoleSuperAdmin, DivisionID: &divisionID, Status: domain.StatusActive},
		{Role: domain.RoleBPH, DivisionID: &divisionID, Status: domain.StatusActive},
		{Role: domain.RoleKepala, Status: domain.StatusActive},
	}
	for _, principal := range invalidPrincipalStates {
		if service.HasValidAuthorizationState(principal) {
			t.Fatalf("invalid role/division combination was accepted: %+v", principal)
		}
	}
	if !service.HasValidAuthorizationState(domain.User{Role: domain.RoleKepala, DivisionID: &divisionID, Status: domain.StatusActive}) {
		t.Fatal("KEPALA_BIDANG with one division should have valid authorization state")
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET role='BPH' WHERE email='admin@example.org'`); err != nil {
		t.Fatal(err)
	}
	me := getWithToken(engine, "/api/v1/auth/me", loginBody.Data.AccessToken)
	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), `"role":"BPH"`) {
		t.Fatalf("auth/me did not return current database role: %d %s", me.Code, me.Body.String())
	}

	invalid := postJSON(engine, "/api/v1/auth/login", `{"email":"admin@example.org","password":"incorrect credential"}`, nil)
	unknownEmail := postJSON(engine, "/api/v1/auth/login", `{"email":"missing@example.org","password":"incorrect credential"}`, nil)
	if invalid.Code != http.StatusUnauthorized || unknownEmail.Code != http.StatusUnauthorized || invalid.Body.String() != unknownEmail.Body.String() || strings.Contains(invalid.Body.String(), "access_token") {
		t.Fatalf("login errors disclosed credential existence or issued a token: wrong=%d %s unknown=%d %s", invalid.Code, invalid.Body.String(), unknownEmail.Code, unknownEmail.Body.String())
	}
	var lastLoginAfterFailure sql.NullTime
	var adminSessionsAfterFailure int
	if err := db.QueryRowContext(ctx, `SELECT last_login_at FROM users WHERE id=?`, adminID).Scan(&lastLoginAfterFailure); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_sessions WHERE user_id=?`, adminID).Scan(&adminSessionsAfterFailure); err != nil {
		t.Fatal(err)
	}
	if !lastLoginAfterFailure.Valid || !lastLoginAfterFailure.Time.Equal(initialLastLogin.Time) || adminSessionsAfterFailure != adminSessionCount {
		t.Fatalf("failed logins changed last_login or created sessions: last_login=%v sessions=%d", lastLoginAfterFailure, adminSessionsAfterFailure)
	}
	var loginEventsAfterFailure int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_logs WHERE actor_user_id=? AND action='LOGIN'`, adminID).Scan(&loginEventsAfterFailure); err != nil || loginEventsAfterFailure != 1 {
		t.Fatalf("failed login attempts created product activity rows: count=%d err=%v", loginEventsAfterFailure, err)
	}
	inactive := postJSON(engine, "/api/v1/auth/login", `{"email":"inactive@example.org","password":"an inactive test password"}`, nil)
	if inactive.Code != http.StatusUnauthorized {
		t.Fatalf("inactive account login status = %d body=%s", inactive.Code, inactive.Body.String())
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET status='ACTIVE' WHERE email='inactive@example.org'`); err != nil {
		t.Fatal(err)
	}
	inactiveSessionLogin := postJSON(engine, "/api/v1/auth/login", `{"email":"inactive@example.org","password":"an inactive test password"}`, nil)
	if inactiveSessionLogin.Code != http.StatusOK {
		t.Fatalf("active account test login status = %d body=%s", inactiveSessionLogin.Code, inactiveSessionLogin.Body.String())
	}
	inactiveAccessToken := accessTokenFromRecorder(t, inactiveSessionLogin)
	inactiveRefreshCookie := responseCookie(t, inactiveSessionLogin, "bem_refresh")
	if _, err := db.ExecContext(ctx, `UPDATE users SET status='INACTIVE' WHERE email='inactive@example.org'`); err != nil {
		t.Fatal(err)
	}
	if response := getWithToken(engine, "/api/v1/auth/me", inactiveAccessToken); response.Code != http.StatusUnauthorized {
		t.Fatalf("deactivated user's existing access JWT was accepted: %d", response.Code)
	}
	if response := postWithCookie(engine, "/api/v1/auth/refresh", inactiveRefreshCookie, nil); response.Code != http.StatusUnauthorized || strings.Contains(response.Body.String(), "access_token") {
		t.Fatalf("deactivated user's existing refresh token was accepted: %d %s", response.Code, response.Body.String())
	}

	rotated := postWithCookie(engine, "/api/v1/auth/refresh", refreshCookie, nil)
	if rotated.Code != http.StatusOK {
		t.Fatalf("refresh status = %d body=%s", rotated.Code, rotated.Body.String())
	}
	newCookie := responseCookie(t, rotated, "bem_refresh")
	if newCookie.Value == refreshCookie.Value {
		t.Fatal("refresh did not rotate the refresh token")
	}
	if !newCookie.HttpOnly || newCookie.Path != "/api/v1/auth" || strings.Contains(rotated.Body.String(), newCookie.Value) {
		t.Fatal("rotated refresh token was not cookie-only with the expected cookie attributes")
	}
	rotatedAccessToken := accessTokenFromRecorder(t, rotated)
	if rotatedAccessToken == loginBody.Data.AccessToken {
		t.Fatal("refresh did not issue a new access token")
	}
	var oldTokenRows, rawTokenRows int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_sessions WHERE token_hash=?`, security.HashRefreshToken(refreshCookie.Value)).Scan(&oldTokenRows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_sessions WHERE token_hash=?`, refreshCookie.Value).Scan(&rawTokenRows); err != nil {
		t.Fatal(err)
	}
	if oldTokenRows != 1 || rawTokenRows != 0 {
		t.Fatalf("refresh token storage is not hash-only: old_hash_rows=%d raw_rows=%d", oldTokenRows, rawTokenRows)
	}
	reuse := postWithCookie(engine, "/api/v1/auth/refresh", refreshCookie, nil)
	if reuse.Code != http.StatusUnauthorized {
		t.Fatalf("old refresh token reuse status = %d body=%s", reuse.Code, reuse.Body.String())
	}
	if strings.Contains(reuse.Body.String(), "access_token") {
		t.Fatal("refresh token reuse returned a new access token")
	}
	familyRefresh := postWithCookie(engine, "/api/v1/auth/refresh", newCookie, nil)
	if familyRefresh.Code != http.StatusUnauthorized {
		t.Fatalf("reuse did not revoke the session family: %d %s", familyRefresh.Code, familyRefresh.Body.String())
	}
	var familyIDForRotated string
	if err := db.QueryRowContext(ctx, `SELECT family_id FROM refresh_sessions WHERE token_hash=?`, security.HashRefreshToken(newCookie.Value)).Scan(&familyIDForRotated); err != nil {
		t.Fatal(err)
	}
	var activeFamilySessions int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_sessions WHERE family_id=? AND revoked_at IS NULL`, familyIDForRotated).Scan(&activeFamilySessions); err != nil || activeFamilySessions != 0 {
		t.Fatalf("refresh replay left active family sessions: count=%d err=%v", activeFamilySessions, err)
	}
	expiredToken, err := security.NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	familyID, err := security.NewSessionFamilyID()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT id FROM users WHERE email='admin@example.org'`).Scan(&adminID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO refresh_sessions (user_id, token_hash, family_id, expires_at) VALUES (?, ?, ?, ?)`, adminID, security.HashRefreshToken(expiredToken), familyID, time.Now().UTC().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	expiredResponse := postWithCookie(engine, "/api/v1/auth/refresh", &http.Cookie{Name: "bem_refresh", Value: expiredToken}, nil)
	if expiredResponse.Code != http.StatusUnauthorized || strings.Contains(expiredResponse.Body.String(), "access_token") {
		t.Fatalf("expired refresh token status = %d", expiredResponse.Code)
	}

	secondLogin := postJSON(engine, "/api/v1/auth/login", `{"email":"admin@example.org","password":"a secure integration passphrase"}`, nil)
	if secondLogin.Code != http.StatusOK {
		t.Fatalf("second login status = %d body=%s", secondLogin.Code, secondLogin.Body.String())
	}
	secondCookie := responseCookie(t, secondLogin, "bem_refresh")
	logout := postWithCookie(engine, "/api/v1/auth/logout", secondCookie, map[string]string{"Authorization": "Bearer " + loginBody.Data.AccessToken})
	if logout.Code != http.StatusOK {
		t.Fatalf("logout status = %d body=%s", logout.Code, logout.Body.String())
	}
	clearCookie := responseCookie(t, logout, "bem_refresh")
	if clearCookie.MaxAge >= 0 || !clearCookie.HttpOnly {
		t.Fatalf("logout did not clear refresh cookie securely: %+v", clearCookie)
	}
	var logoutFamilyID string
	if err := db.QueryRowContext(ctx, `SELECT family_id FROM refresh_sessions WHERE token_hash=?`, security.HashRefreshToken(secondCookie.Value)).Scan(&logoutFamilyID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_sessions WHERE family_id=? AND revoked_at IS NULL`, logoutFamilyID).Scan(&activeFamilySessions); err != nil || activeFamilySessions != 0 {
		t.Fatalf("logout left active family sessions: count=%d err=%v", activeFamilySessions, err)
	}
	if response := postJSON(engine, "/api/v1/auth/logout", "{}", nil); response.Code != http.StatusOK {
		t.Fatalf("idempotent logout without cookie status = %d", response.Code)
	}

	if response := getWithToken(probe, "/super-admin", loginBody.Data.AccessToken); response.Code != http.StatusForbidden {
		t.Fatalf("updated BPH principal reached SUPER_ADMIN policy route: %d", response.Code)
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET status='INACTIVE' WHERE email='admin@example.org'`); err != nil {
		t.Fatal(err)
	}
	if inactiveMe := getWithToken(engine, "/api/v1/auth/me", loginBody.Data.AccessToken); inactiveMe.Code != http.StatusUnauthorized {
		t.Fatalf("existing JWT remained valid after deactivation: %d", inactiveMe.Code)
	}

	var loginEvents, logoutEvents int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_logs WHERE action='LOGIN'`).Scan(&loginEvents); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_logs WHERE action='LOGOUT'`).Scan(&logoutEvents); err != nil {
		t.Fatal(err)
	}
	if loginEvents != 6 || logoutEvents != 1 {
		t.Fatalf("auth activity counts LOGIN=%d LOGOUT=%d", loginEvents, logoutEvents)
	}
	var logoutActorID int64
	var logoutAt time.Time
	if err := db.QueryRowContext(ctx, `SELECT actor_user_id, created_at FROM activity_logs WHERE action='LOGOUT' ORDER BY id LIMIT 1`).Scan(&logoutActorID, &logoutAt); err != nil {
		t.Fatal(err)
	}
	if logoutActorID != adminID || logoutAt.Location() != time.UTC {
		t.Fatalf("LOGOUT activity actor or UTC timestamp is incorrect: actor=%d location=%s", logoutActorID, logoutAt.Location())
	}
	var sensitiveColumns int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=? AND LOWER(COLUMN_NAME) IN ('password','refresh_token','access_token','jwt_secret')`, dsn.DBName).Scan(&sensitiveColumns); err != nil || sensitiveColumns != 0 {
		t.Fatalf("database schema has raw credential/token columns: count=%d err=%v", sensitiveColumns, err)
	}
	rows, err := db.QueryContext(ctx, `SELECT COALESCE(details_json, '') FROM activity_logs`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var details string
		if err := rows.Scan(&details); err != nil {
			t.Fatal(err)
		}
		for _, sensitiveValue := range []string{"a secure integration passphrase", "incorrect credential", refreshCookie.Value, newCookie.Value, loginBody.Data.AccessToken, string(secret[:])} {
			if sensitiveValue != "" && strings.Contains(details, sensitiveValue) {
				t.Fatal("sensitive auth material appeared in activity details")
			}
		}
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	logs := loggerBuffer.String()
	for _, sensitiveValue := range []string{refreshCookie.Value, newCookie.Value, string(secret[:]), "a secure integration passphrase", "incorrect credential"} {
		if strings.Contains(logs, sensitiveValue) {
			t.Fatal("password, raw refresh token, or JWT secret appeared in structured logs")
		}
	}

	blockedOrigin := postJSONWithOrigin(engine, "/api/v1/auth/login", `{"email":"admin@example.org","password":"x"}`, "https://attacker.example")
	if blockedOrigin.Code != http.StatusForbidden || blockedOrigin.Header().Get("Access-Control-Allow-Origin") != "" || blockedOrigin.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatalf("untrusted Origin was not denied: %d", blockedOrigin.Code)
	}
	if got := login.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" || login.Header().Get("Access-Control-Allow-Credentials") != "true" || got == "*" {
		t.Fatalf("CORS response does not use the explicit credentialed allow-list: %+v", login.Header())
	}

	limitedConfig := cfg
	limitedConfig.LoginLimitCount = 1
	limitedEngine, err := NewWithDependencies(logger, limitedConfig, db)
	if err != nil {
		t.Fatalf("create rate-limited router: %v", err)
	}
	firstFailure := postJSON(limitedEngine, "/api/v1/auth/login", `{"email":"missing@example.org","password":"do not log this credential"}`, nil)
	secondFailure := postJSON(limitedEngine, "/api/v1/auth/login", `{"email":"missing@example.org","password":"do not log this credential"}`, nil)
	if firstFailure.Code != http.StatusUnauthorized || secondFailure.Code != http.StatusTooManyRequests {
		t.Fatalf("login IP rate limit did not enforce configured threshold: first=%d second=%d", firstFailure.Code, secondFailure.Code)
	}
	if strings.Contains(loggerBuffer.String(), "do not log this credential") {
		t.Fatal("rate-limited login password appeared in structured logs")
	}
}

func postJSON(engine http.Handler, path, body string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	return postJSONWithOriginAndCookies(engine, path, body, "http://localhost:3000", cookies, nil)
}

func accessTokenFromRecorder(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	var payload struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("parse access token response: %v (%s)", err, response.Body.String())
	}
	if payload.Data.AccessToken == "" {
		t.Fatalf("response omitted access token: %s", response.Body.String())
	}
	return payload.Data.AccessToken
}

func postJSONWithOrigin(engine http.Handler, path, body, origin string) *httptest.ResponseRecorder {
	return postJSONWithOriginAndCookies(engine, path, body, origin, nil, nil)
}

func postWithCookie(engine http.Handler, path string, cookie *http.Cookie, headers map[string]string) *httptest.ResponseRecorder {
	return postJSONWithOriginAndCookies(engine, path, "{}", "http://localhost:3000", []*http.Cookie{cookie}, headers)
}

func postJSONWithOriginAndCookies(engine http.Handler, path, body, origin string, cookies []*http.Cookie, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", origin)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}

func getWithToken(engine http.Handler, path, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Origin", "http://localhost:3000")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}

func responseCookie(t *testing.T, response *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("response did not set cookie %q: %v", name, response.Header().Values("Set-Cookie"))
	return nil
}

func mustAuthService(t *testing.T, db *sql.DB, cfg config.Config, logger *slog.Logger) *service.AuthService {
	t.Helper()
	auth, err := service.NewAuthService(db, cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	return auth
}
