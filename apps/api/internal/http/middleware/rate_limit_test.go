package middleware

import (
	"net/http"
	"net/http/httptest"
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
