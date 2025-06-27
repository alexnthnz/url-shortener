package handlers

import (
	"fmt"
	"time"

	"github.com/alexnthnz/url-shortener/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware creates a Gin middleware for logging
func LoggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.WithFields(logrus.Fields{
			"status":     param.StatusCode,
			"method":     param.Method,
			"path":       param.Path,
			"ip":         param.ClientIP,
			"latency":    param.Latency,
			"user_agent": param.Request.UserAgent(),
		}).Info("HTTP Request")
		return ""
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// SecurityMiddleware adds security headers
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RateLimitMiddleware implements distributed rate limiting using Redis
func RateLimitMiddleware(cache *repository.RedisCache) gin.HandlerFunc {
	const (
		maxRequests = 100
		timeWindow  = time.Minute
	)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Get current count from Redis
		countStr, err := cache.Get(key)
		var count int
		if err != nil {
			// Key doesn't exist, start with 1
			count = 1
			if err := cache.SetWithTTL(key, "1", timeWindow); err != nil {
				// If Redis fails, allow request but log error
				c.Next()
				return
			}
		} else {
			// Parse current count
			fmt.Sscanf(countStr, "%d", &count)
			count++

			// Check if rate limit exceeded
			if count > maxRequests {
				c.JSON(429, gin.H{
					"error":   "Rate limit exceeded",
					"message": fmt.Sprintf("Maximum %d requests per minute allowed", maxRequests),
				})
				c.Abort()
				return
			}

			// Increment counter
			if err := cache.SetWithTTL(key, fmt.Sprintf("%d", count), timeWindow); err != nil {
				// If Redis fails, allow request but log error
				c.Next()
				return
			}
		}

		c.Next()
	}
}

// Deprecated: InMemoryRateLimitMiddleware - kept for backward compatibility
// Use RateLimitMiddleware with Redis instead
func InMemoryRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a no-op now - use RateLimitMiddleware instead
		c.Next()
	}
}
