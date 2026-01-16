package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/rootlock/rootlock/backend/config"
)

// RateLimiter handles rate limiting for API endpoints
type RateLimiter struct {
	redis          *redis.Client
	enabled        bool
	requestsPerMin int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg *config.Config) (*RateLimiter, error) {
	if !cfg.Security.RateLimitEnabled {
		return &RateLimiter{
			enabled: false,
		}, nil
	}

	// Parse Redis URL
	opts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	// Override password if provided
	if cfg.Redis.Password != "" {
		opts.Password = cfg.Redis.Password
	}

	// Create Redis client
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RateLimiter{
		redis:          client,
		enabled:        true,
		requestsPerMin: cfg.Security.RateLimitPerMinute,
	}, nil
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if rate limiting is disabled
		if !rl.enabled {
			c.Next()
			return
		}

		// Determine identifier (IP address or user ID if authenticated)
		identifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		// Get endpoint
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Check rate limit
		allowed, remaining, resetTime, err := rl.checkRateLimit(c.Request.Context(), identifier, endpoint)
		if err != nil {
			// Log error but don't block request on rate limiter failure
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.requestsPerMin))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": time.Until(resetTime).Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit checks if a request is allowed based on rate limits
// Returns: (allowed, remaining, resetTime, error)
func (rl *RateLimiter) checkRateLimit(ctx context.Context, identifier, endpoint string) (bool, int, time.Time, error) {
	now := time.Now()
	window := now.Truncate(time.Minute)
	key := fmt.Sprintf("ratelimit:%s:%s:%d", identifier, endpoint, window.Unix())

	// Use Redis pipeline for atomic increment and expire
	pipe := rl.redis.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 2*time.Minute) // Keep for 2 minutes to handle clock skew

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to execute redis pipeline: %w", err)
	}

	count := int(incrCmd.Val())
	remaining := rl.requestsPerMin - count
	if remaining < 0 {
		remaining = 0
	}

	resetTime := window.Add(time.Minute)
	allowed := count <= rl.requestsPerMin

	return allowed, remaining, resetTime, nil
}

// Close closes the rate limiter's resources
func (rl *RateLimiter) Close() error {
	if rl.redis != nil {
		return rl.redis.Close()
	}
	return nil
}

// StrictRateLimitMiddleware applies stricter rate limits for sensitive endpoints
func StrictRateLimitMiddleware(rl *RateLimiter, requestsPerMin int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if rate limiting is disabled
		if !rl.enabled {
			c.Next()
			return
		}

		// Determine identifier (IP address or user ID if authenticated)
		identifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		// Get endpoint
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Use strict key to separate from normal rate limits
		strictEndpoint := fmt.Sprintf("strict:%s", endpoint)

		// Check rate limit with custom limit
		allowed, remaining, resetTime, err := rl.checkStrictRateLimit(
			c.Request.Context(),
			identifier,
			strictEndpoint,
			requestsPerMin,
		)
		if err != nil {
			// Log error but don't block request on rate limiter failure
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMin))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": time.Until(resetTime).Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkStrictRateLimit is similar to checkRateLimit but with custom limit
func (rl *RateLimiter) checkStrictRateLimit(ctx context.Context, identifier, endpoint string, limit int) (bool, int, time.Time, error) {
	now := time.Now()
	window := now.Truncate(time.Minute)
	key := fmt.Sprintf("ratelimit:%s:%s:%d", identifier, endpoint, window.Unix())

	// Use Redis pipeline for atomic increment and expire
	pipe := rl.redis.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 2*time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to execute redis pipeline: %w", err)
	}

	count := int(incrCmd.Val())
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	resetTime := window.Add(time.Minute)
	allowed := count <= limit

	return allowed, remaining, resetTime, nil
}
