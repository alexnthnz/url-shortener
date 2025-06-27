package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/alexnthnz/url-shortener/internal/models"
	"github.com/alexnthnz/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type URLHandler struct {
	urlService       *services.URLService
	analyticsService *services.AnalyticsService
	logger           *logrus.Logger
}

func NewURLHandler(urlService *services.URLService, analyticsService *services.AnalyticsService, logger *logrus.Logger) *URLHandler {
	return &URLHandler{
		urlService:       urlService,
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// ShortenURL handles POST /api/v1/shorten
func (h *URLHandler) ShortenURL(c *gin.Context) {
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Create short URL
	urlRecord, err := h.urlService.ShortenURL(req.URL, req.CustomAlias)
	if err != nil {
		h.logger.Errorf("Failed to shorten URL: %v", err)

		// Handle specific error cases
		if strings.Contains(err.Error(), "invalid URL") ||
			strings.Contains(err.Error(), "invalid custom alias") ||
			strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short URL"})
		return
	}

	// Build response
	baseURL := c.GetHeader("X-Base-URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Fallback
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	response := models.ShortenResponse{
		ShortCode:   urlRecord.ShortCode,
		ShortURL:    baseURL + "/" + urlRecord.ShortCode,
		OriginalURL: urlRecord.OriginalURL,
	}

	c.JSON(http.StatusCreated, response)
}

// RedirectURL handles GET /:short_code
func (h *URLHandler) RedirectURL(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get original URL
	originalURL, err := h.urlService.GetOriginalURL(shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
			return
		}

		h.logger.Errorf("Failed to get original URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve URL"})
		return
	}

	// Record analytics asynchronously (non-blocking)
	ipAddress := h.getClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	h.analyticsService.RecordClickAsync(shortCode, ipAddress, userAgent)

	// Redirect to original URL immediately
	c.Redirect(http.StatusMovedPermanently, originalURL)
}

// GetURLStats handles GET /api/v1/urls/:short_code/stats
func (h *URLHandler) GetURLStats(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get URL statistics
	stats, err := h.urlService.GetURLStats(shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
			return
		}

		h.logger.Errorf("Failed to get URL stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getClientIP extracts the real client IP address
func (h *URLHandler) getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP from the comma-separated list
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to remote address
	return c.ClientIP()
}

// HealthCheck handles GET /health with comprehensive system checks
func (h *URLHandler) HealthCheck(c *gin.Context) {
	status := "healthy"
	checks := make(map[string]interface{})
	httpStatus := 200

	// Check database connectivity
	if err := h.urlService.HealthCheck(); err != nil {
		checks["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		status = "unhealthy"
		httpStatus = 503
	} else {
		checks["database"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check cache connectivity
	if err := h.urlService.CacheHealthCheck(); err != nil {
		checks["cache"] = map[string]interface{}{
			"status": "degraded",
			"error":  err.Error(),
		}
		if status == "healthy" {
			status = "degraded"
		}
	} else {
		checks["cache"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Add service metadata
	checks["service"] = map[string]interface{}{
		"name":    "url-shortener",
		"version": "1.0.0",
		"uptime":  time.Since(startTime).String(),
	}

	response := gin.H{
		"status": status,
		"checks": checks,
	}

	c.JSON(httpStatus, response)
}

var startTime = time.Now() // Track service start time

// MetricsHandler provides basic metrics for monitoring
func (h *URLHandler) MetricsHandler(c *gin.Context) {
	// This is a basic implementation - in production you'd use Prometheus
	metrics := gin.H{
		"service": gin.H{
			"name":    "url-shortener",
			"version": "1.0.0",
			"uptime":  time.Since(startTime).String(),
		},
		"system": gin.H{
			"timestamp": time.Now().Unix(),
		},
		// Add more metrics as needed
	}

	c.JSON(200, metrics)
}
