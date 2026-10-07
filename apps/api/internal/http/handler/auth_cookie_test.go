package handler

import (
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
	"github.com/gin-gonic/gin"
)

func TestRefreshCookieSecuritySettings(t *testing.T) {
	for _, test := range []struct {
		name     string
		secure   bool
		sameSite string
		wantMode int
	}{
		{"development lax", false, "lax", 2},
		{"production strict", true, "strict", 3},
		{"production none", true, "none", 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := NewAuth(nil, config.Config{Cookie: config.Cookie{Name: "bem_refresh", Secure: test.secure, SameSite: test.sameSite, Path: "/api/v1/auth"}}, slog.Default())
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			handler.setRefreshCookie(context, "opaque-value", 7*24*time.Hour)
			cookie := response.Result().Cookies()[0]
			if !cookie.HttpOnly || cookie.Secure != test.secure || cookie.Path != "/api/v1/auth" || int(cookie.SameSite) != test.wantMode {
				t.Fatalf("unexpected refresh cookie: %+v", cookie)
			}
		})
	}
}
