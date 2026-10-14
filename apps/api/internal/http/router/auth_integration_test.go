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
	// This is an explicitly isolated *_test database. Remove prior integration media
	// fixtures before rolling back 00009, whose DOWN constraints target 00008 formats.
	var existingMediaTable int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=? AND table_name='media_assets'`, dsn.DBName).Scan(&existingMediaTable); err != nil {
		t.Fatalf("inspect dedicated test media table before reset: %v", err)
	}
	if existingMediaTable == 1 {
		if _, err := db.ExecContext(ctx, `DELETE FROM media_assets`); err != nil {
			t.Fatalf("clear dedicated test media fixtures before reset: %v", err)
		}
	}
	if err := migrate.DownToZero(ctx, db); err != nil {
		t.Fatalf("reset dedicated test database: %v", err)
	}
	if err := migrate.Up(ctx, db); err != nil {
		t.Fatalf("apply migrations from empty database: %v", err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, err := db.ExecContext(cleanupCtx, `DELETE FROM media_assets`); err != nil {
			t.Errorf("clean dedicated test media fixtures: %v", err)
			return
		}
		if err := migrate.DownToZero(cleanupCtx, db); err != nil {
			t.Errorf("clean dedicated test database: %v", err)
		}
	}()
	var tableCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name IN ('periods','divisions','users','refresh_sessions','activity_logs','article_categories','articles','article_category_map','gallery_categories','gallery_items','announcements','organization_members','aspirations','registration_forms','registration_fields','registration_field_options','registration_submissions','registration_submission_answers','registration_submission_files','documents','landing_settings','media_assets')`, dsn.DBName).Scan(&tableCount); err != nil || tableCount != 22 {
		t.Fatalf("phase 11 schema table count = %d, err = %v", tableCount, err)
	}
	var mediaIndexes, mediaForeignKeys, mediaChecks int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT index_name) FROM information_schema.statistics WHERE table_schema=? AND table_name='media_assets' AND index_name IN ('uq_media_assets_storage_key','idx_media_assets_uploader_purpose','idx_media_assets_sha256')`, dsn.DBName).Scan(&mediaIndexes); err != nil || mediaIndexes != 3 {
		t.Fatalf("media indexes = %d, err = %v", mediaIndexes, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.key_column_usage WHERE constraint_schema=? AND table_name='media_assets' AND constraint_name='fk_media_assets_uploader' AND referenced_table_name='users'`, dsn.DBName).Scan(&mediaForeignKeys); err != nil || mediaForeignKeys != 1 {
		t.Fatalf("media uploader foreign key count = %d, err = %v", mediaForeignKeys, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_schema=? AND table_name='media_assets' AND constraint_type='CHECK' AND constraint_name LIKE 'chk_media_assets_%'`, dsn.DBName).Scan(&mediaChecks); err != nil || mediaChecks != 5 {
		t.Fatalf("media check constraint count = %d, err = %v", mediaChecks, err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback public-structure visibility migration 00011: %v", err)
	}
	var visibilityColumns int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=? AND table_name='divisions' AND column_name='show_in_public_structure'`, dsn.DBName).Scan(&visibilityColumns); err != nil || visibilityColumns != 0 {
		t.Fatalf("public-structure visibility column remains after 00011 DOWN: %d err=%v", visibilityColumns, err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO periods(name,start_date,end_date,status) VALUES('Phase11 visibility backfill','2035-01-01','2035-12-31','ARCHIVED')`); err != nil {
		t.Fatalf("create migration visibility fixture period: %v", err)
	}
	var visibilityPeriodID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM periods WHERE name='Phase11 visibility backfill'`).Scan(&visibilityPeriodID); err != nil {
		t.Fatalf("read migration visibility fixture period: %v", err)
	}
	for _, divisionName := range []string{"Fakultas", "BPH", "Bidang Migrasi"} {
		if _, err := db.ExecContext(ctx, `INSERT INTO divisions(period_id,name,status) VALUES(?,?,'ACTIVE')`, visibilityPeriodID, divisionName); err != nil {
			t.Fatalf("create migration visibility fixture %q: %v", divisionName, err)
		}
	}
	if err := migrate.Up(ctx, db); err != nil {
		t.Fatalf("reapply public-structure visibility migration: %v", err)
	}
	var facultyVisible, bphVisible, workDivisionVisible bool
	if err := db.QueryRowContext(ctx, `SELECT MAX(CASE WHEN name='Fakultas' THEN show_in_public_structure END),MAX(CASE WHEN name='BPH' THEN show_in_public_structure END),MAX(CASE WHEN name='Bidang Migrasi' THEN show_in_public_structure END) FROM divisions WHERE period_id=?`, visibilityPeriodID).Scan(&facultyVisible, &bphVisible, &workDivisionVisible); err != nil || facultyVisible || bphVisible || !workDivisionVisible {
		t.Fatalf("migration backfill visibility: Fakultas=%v BPH=%v work=%v err=%v", facultyVisible, bphVisible, workDivisionVisible, err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback re-applied public-structure visibility migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM divisions WHERE period_id=?`, visibilityPeriodID); err != nil {
		t.Fatalf("remove migration visibility fixtures: %v", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM periods WHERE id=?`, visibilityPeriodID); err != nil {
		t.Fatalf("remove migration visibility fixture period: %v", err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback registration-image migration 00010: %v", err)
	}
	var privateFileTables, imageFieldChecks int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=? AND table_name='registration_submission_files'`, dsn.DBName).Scan(&privateFileTables); err != nil || privateFileTables != 0 {
		t.Fatalf("registration file table remains after 00010 DOWN: %d err=%v", privateFileTables, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_schema=? AND table_name='registration_fields' AND constraint_name='chk_registration_fields_type' AND constraint_type='CHECK'`, dsn.DBName).Scan(&imageFieldChecks); err != nil || imageFieldChecks != 1 {
		t.Fatalf("registration field check missing after 00010 DOWN: %d err=%v", imageFieldChecks, err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback multi-format media migration 00009: %v", err)
	}
	var mediaTables int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=? AND table_name='media_assets'`, dsn.DBName).Scan(&mediaTables); err != nil || mediaTables != 1 {
		t.Fatalf("media_assets table missing after migration 00009 DOWN: count=%d err=%v", mediaTables, err)
	}
	var originalMIMEColumns int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=? AND table_name='media_assets' AND column_name='original_mime_type'`, dsn.DBName).Scan(&originalMIMEColumns); err != nil || originalMIMEColumns != 0 {
		t.Fatalf("original_mime_type remains after migration 00009 DOWN: count=%d err=%v", originalMIMEColumns, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_schema=? AND table_name='media_assets' AND constraint_type='CHECK' AND constraint_name LIKE 'chk_media_assets_%'`, dsn.DBName).Scan(&mediaChecks); err != nil || mediaChecks != 4 {
		t.Fatalf("baseline media check constraint count after migration 00009 DOWN = %d, err = %v", mediaChecks, err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback public-media migration: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=? AND table_name='media_assets'`, dsn.DBName).Scan(&mediaTables); err != nil || mediaTables != 0 {
		t.Fatalf("media_assets table remains after migration 00008 DOWN: count=%d err=%v", mediaTables, err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback system-management migration: %v", err)
	}
	var settingsTables int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=? AND table_name='landing_settings'`, dsn.DBName).Scan(&settingsTables); err != nil || settingsTables != 0 {
		t.Fatalf("landing settings table remains after migration DOWN: count=%d err=%v", settingsTables, err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback documents migration: %v", err)
	}
	var documentTables int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=? AND table_name='documents'`, dsn.DBName).Scan(&documentTables); err != nil || documentTables != 0 {
		t.Fatalf("documents table remains after migration DOWN: count=%d err=%v", documentTables, err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback registration migration: %v", err)
	}
	var registrationTables int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=? AND table_name IN ('registration_forms','registration_fields','registration_field_options','registration_submissions','registration_submission_answers')`, dsn.DBName).Scan(&registrationTables); err != nil || registrationTables != 0 {
		t.Fatalf("registration tables remain after migration DOWN: count=%d err=%v", registrationTables, err)
	}
	if err := migrate.Down(ctx, db); err != nil {
		t.Fatalf("rollback aspiration migration: %v", err)
	}
	var aspirationsExists int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=? AND table_name='aspirations'`, dsn.DBName).Scan(&aspirationsExists); err != nil || aspirationsExists != 0 {
		t.Fatalf("aspirations table remains after migration DOWN: count=%d err=%v", aspirationsExists, err)
	}
	if err := migrate.Up(ctx, db); err != nil {
		t.Fatalf("reapply migrations after rollback: %v", err)
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
		RefreshLimitCount: 100, RefreshLimitWindow: time.Minute,
		AspirationLimitCount: 100, AspirationLimitWindow: 15 * time.Minute,
		RegistrationLimitCount: 100, RegistrationLimitWindow: 30 * time.Minute,
		DocumentStoragePath: filepath.Join(t.TempDir(), "private-document-root"), DocumentMaxUploadBytes: 10 << 20,
		MediaStoragePath: filepath.Join(t.TempDir(), "public-media-root"), MediaMaxUploadBytes: 5 << 20,
		RegistrationStoragePath: filepath.Join(t.TempDir(), "private-registration-root"), RegistrationImageMaxUploadBytes: 5 << 20,
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
	kepalaBAccessToken := accessTokenFromRecorder(t, kepalaBLogin)
	contentMedia := runMediaIntegration(t, ctx, db, engine, loginBody.Data.AccessToken, bphAccessToken, kepalaAccessToken, cfg.MediaStoragePath)
	runContentIntegration(t, ctx, db, engine, loginBody.Data.AccessToken, bphAccessToken, kepalaAccessToken, kepalaBAccessToken, divisionID, divisionBID, contentMedia)
	runOrganizationIntegration(t, ctx, db, engine, loginBody.Data.AccessToken, bphAccessToken, kepalaAccessToken)
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
	if _, err := db.ExecContext(ctx, `UPDATE users SET status='ACTIVE' WHERE email='admin@example.org'`); err != nil {
		t.Fatal(err)
	}
	// Earlier assertions intentionally changed the seeded administrator's
	// database role to verify that current DB role overrides JWT claims. Restore
	// the fixture before aspiration tests that require SUPER_ADMIN privileges.
	if _, err := db.ExecContext(ctx, `UPDATE users SET role='SUPER_ADMIN' WHERE email='admin@example.org'`); err != nil {
		t.Fatal(err)
	}
	runAspirationIntegration(t, ctx, db, engine, cfg, logger, loggerBuffer, loginBody.Data.AccessToken, bphAccessToken, kepalaAccessToken, kepalaBAccessToken)
	runRegistrationIntegration(t, ctx, db, engine, loggerBuffer, loginBody.Data.AccessToken, bphAccessToken, kepalaAccessToken, kepalaBAccessToken, divisionID, divisionBID)
	runDocumentIntegration(t, ctx, db, engine, loginBody.Data.AccessToken, bphAccessToken, kepalaAccessToken, kepalaBAccessToken, archivedPeriodID, divisionID, divisionBID, cfg.DocumentStoragePath, loggerBuffer)
	runSystemManagementIntegration(t, ctx, db, engine, loginBody.Data.AccessToken, bphAccessToken, kepalaAccessToken, divisionID, divisionBID)
	runDashboardIntegration(t, ctx, db, engine, loginBody.Data.AccessToken, bphAccessToken, kepalaAccessToken, kepalaBAccessToken, divisionID, divisionBID)
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

func runContentIntegration(t *testing.T, ctx context.Context, db *sql.DB, engine http.Handler, superToken, bphToken, kepalaAToken, kepalaBToken string, divisionA, divisionB int64, media contentMediaFixtures) {
	t.Helper()
	runArticleCoverIntegration(t, engine, superToken, media.ArticleCovers)

	categoryResponse := contentRequest(engine, http.MethodPost, "/api/v1/admin/article-categories", `{"name":"Kegiatan Kampus","slug":"","description":"Agenda organisasi"}`, superToken)
	if categoryResponse.Code != http.StatusCreated {
		t.Fatalf("create article category = %d %s", categoryResponse.Code, categoryResponse.Body.String())
	}
	var category struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(categoryResponse.Body.Bytes(), &category); err != nil || category.Data.ID == 0 {
		t.Fatalf("parse article category response: %v (%s)", err, categoryResponse.Body.String())
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/article-categories/%d", category.Data.ID), `{"name":"Kegiatan Resmi","slug":"kegiatan-resmi","description":"Kategori diperbarui"}`, superToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Kegiatan Resmi") {
		t.Fatalf("article category update failed: %d %s", response.Code, response.Body.String())
	}

	articleBody := fmt.Sprintf(`{"title":"Artikel Scope Akademik","slug":"artikel-scope-akademik","excerpt":"Ringkasan aman","content_html":"<p>Konten <strong>aman</strong><script>alert(1)</script><a href=\"javascript:alert(1)\" onclick=\"run()\">tautan</a></p>","division_id":%d,"category_ids":[%d]}`, divisionB, category.Data.ID)
	articleResponse := contentRequest(engine, http.MethodPost, "/api/v1/admin/articles", articleBody, kepalaAToken)
	if articleResponse.Code != http.StatusCreated {
		t.Fatalf("create scoped article = %d %s", articleResponse.Code, articleResponse.Body.String())
	}
	var article struct {
		Data struct {
			ID      int64  `json:"id"`
			Slug    string `json:"slug"`
			Content string `json:"content_html"`
			Status  string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(articleResponse.Body.Bytes(), &article); err != nil || article.Data.ID == 0 {
		t.Fatalf("parse article create response: %v (%s)", err, articleResponse.Body.String())
	}
	if article.Data.Status != "DRAFT" || strings.Contains(strings.ToLower(article.Data.Content), "script") || strings.Contains(strings.ToLower(article.Data.Content), "javascript:") || strings.Contains(strings.ToLower(article.Data.Content), "onclick") {
		t.Fatalf("article draft state or rich-text sanitization failed: %+v", article.Data)
	}
	var storedDivision sql.NullInt64
	if err := db.QueryRowContext(ctx, `SELECT division_id FROM articles WHERE id=?`, article.Data.ID).Scan(&storedDivision); err != nil || !storedDivision.Valid || storedDivision.Int64 != divisionA {
		t.Fatalf("KB create did not force authenticated division: value=%v err=%v", storedDivision, err)
	}

	if response := contentRequest(engine, http.MethodGet, fmt.Sprintf("/api/v1/admin/articles/%d", article.Data.ID), "", kepalaBToken); response.Code != http.StatusForbidden && response.Code != http.StatusNotFound {
		t.Fatalf("cross-division article detail status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, fmt.Sprintf("/api/v1/admin/articles/%d", article.Data.ID), "", bphToken); response.Code != http.StatusOK {
		t.Fatalf("BPH cross-division article detail status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/articles", "", ""); response.Code != http.StatusOK || strings.Contains(response.Body.String(), "Artikel Scope Akademik") {
		t.Fatalf("public article list leaked draft: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/articles/%d/publish", article.Data.ID), `{}`, kepalaAToken); response.Code != http.StatusOK {
		t.Fatalf("publish scoped article = %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/articles", "", ""); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Artikel Scope Akademik") || strings.Contains(response.Body.String(), "javascript:") || strings.Contains(response.Body.String(), "onclick") {
		t.Fatalf("public published article or sanitizer result invalid: %d %s", response.Code, response.Body.String())
	}
	updatedArticleBody := fmt.Sprintf(`{"title":"Artikel Scope Akademik Diperbarui","slug":"artikel-scope-akademik","excerpt":"Ringkasan diperbarui","content_html":"<p>Konten aman diperbarui</p>","division_id":%d,"category_ids":[%d]}`, divisionB, category.Data.ID)
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/articles/%d", article.Data.ID), updatedArticleBody, kepalaAToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Artikel Scope Akademik Diperbarui") {
		t.Fatalf("scoped article update failed or did not force the authenticated division: %d %s", response.Code, response.Body.String())
	}
	var articleDivision sql.NullInt64
	var updatedArticleStatus string
	if err := db.QueryRowContext(ctx, `SELECT division_id,status FROM articles WHERE id=?`, article.Data.ID).Scan(&articleDivision, &updatedArticleStatus); err != nil || !articleDivision.Valid || articleDivision.Int64 != divisionA || updatedArticleStatus != "PUBLISHED" {
		t.Fatalf("article update changed scope or publication state: division=%v status=%s err=%v", articleDivision, updatedArticleStatus, err)
	}
	duplicateArticle := strings.Replace(articleBody, `"division_id":`+fmt.Sprint(divisionB), `"division_id":`+fmt.Sprint(divisionA), 1)
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/articles", duplicateArticle, superToken); response.Code != http.StatusConflict {
		t.Fatalf("duplicate article slug status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/articles/%d/unpublish", article.Data.ID), `{}`, kepalaAToken); response.Code != http.StatusOK {
		t.Fatalf("unpublish scoped article = %d %s", response.Code, response.Body.String())
	}
	var articleStatus string
	var publishedAt sql.NullTime
	if err := db.QueryRowContext(ctx, `SELECT status,published_at FROM articles WHERE id=?`, article.Data.ID).Scan(&articleStatus, &publishedAt); err != nil || articleStatus != "DRAFT" || publishedAt.Valid {
		t.Fatalf("unpublish state did not clear published_at: status=%s date=%v err=%v", articleStatus, publishedAt, err)
	}
	deleteArticle := contentRequest(engine, http.MethodPost, "/api/v1/admin/articles", `{"title":"Artikel Hapus Uji","slug":"artikel-hapus-uji","content_html":"<p>Konten sementara</p>"}`, superToken)
	if deleteArticle.Code != http.StatusCreated {
		t.Fatalf("create disposable article = %d %s", deleteArticle.Code, deleteArticle.Body.String())
	}
	var deleteArticleID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM articles WHERE slug='artikel-hapus-uji'`).Scan(&deleteArticleID); err != nil {
		t.Fatal(err)
	}
	if response := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/articles/%d", deleteArticleID), "", superToken); response.Code != http.StatusNoContent {
		t.Fatalf("article delete failed: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, fmt.Sprintf("/api/v1/admin/articles/%d", deleteArticleID), "", superToken); response.Code != http.StatusNotFound {
		t.Fatalf("soft-deleted article remained readable: %d %s", response.Code, response.Body.String())
	}
	unusedArticleCategory := contentRequest(engine, http.MethodPost, "/api/v1/admin/article-categories", `{"name":"Kategori Hapus Uji","slug":"kategori-hapus-uji"}`, superToken)
	if unusedArticleCategory.Code != http.StatusCreated {
		t.Fatalf("create disposable article category = %d %s", unusedArticleCategory.Code, unusedArticleCategory.Body.String())
	}
	var unusedArticleCategoryID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM article_categories WHERE slug='kategori-hapus-uji'`).Scan(&unusedArticleCategoryID); err != nil {
		t.Fatal(err)
	}
	if response := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/article-categories/%d", unusedArticleCategoryID), "", superToken); response.Code != http.StatusNoContent {
		t.Fatalf("unused article category delete failed: %d %s", response.Code, response.Body.String())
	}

	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/gallery-categories", `{"title":"Kategori Terlarang","slug":"kategori-terlarang"}`, kepalaAToken); response.Code != http.StatusForbidden {
		t.Fatalf("KEPALA_BIDANG was allowed to change shared gallery taxonomy: %d %s", response.Code, response.Body.String())
	}
	galleryCategory := contentRequest(engine, http.MethodPost, "/api/v1/admin/gallery-categories", `{"title":"Agenda Aman","slug":"","description":"Dokumentasi kegiatan","division_id":`+fmt.Sprint(divisionA)+`}`, bphToken)
	if galleryCategory.Code != http.StatusCreated {
		t.Fatalf("create scoped gallery category=%d %s", galleryCategory.Code, galleryCategory.Body.String())
	}
	var gallery struct {
		Data struct {
			ID   int64  `json:"id"`
			Slug string `json:"slug"`
		} `json:"data"`
	}
	if err := json.Unmarshal(galleryCategory.Body.Bytes(), &gallery); err != nil || gallery.Data.ID == 0 {
		t.Fatalf("parse gallery category response: %v (%s)", err, galleryCategory.Body.String())
	}
	var galleryDivision sql.NullInt64
	if err := db.QueryRowContext(ctx, `SELECT division_id FROM gallery_categories WHERE id=?`, gallery.Data.ID).Scan(&galleryDivision); err != nil || !galleryDivision.Valid || galleryDivision.Int64 != divisionA {
		t.Fatalf("gallery category has unexpected division: division=%v err=%v", galleryDivision, err)
	}
	categoryReload := contentRequest(engine, http.MethodGet, "/api/v1/admin/gallery-categories?page=1&limit=100&search=Agenda%20Aman", "", bphToken)
	if categoryReload.Code != http.StatusOK || !strings.Contains(categoryReload.Body.String(), `"slug":"agenda-aman"`) {
		t.Fatalf("saved gallery category missing after list reload: %d %s", categoryReload.Code, categoryReload.Body.String())
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/gallery-categories/%d", gallery.Data.ID), `{"title":"Perubahan lintas divisi","slug":"agenda-aman"}`, kepalaBToken); response.Code != http.StatusForbidden {
		t.Fatalf("cross-division gallery category update status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/gallery-items", `{"category_id":0,"title":"Kategori Kosong","google_photos_url":"https://photos.app.goo.gl/nope"}`, bphToken); response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), `"category_id"`) {
		t.Fatalf("missing gallery category did not return field validation: %d %s", response.Code, response.Body.String())
	}
	galleryItemBody := fmt.Sprintf(`{"category_id":%d,"title":"Album Rapat","google_photos_url":"https://photos.app.goo.gl/aman","sort_order":1}`, gallery.Data.ID)
	galleryItem := contentRequest(engine, http.MethodPost, "/api/v1/admin/gallery-items", galleryItemBody, kepalaAToken)
	if galleryItem.Code != http.StatusCreated {
		t.Fatalf("create gallery item=%d %s", galleryItem.Code, galleryItem.Body.String())
	}
	var item struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(galleryItem.Body.Bytes(), &item); err != nil || item.Data.ID == 0 {
		t.Fatalf("parse gallery item response: %v (%s)", err, galleryItem.Body.String())
	}
	var itemWithoutPreview sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT preview_image_url FROM gallery_items WHERE id=?`, item.Data.ID).Scan(&itemWithoutPreview); err != nil || itemWithoutPreview.Valid {
		t.Fatalf("optional gallery preview should persist as NULL: value=%v err=%v", itemWithoutPreview, err)
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/gallery-items/%d", item.Data.ID), galleryItemBody, kepalaBToken); response.Code != http.StatusForbidden {
		t.Fatalf("cross-division gallery item update status=%d body=%s", response.Code, response.Body.String())
	}
	updatedGalleryItem := fmt.Sprintf(`{"category_id":%d,"title":"Album Rapat Diperbarui","google_photos_url":"https://photos.app.goo.gl/aman","sort_order":2}`, gallery.Data.ID)
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/gallery-items/%d", item.Data.ID), updatedGalleryItem, kepalaAToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Album Rapat Diperbarui") {
		t.Fatalf("same-division gallery item update failed: %d %s", response.Code, response.Body.String())
	}
	internalGalleryBody := fmt.Sprintf(`{"category_id":%d,"title":"Album Preview Internal","description":"Preview tersimpan","google_photos_url":"https://photos.app.goo.gl/preview","preview_image_url":%q,"sort_order":3}`, gallery.Data.ID, "http://localhost:8080"+media.GalleryPreview)
	internalGallery := contentRequest(engine, http.MethodPost, "/api/v1/admin/gallery-items", internalGalleryBody, kepalaAToken)
	if internalGallery.Code != http.StatusCreated {
		t.Fatalf("create gallery item with uploaded preview failed: %d %s", internalGallery.Code, internalGallery.Body.String())
	}
	var internalItem struct {
		Data struct {
			ID              int64   `json:"id"`
			PreviewImageURL *string `json:"preview_image_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(internalGallery.Body.Bytes(), &internalItem); err != nil || internalItem.Data.ID == 0 || internalItem.Data.PreviewImageURL == nil || *internalItem.Data.PreviewImageURL != media.GalleryPreview {
		t.Fatalf("uploaded gallery preview response was not normalized: %v (%s)", err, internalGallery.Body.String())
	}
	var persistedPreview sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT preview_image_url FROM gallery_items WHERE id=?`, internalItem.Data.ID).Scan(&persistedPreview); err != nil || !persistedPreview.Valid || persistedPreview.String != media.GalleryPreview {
		t.Fatalf("gallery media reference was not persisted canonically: value=%v err=%v", persistedPreview, err)
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/admin/gallery-items?category_id="+fmt.Sprint(gallery.Data.ID)+"&page=1&limit=100", "", bphToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Album Preview Internal") || !strings.Contains(response.Body.String(), media.GalleryPreview) {
		t.Fatalf("gallery item missing after admin list reload: %d %s", response.Code, response.Body.String())
	}
	publicGallery := contentRequest(engine, http.MethodGet, "/api/v1/public/gallery-categories/agenda-aman", "", "")
	if publicGallery.Code != http.StatusOK || !strings.Contains(publicGallery.Body.String(), "https://photos.app.goo.gl/preview") || !strings.Contains(publicGallery.Body.String(), media.GalleryPreview) {
		t.Fatalf("public gallery item did not expose distinct album and stored preview values: %d %s", publicGallery.Code, publicGallery.Body.String())
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/gallery-items/%d", internalItem.Data.ID), fmt.Sprintf(`{"category_id":%d,"title":"Album Preview Internal","google_photos_url":"https://photos.app.goo.gl/preview","preview_image_url":"https://cdn.example.test/legacy-preview.jpg","sort_order":3}`, gallery.Data.ID), kepalaAToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "cdn.example.test/legacy-preview.jpg") {
		t.Fatalf("legacy HTTPS gallery preview update rejected: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/gallery-items/%d", internalItem.Data.ID), fmt.Sprintf(`{"category_id":%d,"title":"Invalid Preview","google_photos_url":"https://photos.app.goo.gl/preview","preview_image_url":"/uploads/local.png","sort_order":3}`, gallery.Data.ID), kepalaAToken); response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), `"preview_image_url"`) {
		t.Fatalf("arbitrary gallery preview path was not rejected with field feedback: %d %s", response.Code, response.Body.String())
	}
	publicGallery = contentRequest(engine, http.MethodGet, "/api/v1/public/gallery-categories/agenda-aman", "", "")
	if publicGallery.Code != http.StatusOK || !strings.Contains(publicGallery.Body.String(), "https://photos.app.goo.gl/aman") {
		t.Fatalf("public gallery category/item response=%d %s", publicGallery.Code, publicGallery.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/gallery-items", `{"category_id":`+fmt.Sprint(gallery.Data.ID)+`,"title":"HTTP tidak aman","google_photos_url":"http://photos.app.goo.gl/nope"}`, kepalaAToken); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("non-HTTPS album URL accepted: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/gallery-items/%d", item.Data.ID), "", kepalaAToken); response.Code != http.StatusNoContent {
		t.Fatalf("same-division gallery item delete failed: %d %s", response.Code, response.Body.String())
	}
	unusedGalleryCategory := contentRequest(engine, http.MethodPost, "/api/v1/admin/gallery-categories", `{"title":"Album Hapus Uji","slug":"album-hapus-uji"}`, bphToken)
	if unusedGalleryCategory.Code != http.StatusCreated {
		t.Fatalf("create disposable gallery category = %d %s", unusedGalleryCategory.Code, unusedGalleryCategory.Body.String())
	}
	var unusedGalleryCategoryID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM gallery_categories WHERE slug='album-hapus-uji'`).Scan(&unusedGalleryCategoryID); err != nil {
		t.Fatal(err)
	}
	if response := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/gallery-categories/%d", unusedGalleryCategoryID), "", bphToken); response.Code != http.StatusNoContent {
		t.Fatalf("empty gallery category delete failed: %d %s", response.Code, response.Body.String())
	}

	start := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	end := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	announcementBody := fmt.Sprintf(`{"title":"Pengumuman Scope Test","slug":"pengumuman-scope-test","content_html":"<p>Info resmi</p>","start_at":%q,"end_at":%q}`, start, end)
	announcementResponse := contentRequest(engine, http.MethodPost, "/api/v1/admin/announcements", announcementBody, bphToken)
	if announcementResponse.Code != http.StatusCreated {
		t.Fatalf("create announcement=%d %s", announcementResponse.Code, announcementResponse.Body.String())
	}
	var announcement struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(announcementResponse.Body.Bytes(), &announcement); err != nil || announcement.Data.ID == 0 || announcement.Data.Status != "INACTIVE" {
		t.Fatalf("announcement initial state invalid: %v (%s)", err, announcementResponse.Body.String())
	}
	var noAnnouncementImage sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT image_url FROM announcements WHERE id=?`, announcement.Data.ID).Scan(&noAnnouncementImage); err != nil || noAnnouncementImage.Valid {
		t.Fatalf("optional announcement image should persist as NULL: value=%v err=%v", noAnnouncementImage, err)
	}
	startOnlyBody := fmt.Sprintf(`{"title":"Pengumuman Tanpa Akhir","slug":"pengumuman-tanpa-akhir","content_html":"<p>Tanpa tanggal akhir</p>","start_at":%q}`, start)
	startOnly := contentRequest(engine, http.MethodPost, "/api/v1/admin/announcements", startOnlyBody, bphToken)
	if startOnly.Code != http.StatusCreated {
		t.Fatalf("announcement with optional end_at omitted failed: %d %s", startOnly.Code, startOnly.Body.String())
	}
	var startOnlyData struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(startOnly.Body.Bytes(), &startOnlyData); err != nil || startOnlyData.Data.ID == 0 {
		t.Fatalf("parse start-only announcement response: %v %s", err, startOnly.Body.String())
	}
	var persistedEnd sql.NullTime
	if err := db.QueryRowContext(ctx, `SELECT end_at FROM announcements WHERE id=?`, startOnlyData.Data.ID).Scan(&persistedEnd); err != nil || persistedEnd.Valid {
		t.Fatalf("optional announcement end_at should remain NULL: value=%v err=%v", persistedEnd, err)
	}
	mediaAnnouncementBody := fmt.Sprintf(`{"title":"Pengumuman Gambar Uji","slug":"pengumuman-gambar-uji","content_html":"<p>Info dengan gambar</p>","image_url":%q,"start_at":%q,"end_at":%q}`, "http://localhost:8080"+media.AnnouncementImage, start, end)
	mediaAnnouncement := contentRequest(engine, http.MethodPost, "/api/v1/admin/announcements", mediaAnnouncementBody, bphToken)
	if mediaAnnouncement.Code != http.StatusCreated {
		t.Fatalf("create announcement with uploaded image failed: %d %s", mediaAnnouncement.Code, mediaAnnouncement.Body.String())
	}
	var mediaAnnouncementData struct {
		Data struct {
			ID       int64   `json:"id"`
			ImageURL *string `json:"image_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(mediaAnnouncement.Body.Bytes(), &mediaAnnouncementData); err != nil || mediaAnnouncementData.Data.ID == 0 || mediaAnnouncementData.Data.ImageURL == nil || *mediaAnnouncementData.Data.ImageURL != media.AnnouncementImage {
		t.Fatalf("announcement media response was not normalized: %v (%s)", err, mediaAnnouncement.Body.String())
	}
	var persistedAnnouncementImage sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT image_url FROM announcements WHERE id=?`, mediaAnnouncementData.Data.ID).Scan(&persistedAnnouncementImage); err != nil || !persistedAnnouncementImage.Valid || persistedAnnouncementImage.String != media.AnnouncementImage {
		t.Fatalf("announcement media reference was not persisted canonically: value=%v err=%v", persistedAnnouncementImage, err)
	}
	if response := contentRequest(engine, http.MethodGet, fmt.Sprintf("/api/v1/admin/announcements/%d", mediaAnnouncementData.Data.ID), "", bphToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), media.AnnouncementImage) {
		t.Fatalf("saved announcement image missing after detail reload: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/announcements/%d/activate", mediaAnnouncementData.Data.ID), `{}`, bphToken); response.Code != http.StatusOK {
		t.Fatalf("media announcement activation failed: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/announcements/active", "", ""); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "pengumuman-gambar-uji") || !strings.Contains(response.Body.String(), media.AnnouncementImage) {
		t.Fatalf("active public announcement did not render the saved image reference: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/announcements/%d/deactivate", mediaAnnouncementData.Data.ID), `{}`, bphToken); response.Code != http.StatusOK {
		t.Fatalf("media announcement deactivation failed: %d %s", response.Code, response.Body.String())
	}
	legacyAnnouncementBody := fmt.Sprintf(`{"title":"Pengumuman Gambar Uji","slug":"pengumuman-gambar-uji","content_html":"<p>Info dengan gambar</p>","image_url":"https://cdn.example.test/legacy.jpg","start_at":%q,"end_at":%q}`, start, end)
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/announcements/%d", mediaAnnouncementData.Data.ID), legacyAnnouncementBody, bphToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "cdn.example.test/legacy.jpg") {
		t.Fatalf("legacy HTTPS announcement image update rejected: %d %s", response.Code, response.Body.String())
	}
	invalidAnnouncementImage := strings.Replace(mediaAnnouncementBody, "http://localhost:8080"+media.AnnouncementImage, "data:text/html,<h1>x</h1>", 1)
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/announcements", invalidAnnouncementImage, bphToken); response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), `"image_url"`) {
		t.Fatalf("unsafe announcement image was not rejected with field feedback: %d %s", response.Code, response.Body.String())
	}
	invalidSchedule := fmt.Sprintf(`{"title":"Tanggal Salah","slug":"tanggal-salah","content_html":"<p>Isi</p>","start_at":%q,"end_at":%q}`, end, start)
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/announcements", invalidSchedule, bphToken); response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), `"end_at"`) {
		t.Fatalf("invalid announcement date order did not identify end_at: %d %s", response.Code, response.Body.String())
	}
	invalidDate := `{"title":"Tanggal Format Salah","slug":"tanggal-format-salah","content_html":"<p>Isi</p>","start_at":"not-a-date"}`
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/announcements", invalidDate, bphToken); response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), `"start_at"`) {
		t.Fatalf("invalid announcement datetime did not identify start_at: %d %s", response.Code, response.Body.String())
	}
	updatedAnnouncementBody := fmt.Sprintf(`{"title":"Pengumuman Scope Diperbarui","slug":"pengumuman-scope-test","content_html":"<p>Info resmi diperbarui</p>","start_at":%q,"end_at":%q}`, start, end)
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/announcements/%d", announcement.Data.ID), updatedAnnouncementBody, bphToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Pengumuman Scope Diperbarui") {
		t.Fatalf("announcement update failed: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/announcements/active", "", ""); response.Code != http.StatusOK || strings.Contains(response.Body.String(), "Pengumuman Scope Test") {
		t.Fatalf("inactive announcement appeared publicly: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/announcements/%d/activate", announcement.Data.ID), `{}`, bphToken); response.Code != http.StatusOK {
		t.Fatalf("activate announcement=%d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/announcements/active", "", ""); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Pengumuman Scope Diperbarui") {
		t.Fatalf("active announcement absent from public active list: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/announcements", announcementBody, kepalaAToken); response.Code != http.StatusForbidden {
		t.Fatalf("KEPALA_BIDANG was allowed to manage global announcement: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/announcements/%d/deactivate", announcement.Data.ID), `{}`, bphToken); response.Code != http.StatusOK {
		t.Fatalf("deactivate announcement=%d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/announcements/pengumuman-scope-test", "", ""); response.Code != http.StatusNotFound {
		t.Fatalf("inactive announcement detail exposed: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/announcements/%d", announcement.Data.ID), "", bphToken); response.Code != http.StatusNoContent {
		t.Fatalf("announcement delete failed: %d %s", response.Code, response.Body.String())
	}

	var activityCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_logs WHERE entity_type IN ('ARTICLE_CATEGORY','ARTICLE','GALLERY_CATEGORY','GALLERY_ITEM','ANNOUNCEMENT') AND action IN ('CREATE','UPDATE','DELETE','PUBLISH','UNPUBLISH','ACTIVATE','DEACTIVATE')`).Scan(&activityCount); err != nil || activityCount < 9 {
		t.Fatalf("content activity events missing: count=%d err=%v", activityCount, err)
	}
}

