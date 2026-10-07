package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type attemptWindow struct {
	count   int
	resetAt time.Time
}

type LoginLimiter struct {
	mu      sync.Mutex
	entries map[string]attemptWindow
	limit   int
	window  time.Duration
	clock   func() time.Time
}

func NewLoginLimiter(limit int, window time.Duration) *LoginLimiter {
	return &LoginLimiter{entries: make(map[string]attemptWindow), limit: limit, window: window, clock: time.Now}
}

func (limiter *LoginLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := limiter.clock()
		key := c.ClientIP()
		limiter.mu.Lock()
		entry, exists := limiter.entries[key]
		if !exists || !now.Before(entry.resetAt) {
			entry = attemptWindow{resetAt: now.Add(limiter.window)}
		}
		if entry.count >= limiter.limit {
			retryAfter := max(1, int(entry.resetAt.Sub(now).Seconds()))
			limiter.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "Terlalu banyak percobaan login", "code": "RATE_LIMITED"})
			return
		}
		entry.count++
		limiter.entries[key] = entry
		if len(limiter.entries) > 4096 {
			for address, window := range limiter.entries {
				if !now.Before(window.resetAt) {
					delete(limiter.entries, address)
				}
			}
		}
		limiter.mu.Unlock()
		c.Next()
	}
}
