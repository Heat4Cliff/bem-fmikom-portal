package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

const RequestIDHeader = "X-Request-ID"

var safeRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if !safeRequestID.MatchString(requestID) {
			requestID = newRequestID()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)
		c.Next()
	}
}

func newRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err == nil {
		return hex.EncodeToString(value[:])
	}
	// Keep serving requests if the OS random source is unavailable. The timestamp
	// fallback is only for log correlation and is not used as a security token.
	return "fallback-" + time.Now().UTC().Format("20060102T150405.000000000Z07:00")
}

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		requestID, _ := c.Get("request_id")
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		logger.Info("http_request",
			slog.String("request_id", requestIDString(requestID)),
			slog.String("method", c.Request.Method),
			slog.String("route", route),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(started)),
		)
	}
}

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) {
		requestID, _ := c.Get("request_id")
		logger.Error("http_panic_recovered", slog.String("request_id", requestIDString(requestID)))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error", "code": "INTERNAL_ERROR"})
	})
}

func requestIDString(value any) string {
	if requestID, ok := value.(string); ok {
		return requestID
	}
	return "unknown"
}
