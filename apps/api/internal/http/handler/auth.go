package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/gin-gonic/gin"
)

type Auth struct {
	service *service.AuthService
	cfg     config.Config
	logger  *slog.Logger
}

func NewAuth(authService *service.AuthService, cfg config.Config, logger *slog.Logger) *Auth {
	return &Auth{service: authService, cfg: cfg, logger: logger}
}

func (handler *Auth) Login(c *gin.Context) {
	var input service.AuthInput
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16*1024)
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, FailureCode("VALIDATION_ERROR", "Periksa kembali data login", nil))
		return
	}
	result, err := handler.service.Login(c.Request.Context(), input, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		handler.writeAuthError(c, "login", err)
		return
	}
	handler.setRefreshCookie(c, result.RefreshToken, handler.cfg.RefreshTokenTTL)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, Success("Login berhasil", authPayload(result)))
}

func (handler *Auth) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(handler.cfg.Cookie.Name)
	if err != nil {
		c.JSON(http.StatusUnauthorized, FailureCode("UNAUTHENTICATED", "Authentication required", nil))
		return
	}
	result, err := handler.service.Refresh(c.Request.Context(), refreshToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		handler.writeAuthError(c, "refresh", err)
		return
	}
	handler.setRefreshCookie(c, result.RefreshToken, handler.cfg.RefreshTokenTTL)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, Success("Token diperbarui", authPayload(result)))
}

func (handler *Auth) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(handler.cfg.Cookie.Name)
	var authenticatedUserID *int64
	if rawToken := bearerToken(c.GetHeader("Authorization")); rawToken != "" {
		if user, err := handler.service.CurrentUser(c.Request.Context(), rawToken); err == nil {
			id := user.ID
			authenticatedUserID = &id
		}
	}
	if err := handler.service.Logout(c.Request.Context(), refreshToken, authenticatedUserID, c.ClientIP(), c.Request.UserAgent()); err != nil {
		handler.writeAuthError(c, "logout", err)
		return
	}
	handler.clearRefreshCookie(c)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, Success("Logout berhasil", nil))
}

func (handler *Auth) Me(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, FailureCode("UNAUTHENTICATED", "Authentication required", nil))
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, Success("Data berhasil diambil", user))
}

func (handler *Auth) writeAuthError(c *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		c.JSON(http.StatusBadRequest, FailureCode("VALIDATION_ERROR", "Periksa kembali data yang dikirim", nil))
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, FailureCode("UNAUTHENTICATED", "Email atau password tidak valid", nil))
	case errors.Is(err, service.ErrUnauthenticated):
		c.JSON(http.StatusUnauthorized, FailureCode("UNAUTHENTICATED", "Authentication required", nil))
	case errors.Is(err, service.ErrConflict):
		c.JSON(http.StatusConflict, FailureCode("CONFLICT", "Data akun sudah ada", nil))
	default:
		requestID, _ := c.Get("request_id")
		requestIDValue, _ := requestID.(string)
		handler.logger.Error("auth_operation_failed", slog.String("operation", operation), slog.String("request_id", requestIDValue))
		c.JSON(http.StatusInternalServerError, FailureCode("INTERNAL_ERROR", "Terjadi kesalahan pada server", nil))
	}
}

func (handler *Auth) setRefreshCookie(c *gin.Context, token string, ttl time.Duration) {
	cookie := &http.Cookie{Name: handler.cfg.Cookie.Name, Value: token, Path: handler.cfg.Cookie.Path, Domain: handler.cfg.Cookie.Domain, MaxAge: int(ttl.Seconds()), HttpOnly: true, Secure: handler.cfg.Cookie.Secure, SameSite: sameSiteMode(handler.cfg.Cookie.SameSite)}
	http.SetCookie(c.Writer, cookie)
}

func (handler *Auth) clearRefreshCookie(c *gin.Context) {
	cookie := &http.Cookie{Name: handler.cfg.Cookie.Name, Value: "", Path: handler.cfg.Cookie.Path, Domain: handler.cfg.Cookie.Domain, MaxAge: -1, Expires: time.Unix(1, 0), HttpOnly: true, Secure: handler.cfg.Cookie.Secure, SameSite: sameSiteMode(handler.cfg.Cookie.SameSite)}
	http.SetCookie(c.Writer, cookie)
}

func sameSiteMode(value string) http.SameSite {
	switch value {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func authPayload(result service.AuthResult) gin.H {
	return gin.H{"access_token": result.AccessToken, "expires_in": result.ExpiresIn, "user": result.User}
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
