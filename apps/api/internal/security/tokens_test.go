package security

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAccessTokenValidation(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	service := NewTokenService([]byte(strings.Repeat("s", 32)), "test-issuer", time.Minute)
	raw, _, err := service.IssueAccess(42, now)
	if err != nil {
		t.Fatalf("IssueAccess() error = %v", err)
	}
	if id, err := service.ParseAccess(raw, now.Add(30*time.Second)); err != nil || id != 42 {
		t.Fatalf("ParseAccess() = %d, %v", id, err)
	}
	if _, err := service.ParseAccess(raw, now.Add(61*time.Second)); err == nil {
		t.Fatal("expired access token should be rejected")
	}
	other := NewTokenService([]byte(strings.Repeat("o", 32)), "test-issuer", time.Minute)
	if _, err := other.ParseAccess(raw, now); err == nil {
		t.Fatal("token signed with another secret should be rejected")
	}
}

func TestAccessTokenRejectsInvalidRegisteredClaimsAndAlgorithms(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	secret := []byte(strings.Repeat("s", 32))
	service := NewTokenService(secret, "test-issuer", time.Minute)
	registered := func(issuer, subject string, issuedAt, expiresAt *jwt.NumericDate, id string) jwt.RegisteredClaims {
		return jwt.RegisteredClaims{Issuer: issuer, Subject: subject, IssuedAt: issuedAt, ExpiresAt: expiresAt, ID: id}
	}
	validExp := jwt.NewNumericDate(now.Add(time.Minute))
	validIat := jwt.NewNumericDate(now)
	tests := []struct {
		name   string
		claims jwt.RegisteredClaims
		method jwt.SigningMethod
		key    any
	}{
		{"wrong issuer", registered("other", "42", validIat, validExp, "jti"), jwt.SigningMethodHS256, secret},
		{"missing expiration", registered("test-issuer", "42", validIat, nil, "jti"), jwt.SigningMethodHS256, secret},
		{"missing issued at", registered("test-issuer", "42", nil, validExp, "jti"), jwt.SigningMethodHS256, secret},
		{"future issued at", registered("test-issuer", "42", jwt.NewNumericDate(now.Add(time.Minute)), validExp, "jti"), jwt.SigningMethodHS256, secret},
		{"invalid subject", registered("test-issuer", "0", validIat, validExp, "jti"), jwt.SigningMethodHS256, secret},
		{"missing token id", registered("test-issuer", "42", validIat, validExp, ""), jwt.SigningMethodHS256, secret},
		{"algorithm none", registered("test-issuer", "42", validIat, validExp, "jti"), jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType},
		{"bad signature", registered("test-issuer", "42", validIat, validExp, "jti"), jwt.SigningMethodHS256, []byte(strings.Repeat("x", 32))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := jwt.NewWithClaims(test.method, test.claims).SignedString(test.key)
			if err != nil {
				t.Fatalf("sign test token: %v", err)
			}
			if _, err := service.ParseAccess(raw, now); err == nil {
				t.Fatal("ParseAccess() accepted an invalid token")
			}
		})
	}
	valid := jwt.NewWithClaims(jwt.SigningMethodHS256, registered("test-issuer", strconv.FormatInt(42, 10), validIat, validExp, "jti"))
	raw, err := valid.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	if id, err := service.ParseAccess(raw, now.Add(time.Second)); err != nil || id != 42 {
		t.Fatalf("valid claims rejected: id=%d error=%v", id, err)
	}
}

func TestRefreshTokenIsOpaqueHighEntropyAndHashed(t *testing.T) {
	first, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) < 43 || first == second {
		t.Fatalf("refresh token did not have expected entropy: %q", first)
	}
	if !ValidateRefreshToken(first) || ValidateRefreshToken(first+"x") || ValidateRefreshToken("not-a-refresh-token") {
		t.Fatal("refresh token format validation failed")
	}
	hash := HashRefreshToken(first)
	if len(hash) != 64 || hash == first || HashRefreshToken(first) != hash {
		t.Fatal("refresh token hash is not deterministic SHA-256 hex")
	}
}
