package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestGuardsRequireBoundedJSONAndSaneQueries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestGuards())
	engine.POST("/json", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	engine.GET("/list", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	tests := []struct {
		name   string
		method string
		path   string
		body   io.Reader
		type_  string
		want   int
	}{
		{"oversized json", http.MethodPost, "/json", bytes.NewReader(bytes.Repeat([]byte("x"), int(maxJSONBodyBytes+1))), "application/json", http.StatusRequestEntityTooLarge},
		{"unsupported content type", http.MethodPost, "/json", strings.NewReader("<xml/>"), "application/xml", http.StatusUnsupportedMediaType},
		{"pagination maximum", http.MethodGet, "/list?limit=101", nil, "", http.StatusBadRequest},
		{"duplicate query value", http.MethodGet, "/list?division_id=1&division_id=2", nil, "", http.StatusBadRequest},
		{"unbounded offset", http.MethodGet, "/list?page=9999999999", nil, "", http.StatusBadRequest},
		{"accepted bounded json", http.MethodPost, "/json", strings.NewReader(`{"ok":true}`), "application/json; charset=utf-8", http.StatusNoContent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, test.body)
			if test.type_ != "" {
				request.Header.Set("Content-Type", test.type_)
			}
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status=%d want=%d body=%s", response.Code, test.want, response.Body.String())
			}
		})
	}
}

func TestAPISecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(SecurityHeaders())
	engine.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	for name, want := range map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
	} {
		if got := response.Header().Get(name); got != want {
			t.Errorf("%s=%q want=%q", name, got, want)
		}
	}
}

func TestRecoveryReturnsGenericErrorAndRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	engine := gin.New()
	engine.Use(RequestID(), Recovery(slog.New(slog.NewJSONHandler(&logs, nil))))
	engine.GET("/panic", func(c *gin.Context) { panic("secret panic payload") })
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	request.Header[RequestIDHeader] = []string{"bad\r\nforged: yes"}
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "secret panic payload") || response.Header().Get(RequestIDHeader) == "" {
		t.Fatalf("unsafe panic response: status=%d headers=%v body=%s", response.Code, response.Header(), response.Body.String())
	}
	if strings.Contains(logs.String(), "secret panic payload") {
		t.Fatalf("panic contents leaked to logs: %s", logs.String())
	}
}
