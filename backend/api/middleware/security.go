package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Writer.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		// Strict Transport Security (HTTPS only)
		if c.Request.TLS != nil {
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Referrer Policy
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		c.Writer.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
