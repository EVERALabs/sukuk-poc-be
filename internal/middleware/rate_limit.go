package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.Mutex
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   time.Minute,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old requests
	if timestamps, exists := rl.requests[key]; exists {
		valid := make([]time.Time, 0, len(timestamps))
		for _, ts := range timestamps {
			if ts.After(cutoff) {
				valid = append(valid, ts)
			}
		}
		rl.requests[key] = valid
	}

	// Check if limit exceeded
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

var globalRateLimiter *rateLimiter

func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	if globalRateLimiter == nil {
		globalRateLimiter = newRateLimiter(requestsPerMinute)
	}

	return func(c *gin.Context) {
		// Skip rate limiting for health check
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Use client IP as key
		key := c.ClientIP()
		
		if !globalRateLimiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}