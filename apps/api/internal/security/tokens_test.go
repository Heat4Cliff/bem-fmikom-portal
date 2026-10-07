package security

import (
	"strings"
	"testing"
	"time"
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
	hash := HashRefreshToken(first)
	if len(hash) != 64 || hash == first || HashRefreshToken(first) != hash {
		t.Fatal("refresh token hash is not deterministic SHA-256 hex")
	}
}
