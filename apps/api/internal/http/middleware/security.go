package middleware

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	maxJSONBodyBytes = int64(2 << 20)
	maxQueryBytes    = 8 << 10
	maxPageSize      = 100
	maxListOffset    = 10_000_000
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		header.Set("Cross-Origin-Opener-Policy", "same-origin")
		header.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Next()
	}
}

// RequestGuards bounds API inputs before handlers decode them. Private document
// uploads have a route-specific limit and are the only accepted multipart body.
func RequestGuards() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.URL.RawQuery) > maxQueryBytes {
			badRequest(c, "Query terlalu panjang")
			return
		}
		for key, values := range c.Request.URL.Query() {
			if len(values) != 1 {
				badRequest(c, "Parameter query tidak boleh berulang")
				return
			}
			if key == "search" && len(values[0]) > 256 {
				badRequest(c, "Parameter pencarian terlalu panjang")
				return
			}
		}
		if err := validatePagination(c); err != nil {
			badRequest(c, err.Error())
			return
		}

		request := c.Request
		if request.Method == http.MethodGet || request.Method == http.MethodHead || request.Method == http.MethodOptions || request.ContentLength == 0 {
			c.Next()
			return
		}
		mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
		if err != nil {
			unsupportedMediaType(c)
			return
		}
		if request.Method == http.MethodPost && mediaType == "multipart/form-data" && (request.URL.Path == "/api/v1/admin/documents" || request.URL.Path == "/api/v1/admin/media" || strings.HasPrefix(request.URL.Path, "/api/v1/public/registrations/") && strings.HasSuffix(request.URL.Path, "/submissions")) {
			c.Next()
			return
		}
		if mediaType != "application/json" && !strings.HasSuffix(mediaType, "+json") {
			unsupportedMediaType(c)
			return
		}
		if request.ContentLength > maxJSONBodyBytes {
			tooLarge(c)
			return
		}
		request.Body = http.MaxBytesReader(c.Writer, request.Body, maxJSONBodyBytes)
		c.Next()
	}
}

func validatePagination(c *gin.Context) error {
	query := c.Request.URL.Query()
	page, limit := 1, 20
	if raw := query.Get("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return fmt.Errorf("page harus bilangan positif")
		}
		page = value
	}
	if raw := query.Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > maxPageSize {
			return fmt.Errorf("limit harus berada di antara 1 dan %d", maxPageSize)
		}
		limit = value
	}
	if int64(page-1) > maxListOffset/int64(limit) {
		return fmt.Errorf("offset pagination melewati batas")
	}
	return nil
}

func badRequest(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "message": message, "code": "BAD_REQUEST"})
}

func unsupportedMediaType(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"success": false, "message": "Content-Type tidak didukung", "code": "UNSUPPORTED_MEDIA_TYPE"})
}

func tooLarge(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"success": false, "message": "Ukuran request melewati batas", "code": "REQUEST_TOO_LARGE"})
}
