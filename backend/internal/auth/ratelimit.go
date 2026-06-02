package auth

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[ip]; ok {
		return lim
	}
	lim := rate.NewLimiter(rate.Limit(10.0/60.0), 10) // 10 requests per minute, burst 10
	l.limiters[ip] = lim
	return lim
}

var authRateLimiter = &ipLimiter{limiters: make(map[string]*rate.Limiter)}

// AuthRateLimit limits auth endpoints to 10 requests per minute per IP.
func AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authRateLimiter.get(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, slow down"})
			return
		}
		c.Next()
	}
}
