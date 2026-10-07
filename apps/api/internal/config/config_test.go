package config

import "testing"

func TestLoadUsesSafeDevelopmentDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_PORT", "")
	t.Setenv("FRONTEND_URL", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("STORAGE_DIR", "")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.AppEnv != "development" || got.AppPort != 8080 || got.FrontendURL != "http://localhost:3000" {
		t.Fatalf("unexpected defaults: %+v", got)
	}
	if got.DatabaseURL != "" || got.JWTSecret != "" {
		t.Fatal("optional credentials should remain empty")
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("APP_PORT", "70000")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject an out-of-range port")
	}
}

func TestLoadRejectsNonHTTPFrontendURL(t *testing.T) {
	t.Setenv("FRONTEND_URL", "ftp://example.test")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a non-HTTP frontend URL")
	}
}
