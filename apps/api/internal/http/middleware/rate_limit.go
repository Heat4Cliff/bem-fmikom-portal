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
	message string
}

// AspirationLimiter protects anonymous public submission independently from login.
type AspirationLimiter struct {
	mu      sync.Mutex
	entries map[string]attemptWindow
	limit   int
	window  time.Duration
	clock   func() time.Time
}

// RegistrationLimiter independently protects anonymous registration submissions.
type RegistrationLimiter struct {
	mu      sync.Mutex
	entries map[string]attemptWindow
	limit   int
	window  time.Duration
	clock   func() time.Time
}

func NewLoginLimiter(limit int, window time.Duration) *LoginLimiter {
	return &LoginLimiter{entries: make(map[string]attemptWindow), limit: limit, window: window, clock: time.Now, message: "Terlalu banyak percobaan login"}
}

func NewAuthRefreshLimiter(limit int, window time.Duration) *LoginLimiter {
	return &LoginLimiter{entries: make(map[string]attemptWindow), limit: limit, window: window, clock: time.Now, message: "Terlalu banyak percobaan refresh"}
}

func NewAspirationLimiter(limit int, window time.Duration) *AspirationLimiter {
	return &AspirationLimiter{entries: make(map[string]attemptWindow), limit: limit, window: window, clock: time.Now}
}

func NewRegistrationLimiter(limit int, window time.Duration) *RegistrationLimiter {
	return &RegistrationLimiter{entries: make(map[string]attemptWindow), limit: limit, window: window, clock: time.Now}
}

func (limiter *RegistrationLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := limiter.clock()
		key := c.ClientIP()
		limiter.mu.Lock()
		entry, exists := limiter.entries[key]
		if !exists || !now.Before(entry.resetAt) {
			entry = attemptWindow{resetAt: now.Add(limiter.window)}
		}
		if !exists && !makeRoom(limiter.entries, now) {
			limiter.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(max(1, int(limiter.window.Seconds()))))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "Batas kapasitas rate limiter tercapai", "code": "RATE_LIMITED"})
			return
		}
		if entry.count >= limiter.limit {
			retryAfter := max(1, int(entry.resetAt.Sub(now).Seconds()))
			limiter.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "Terlalu banyak pengiriman formulir. Silakan coba beberapa saat lagi.", "code": "RATE_LIMITED"})
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

func (limiter *AspirationLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := limiter.clock()
		key := c.ClientIP()
		limiter.mu.Lock()
		entry, exists := limiter.entries[key]
		if !exists || !now.Before(entry.resetAt) {
			entry = attemptWindow{resetAt: now.Add(limiter.window)}
		}
		if !exists && !makeRoom(limiter.entries, now) {
			limiter.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(max(1, int(limiter.window.Seconds()))))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "Batas kapasitas rate limiter tercapai", "code": "RATE_LIMITED"})
			return
		}
		if entry.count >= limiter.limit {
			retryAfter := max(1, int(entry.resetAt.Sub(now).Seconds()))
			limiter.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "Terlalu banyak pengiriman aspirasi. Silakan coba beberapa saat lagi.", "code": "RATE_LIMITED"})
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

func (limiter *LoginLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := limiter.clock()
		key := c.ClientIP()
		limiter.mu.Lock()
		entry, exists := limiter.entries[key]
		if !exists || !now.Before(entry.resetAt) {
			entry = attemptWindow{resetAt: now.Add(limiter.window)}
		}
		if !exists && !makeRoom(limiter.entries, now) {
			limiter.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(max(1, int(limiter.window.Seconds()))))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "Batas kapasitas rate limiter tercapai", "code": "RATE_LIMITED"})
			return
		}
		if entry.count >= limiter.limit {
			retryAfter := max(1, int(entry.resetAt.Sub(now).Seconds()))
			limiter.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": limiter.message, "code": "RATE_LIMITED"})
			return
		}
		entry.count++
		limiter.entries[key] = entry
		limiter.mu.Unlock()
		c.Next()
	}
}

func makeRoom(entries map[string]attemptWindow, now time.Time) bool {
	if len(entries) < 4096 {
		return true
	}
	for address, window := range entries {
		if !now.Before(window.resetAt) {
			delete(entries, address)
		}
	}
	return len(entries) < 4096
}
