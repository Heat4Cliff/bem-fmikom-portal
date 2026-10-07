package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/gin-gonic/gin"
)

const principalKey = "authenticated_user"

type CurrentUserLoader interface {
	CurrentUser(ctx context.Context, rawToken string) (domain.User, error)
}

func RequireAuth(auth CurrentUserLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.Fields(c.GetHeader("Authorization"))
		if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, failure("UNAUTHENTICATED", "Authentication required"))
			return
		}
		user, err := auth.CurrentUser(c.Request.Context(), header[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, failure("UNAUTHENTICATED", "Authentication required"))
			return
		}
		c.Set(principalKey, user)
		c.Next()
	}
}

func RequireRoles(roles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, failure("UNAUTHENTICATED", "Authentication required"))
			return
		}
		if !service.RoleAllowed(user.Role, roles...) {
			c.AbortWithStatusJSON(http.StatusForbidden, failure("FORBIDDEN", "Akses ditolak"))
			return
		}
		c.Next()
	}
}

func failure(code, message string) gin.H {
	return gin.H{"success": false, "message": message, "code": code}
}

func CurrentUser(c *gin.Context) (domain.User, bool) {
	value, exists := c.Get(principalKey)
	if !exists {
		return domain.User{}, false
	}
	user, ok := value.(domain.User)
	return user, ok
}
