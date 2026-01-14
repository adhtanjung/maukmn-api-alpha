package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds common security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent page from being displayed in an iframe (Clickjacking protection)
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS filtering in browser
		c.Header("X-XSS-Protection", "1; mode=block")

		// Control how much referrer information is sent
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy - restricts where resources can be loaded from
		// This is a strict default. You might need to relax it for images/scripts/styles from CDNs
		c.Header("Content-Security-Policy", "default-src 'self'; object-src 'none'")

		// HTTP Strict Transport Security (HSTS) - force HTTPS
		// Standard: 1 year (31536000 seconds)
		// Only apply this in production usually, or if you have a local HTTPS setup
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}
