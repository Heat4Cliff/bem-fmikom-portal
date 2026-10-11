package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func testEnv(overrides map[string]string) func(string) string {
	values := map[string]string{
		"APP_ENV": "development", "DATABASE_URL": "test:test@tcp(127.0.0.1:3306)/bem_fmikom_test",
		"JWT_SECRET": strings.Repeat("x", 32), "FRONTEND_ORIGINS": "http://localhost:3000",
	}
	for key, value := range overrides {
		values[key] = value
	}
	return func(key string) string { return values[key] }
}

func TestLoadUsesSecureDevelopmentDefaults(t *testing.T) {
	got, err := LoadFrom(testEnv(nil))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if got.AppEnv != "development" || got.AppPort != 8080 || got.AccessTokenTTL.String() != "15m0s" || got.RefreshTokenTTL.String() != "168h0m0s" {
		t.Fatalf("unexpected defaults: %+v", got)
	}
	if got.Cookie.Secure || got.Cookie.SameSite != "lax" || got.Cookie.Path != "/api/v1/auth" {
		t.Fatalf("unexpected development cookie defaults: %+v", got.Cookie)
	}
	if got.BcryptCost != 12 || len(got.FrontendOrigins) != 1 || got.LoginLimitCount != 5 || got.RefreshLimitCount != 30 || got.RefreshLimitWindow.String() != "1m0s" || got.AspirationLimitCount != 5 || got.AspirationLimitWindow.String() != "15m0s" || got.RegistrationLimitCount != 5 || got.RegistrationLimitWindow.String() != "30m0s" {
		t.Fatalf("unexpected auth defaults: %+v", got)
	}
	if got.DocumentStoragePath != "..\\..\\var\\private\\documents" || got.DocumentMaxUploadBytes != 10<<20 {
		t.Fatalf("unexpected document defaults: %+v", got)
	}
}

func TestLoadRejectsInvalidSecuritySettings(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
	}{
		{"weak secret", map[string]string{"JWT_SECRET": "short"}},
		{"wildcard origin", map[string]string{"FRONTEND_ORIGINS": "*"}},
		{"origin path", map[string]string{"FRONTEND_ORIGINS": "https://portal.example/path"}},
		{"none without secure", map[string]string{"AUTH_REFRESH_COOKIE_SAMESITE": "none"}},
		{"invalid same site", map[string]string{"AUTH_REFRESH_COOKIE_SAMESITE": "invalid"}},
		{"invalid duration", map[string]string{"AUTH_ACCESS_TOKEN_TTL": "soon"}},
		{"invalid document limit", map[string]string{"DOCUMENT_MAX_UPLOAD_BYTES": "0"}},
		{"production relative document path", map[string]string{"APP_ENV": "production", "JWT_SECRET": "0123456789abcdef0123456789abcdef", "DOCUMENT_STORAGE_PATH": "var/private/documents"}},
		{"production insecure cookie override", map[string]string{"APP_ENV": "production", "JWT_SECRET": "0123456789abcdef0123456789abcdef", "FRONTEND_ORIGINS": "https://portal.example", "AUTH_REFRESH_COOKIE_SECURE": "false"}},
		{"production local origin", map[string]string{"APP_ENV": "production", "JWT_SECRET": "0123456789abcdef0123456789abcdef", "FRONTEND_ORIGINS": "https://localhost:3000"}},
		{"production non-https origin", map[string]string{"APP_ENV": "production", "JWT_SECRET": "0123456789abcdef0123456789abcdef", "FRONTEND_ORIGINS": "http://portal.example"}},
		{"production default origin rejected", map[string]string{"APP_ENV": "production", "JWT_SECRET": "0123456789abcdef0123456789abcdef"}},
		{"production root database rejected", map[string]string{"APP_ENV": "production", "JWT_SECRET": "0123456789abcdef0123456789abcdef", "DATABASE_URL": "root:test@tcp(127.0.0.1:3306)/bem_fmikom_test", "FRONTEND_ORIGINS": "https://portal.example"}},
		{"unknown environment rejected", map[string]string{"APP_ENV": "prodution"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			overrides := test.overrides
			if overrides["APP_ENV"] == "production" && test.name != "production relative document path" {
				if overrides == nil {
					overrides = make(map[string]string)
				}
				if overrides["DOCUMENT_STORAGE_PATH"] == "" {
					overrides["DOCUMENT_STORAGE_PATH"] = filepath.Join(t.TempDir(), "private-documents")
				}
			}
			if _, err := LoadFrom(testEnv(overrides)); err == nil {
				t.Fatal("LoadFrom() expected validation error")
			}
		})
	}
}

func TestLoadProductionSecureCookieDefaultAndMultipleOrigins(t *testing.T) {
	got, err := LoadFrom(testEnv(map[string]string{"APP_ENV": "production", "JWT_SECRET": "0123456789abcdef0123456789abcdef", "FRONTEND_ORIGINS": "https://portal.example, https://admin.example", "DOCUMENT_STORAGE_PATH": filepath.Join(t.TempDir(), "private-documents")}))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if !got.Cookie.Secure || len(got.FrontendOrigins) != 2 {
		t.Fatalf("unexpected production settings: %+v", got)
	}
}

func TestLoadProductionRequiresExplicitHTTPSOriginsAndNonRootDatabase(t *testing.T) {
	_, err := LoadFrom(testEnv(map[string]string{"APP_ENV": "production", "JWT_SECRET": "0123456789abcdef0123456789abcdef", "FRONTEND_ORIGINS": "https://portal.example", "DATABASE_URL": "bem_app:password@tcp(db.example:3306)/bem_fmikom", "DOCUMENT_STORAGE_PATH": filepath.Join(t.TempDir(), "private-documents")}))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
}

func TestProductionRejectsWeakSecretPlaceholder(t *testing.T) {
	if _, err := LoadFrom(testEnv(map[string]string{"APP_ENV": "production", "JWT_SECRET": "LOCAL_ONLY_GENERATE_A_RANDOM_32_BYTE_SECRET"})); err == nil {
		t.Fatal("production should reject the checked-in local secret marker")
	}
	if _, err := LoadFrom(testEnv(map[string]string{"APP_ENV": "production", "JWT_SECRET": strings.Repeat("x", 32)})); err == nil {
		t.Fatal("production should reject a low-diversity secret")
	}
}