func runArticleCoverIntegration(t *testing.T, engine http.Handler, token string, covers []articleCoverFixture) {
	t.Helper()
	for index, cover := range covers {
		t.Run("article_cover_"+strings.TrimPrefix(filepath.Ext(cover.Filename), "."), func(t *testing.T) {
			title := fmt.Sprintf("Cover Format %d %s", index+1, strings.TrimPrefix(filepath.Ext(cover.Filename), "."))
			body, err := json.Marshal(map[string]any{
				"title": title, "slug": fmt.Sprintf("cover-format-%d-%s", index+1, strings.TrimPrefix(filepath.Ext(cover.Filename), ".")),
				"content_html": "<p>Pengujian cover multi-format.</p>", "thumbnail_url": "http://localhost:8080" + cover.PublicURL,
			})
			if err != nil {
				t.Fatal(err)
			}
			created := contentRequest(engine, http.MethodPost, "/api/v1/admin/articles", string(body), token)
			if created.Code != http.StatusCreated {
				t.Fatalf("create article with %s cover: %d %s", cover.Filename, created.Code, created.Body.String())
			}
			var article struct {
				Data struct {
					ID           int64   `json:"id"`
					Slug         string  `json:"slug"`
					ThumbnailURL *string `json:"thumbnail_url"`
				} `json:"data"`
			}
			if err := json.Unmarshal(created.Body.Bytes(), &article); err != nil || article.Data.ID < 1 || article.Data.ThumbnailURL == nil || *article.Data.ThumbnailURL != cover.PublicURL {
				t.Fatalf("created article did not normalize cover reference: err=%v body=%s", err, created.Body.String())
			}
			adminURL := fmt.Sprintf("/api/v1/admin/articles/%d", article.Data.ID)
			if reloaded := contentRequest(engine, http.MethodGet, adminURL, "", token); reloaded.Code != http.StatusOK || !strings.Contains(reloaded.Body.String(), cover.PublicURL) {
				t.Fatalf("reload %s cover: %d %s", cover.Filename, reloaded.Code, reloaded.Body.String())
			}
			body, err = json.Marshal(map[string]any{
				"title": title + " Diperbarui", "slug": article.Data.Slug,
				"content_html": "<p>Cover diperbarui.</p>", "thumbnail_url": cover.PublicURL,
			})
			if err != nil {
				t.Fatal(err)
			}
			updated := contentRequest(engine, http.MethodPatch, adminURL, string(body), token)
			if updated.Code != http.StatusOK || !strings.Contains(updated.Body.String(), cover.PublicURL) {
				t.Fatalf("edit %s cover: %d %s", cover.Filename, updated.Code, updated.Body.String())
			}
			published := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/articles/%d/publish", article.Data.ID), `{}`, token)
			if published.Code != http.StatusOK {
				t.Fatalf("publish %s cover article: %d %s", cover.Filename, published.Code, published.Body.String())
			}
			public := contentRequest(engine, http.MethodGet, "/api/v1/public/articles/"+article.Data.Slug, "", "")
			if public.Code != http.StatusOK || !strings.Contains(public.Body.String(), cover.PublicURL) {
				t.Fatalf("public %s cover article: %d %s", cover.Filename, public.Code, public.Body.String())
			}
			mediaResponse := httptest.NewRecorder()
			engine.ServeHTTP(mediaResponse, httptest.NewRequest(http.MethodGet, cover.PublicURL, nil))
			if mediaResponse.Code != http.StatusOK || mediaResponse.Header().Get("Content-Type") != cover.StoredMIME || len(mediaResponse.Body.Bytes()) == 0 {
				t.Fatalf("public %s media response status=%d mime=%s size=%d", cover.Filename, mediaResponse.Code, mediaResponse.Header().Get("Content-Type"), mediaResponse.Body.Len())
			}
		})
	}
}

func contentRequest(engine http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}
