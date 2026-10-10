package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Cookie struct {
	Name     string
	Secure   bool
	SameSite string
	Domain   string
	Path     string
}

type Config struct {
	AppEnv                  string
	AppPort                 int
	FrontendOrigins         []string
	DatabaseURL             string
	DBMaxOpenConns          int
	DBMaxIdleConns          int
	DBConnMaxLife           time.Duration
	JWTSecret               []byte
	AccessTokenTTL          time.Duration
	RefreshTokenTTL         time.Duration
	Cookie                  Cookie
	BcryptCost              int
	LoginLimitCount         int
	LoginLimitWindow        time.Duration
	AspirationLimitCount    int
	AspirationLimitWindow   time.Duration
	RegistrationLimitCount  int
	RegistrationLimitWindow time.Duration
	DocumentStoragePath     string
	DocumentMaxUploadBytes  int64
}

type DatabaseSettings struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Load() (Config, error) { return LoadFrom(os.Getenv) }

func LoadDatabaseSettings() (DatabaseSettings, error) { return LoadDatabaseSettingsFrom(os.Getenv) }

func LoadDatabaseSettingsFrom(lookup func(string) string) (DatabaseSettings, error) {
	url := strings.TrimSpace(lookup("DATABASE_URL"))
	if url == "" {
		return DatabaseSettings{}, fmt.Errorf("DATABASE_URL is required")
	}
	if _, err := mysql.ParseDSN(url); err != nil {
		return DatabaseSettings{}, fmt.Errorf("DATABASE_URL must be a valid MySQL driver DSN")
	}
	open, err := parseIntEnv(lookup, "DB_MAX_OPEN_CONNS", 25, 1, 500)
	if err != nil {
		return DatabaseSettings{}, err
	}
	idle, err := parseIntEnv(lookup, "DB_MAX_IDLE_CONNS", 5, 0, open)
	if err != nil {
		return DatabaseSettings{}, err
	}
	lifetime, err := parseDurationEnv(lookup, "DB_CONN_MAX_LIFETIME", "5m", time.Minute)
	if err != nil {
		return DatabaseSettings{}, err
	}
	return DatabaseSettings{URL: url, MaxOpenConns: open, MaxIdleConns: idle, ConnMaxLifetime: lifetime}, nil
}

