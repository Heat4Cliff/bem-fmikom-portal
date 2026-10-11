package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestLoginLimiterReturnsRateLimitResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewLoginLimiter(1, time.Minute)
	engine := gin.New()
	engine.Use(limiter.Middleware())
	engine.POST("/login", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	first := httptest.NewRecorder()
	engine.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/login", nil))
	if first.Code != http.StatusNoContent {
		t.Fatalf("first attempt status = %d", first.Code)
	}
	second := httptest.NewRecorder()
	engine.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/login", nil))
	if second.Code != http.StatusTooManyRequests || second.Header().Get("Retry-After") == "" {
		t.Fatalf("second attempt should be rate limited: %d %v", second.Code, second.Header())
	}
}

func TestAspirationLimiterHasIndependent429Response(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.POST("/aspirations", NewAspirationLimiter(1, time.Minute).Middleware(), func(c *gin.Context) { c.Status(http.StatusCreated) })
	first := httptest.NewRecorder()
	engine.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/aspirations", nil))
	second := httptest.NewRecorder()
	engine.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/aspirations", nil))
	if first.Code != http.StatusCreated || second.Code != http.StatusTooManyRequests || second.Header().Get("Retry-After") == "" {
		t.Fatalf("aspiration limiter responses: first=%d second=%d headers=%v", first.Code, second.Code, second.Header())
	}
}

func TestRegistrationLimiterHasIndependent429Response(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.POST("/registrations", NewRegistrationLimiter(1, 30*time.Minute).Middleware(), func(c *gin.Context) { c.Status(http.StatusCreated) })
	first := httptest.NewRecorder()
	engine.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/registrations", nil))
	second := httptest.NewRecorder()
	engine.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/registrations", nil))
	if first.Code != http.StatusCreated || second.Code != http.StatusTooManyRequests || second.Header().Get("Retry-After") == "" {
		t.Fatalf("registration limiter responses: first=%d second=%d headers=%v", first.Code, second.Code, second.Header())
	}
}

func TestRefreshLimiterReturnsScoped429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.POST("/refresh", NewAuthRefreshLimiter(1, time.Minute).Middleware(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	first := httptest.NewRecorder()
	engine.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/refresh", nil))
	second := httptest.NewRecorder()
	engine.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/refresh", nil))
	if first.Code != http.StatusNoContent || second.Code != http.StatusTooManyRequests || second.Header().Get("Retry-After") == "" || !strings.Contains(second.Body.String(), "percobaan refresh") {
		t.Fatalf("refresh limiter response: first=%d second=%d headers=%v body=%s", first.Code, second.Code, second.Header(), second.Body.String())
	}
}

func TestLimiterCapacityIsBounded(t *testing.T) {
	now := time.Now()
	entries := make(map[string]attemptWindow, 4096)
	for i := 0; i < 4096; i++ {
		entries[strconv.Itoa(i)] = attemptWindow{count: 1, resetAt: now.Add(time.Minute)}
	}
	if makeRoom(entries, now) || len(entries) != 4096 {
		t.Fatalf("active limiter entries should stay bounded: makeRoom=%t size=%d", makeRoom(entries, now), len(entries))
	}
	for key, entry := range entries {
		entry.resetAt = now.Add(-time.Second)
		entries[key] = entry
	}
	if !makeRoom(entries, now) || len(entries) != 0 {
		t.Fatalf("expired limiter entries should be removed: size=%d", len(entries))
	}
}
