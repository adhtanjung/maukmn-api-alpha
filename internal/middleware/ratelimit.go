package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiters for each IP address
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new rate limiter
// r: requests per second
// b: burst size
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// Clean up old entries periodically to prevent memory leak
	go i.cleanupLoop()

	return i
}

// AddIP creates a new limiter for an IP if it doesn't exist
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// GetLimiter returns the limiter for a given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]
	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}
	i.mu.Unlock()
	return limiter
}

// removeOldIPs is a naive cleanup. In a real app, you'd track last access time.
// For now, let's just clear the map every hour or so, or implement a proper LRU/expiry.
// For simplicity in this iteration, we'll skip complex cleanup logic to keep it simple,
// but let's add a placeholder.
func (i *IPRateLimiter) cleanupLoop() {
	for {
		time.Sleep(1 * time.Hour)
		i.mu.Lock()
		// Reset map (simple but effective for refreshing)
		log.Println("Cleaning up rate limiter map")
		i.ips = make(map[string]*rate.Limiter)
		i.mu.Unlock()
	}
}

// RateLimit middleware
func RateLimit() gin.HandlerFunc {
	// 20 requests per second, burst of 50
	limiter := NewIPRateLimiter(20, 50)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "Too many requests",
			})
			return
		}
		c.Next()
	}
}