func LoadFrom(lookup func(string) string) (Config, error) {
	get := func(key, fallback string) string {
		if value := strings.TrimSpace(lookup(key)); value != "" {
			return value
		}
		return fallback
	}
	parseInt := func(key string, fallback, min, max int) (int, error) {
		raw := get(key, strconv.Itoa(fallback))
		value, err := strconv.Atoi(raw)
		if err != nil || value < min || value > max {
			return 0, fmt.Errorf("%s must be an integer between %d and %d", key, min, max)
		}
		return value, nil
	}
	parseDuration := func(key, fallback string, min time.Duration) (time.Duration, error) {
		raw := get(key, fallback)
		value, err := time.ParseDuration(raw)
		if err != nil || value < min {
			return 0, fmt.Errorf("%s must be a duration of at least %s", key, min)
		}
		return value, nil
	}

	env := get("APP_ENV", "development")
	port, err := parseInt("APP_PORT", 8080, 1, 65535)
	if err != nil {
		return Config{}, err
	}
	originsRaw := get("FRONTEND_ORIGINS", get("FRONTEND_URL", "http://localhost:3000"))
	origins, err := parseOrigins(originsRaw)
	if err != nil {
		return Config{}, err
	}
	dbSettings, err := LoadDatabaseSettingsFrom(lookup)
	if err != nil {
		return Config{}, err
	}
	databaseURL := dbSettings.URL
	secret := []byte(lookup("JWT_SECRET"))
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	if env == "production" && (containsInsecureMarker(string(secret)) || !hasSufficientSecretDiversity(secret)) {
		return Config{}, fmt.Errorf("JWT_SECRET must be a unique secret in production")
	}

	accessTTL, err := parseDuration("AUTH_ACCESS_TOKEN_TTL", "15m", time.Minute)
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := parseDuration("AUTH_REFRESH_TOKEN_TTL", "168h", time.Hour)
	if err != nil {
		return Config{}, err
	}
	bcryptCost, err := parseInt("AUTH_BCRYPT_COST", 12, 4, 31)
	if err != nil {
		return Config{}, err
	}
	maxOpen, maxIdle, connectionLifetime := dbSettings.MaxOpenConns, dbSettings.MaxIdleConns, dbSettings.ConnMaxLifetime
	loginCount, err := parseInt("AUTH_LOGIN_RATE_LIMIT", 5, 1, 10000)
	if err != nil {
		return Config{}, err
	}
	loginWindow, err := parseDuration("AUTH_LOGIN_RATE_WINDOW", "1m", time.Second)
	if err != nil {
		return Config{}, err
	}
	aspirationCount, err := parseInt("PUBLIC_ASPIRATION_RATE_LIMIT", 5, 1, 10000)
	if err != nil {
		return Config{}, err
	}
	aspirationWindow, err := parseDuration("PUBLIC_ASPIRATION_RATE_WINDOW", "15m", time.Second)
	if err != nil {
		return Config{}, err
	}
	registrationCount, err := parseInt("PUBLIC_REGISTRATION_RATE_LIMIT", 5, 1, 10000)
	if err != nil {
		return Config{}, err
	}
	registrationWindow, err := parseDuration("PUBLIC_REGISTRATION_RATE_WINDOW", "30m", time.Second)
	if err != nil {
		return Config{}, err
	}
	documentMaxUploadBytes, err := parseInt("DOCUMENT_MAX_UPLOAD_BYTES", 10<<20, 1, 1<<30)
	if err != nil {
		return Config{}, err
	}
	documentStoragePath := filepath.Clean(get("DOCUMENT_STORAGE_PATH", "../../var/private/documents"))
	if env == "production" && !filepath.IsAbs(documentStoragePath) {
		return Config{}, fmt.Errorf("DOCUMENT_STORAGE_PATH must be an absolute private path in production")
	}

	cookieSecureDefault := "false"
	if env == "production" {
		cookieSecureDefault = "true"
	}
	secure, err := strconv.ParseBool(get("AUTH_REFRESH_COOKIE_SECURE", cookieSecureDefault))
	if err != nil {
		return Config{}, fmt.Errorf("AUTH_REFRESH_COOKIE_SECURE must be true or false")
	}
	sameSite := strings.ToLower(get("AUTH_REFRESH_COOKIE_SAMESITE", "lax"))
	if sameSite != "lax" && sameSite != "strict" && sameSite != "none" {
		return Config{}, fmt.Errorf("AUTH_REFRESH_COOKIE_SAMESITE must be lax, strict, or none")
	}
	if sameSite == "none" && !secure {
		return Config{}, fmt.Errorf("AUTH_REFRESH_COOKIE_SECURE must be true when SameSite=None")
	}
	cookieName := get("AUTH_REFRESH_COOKIE_NAME", "bem_refresh")
	if strings.ContainsAny(cookieName, " ;,\r\n") {
		return Config{}, fmt.Errorf("AUTH_REFRESH_COOKIE_NAME is invalid")
	}
	cookieDomain := strings.TrimSpace(lookup("AUTH_REFRESH_COOKIE_DOMAIN"))
	if cookieDomain != "" && strings.ContainsAny(cookieDomain, "/ :\r\n") {
		return Config{}, fmt.Errorf("AUTH_REFRESH_COOKIE_DOMAIN is invalid")
	}

	return Config{
		AppEnv: env, AppPort: port, FrontendOrigins: origins, DatabaseURL: databaseURL,
		DBMaxOpenConns: maxOpen, DBMaxIdleConns: maxIdle, DBConnMaxLife: connectionLifetime,
		JWTSecret: secret, AccessTokenTTL: accessTTL, RefreshTokenTTL: refreshTTL,
		Cookie:     Cookie{Name: cookieName, Secure: secure, SameSite: sameSite, Domain: cookieDomain, Path: "/api/v1/auth"},
		BcryptCost: bcryptCost, LoginLimitCount: loginCount, LoginLimitWindow: loginWindow,
		AspirationLimitCount: aspirationCount, AspirationLimitWindow: aspirationWindow,
		RegistrationLimitCount: registrationCount, RegistrationLimitWindow: registrationWindow,
		DocumentStoragePath: documentStoragePath, DocumentMaxUploadBytes: int64(documentMaxUploadBytes),
	}, nil
}

func containsInsecureMarker(secret string) bool {
	lower := strings.ToLower(secret)
	for _, marker := range []string{"change_me", "local_only", "replace_with", "development_secret", "example_secret"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func hasSufficientSecretDiversity(secret []byte) bool {
	unique := make(map[byte]struct{}, len(secret))
	for _, value := range secret {
		unique[value] = struct{}{}
	}
	return len(unique) >= 16
}

func parseIntEnv(lookup func(string) string, key string, fallback, min, max int) (int, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		raw = strconv.Itoa(fallback)
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", key, min, max)
	}
	return value, nil
}

func parseDurationEnv(lookup func(string) string, key, fallback string, min time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(lookup(key))
	if raw == "" {
		raw = fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value < min {
		return 0, fmt.Errorf("%s must be a duration of at least %s", key, min)
	}
	return value, nil
}

func parseOrigins(raw string) ([]string, error) {
	seen := make(map[string]struct{})
	var origins []string
	for _, entry := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(entry)
		if origin == "" {
			continue
		}
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
			return nil, fmt.Errorf("FRONTEND_ORIGINS must contain explicit HTTP(S) origins without paths")
		}
		canonical := strings.ToLower(parsed.Scheme + "://" + parsed.Host)
		if strings.Contains(canonical, "*") {
			return nil, fmt.Errorf("FRONTEND_ORIGINS must not contain wildcard origins")
		}
		if _, exists := seen[canonical]; !exists {
			seen[canonical] = struct{}{}
			origins = append(origins, canonical)
		}
	}
	if len(origins) == 0 {
		return nil, fmt.Errorf("FRONTEND_ORIGINS must contain at least one trusted origin")
	}
	return origins, nil
}
