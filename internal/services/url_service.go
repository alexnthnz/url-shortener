package services

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/alexnthnz/url-shortener/internal/models"
	"github.com/alexnthnz/url-shortener/internal/repository"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type URLService struct {
	urlRepo *repository.URLRepository
	cache   *repository.RedisCache
	logger  *logrus.Logger
}

func NewURLService(urlRepo *repository.URLRepository, cache *repository.RedisCache, logger *logrus.Logger) *URLService {
	return &URLService{
		urlRepo: urlRepo,
		cache:   cache,
		logger:  logger,
	}
}

// ShortenURL creates a short URL from a long URL
func (s *URLService) ShortenURL(originalURL, customAlias string) (*models.URL, error) {
	// Validate and normalize URL
	if err := s.validateURL(originalURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	normalizedURL := s.normalizeURL(originalURL)

	var shortCode string
	var isCustom bool

	if customAlias != "" {
		// Validate custom alias
		if err := s.validateCustomAlias(customAlias); err != nil {
			return nil, fmt.Errorf("invalid custom alias: %w", err)
		}

		// Check if custom alias already exists
		exists, err := s.urlRepo.Exists(customAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to check alias existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("custom alias already exists")
		}

		shortCode = customAlias
		isCustom = true
	} else {
		// Generate short code using counter-based approach
		nextID, err := s.urlRepo.GetNextID()
		if err != nil {
			return nil, fmt.Errorf("failed to get next ID: %w", err)
		}
		shortCode = s.encodeBase62(nextID)
	}

	// Create URL record
	urlRecord := &models.URL{
		ShortCode:   shortCode,
		OriginalURL: normalizedURL,
		CustomAlias: isCustom,
	}

	if err := s.urlRepo.Create(urlRecord); err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Cache the mapping
	if err := s.cache.Set(shortCode, normalizedURL); err != nil {
		s.logger.Warnf("Failed to cache URL mapping: %v", err)
	}

	return urlRecord, nil
}

// GetOriginalURL retrieves the original URL for a short code
func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	// Try cache first
	originalURL, err := s.cache.Get(shortCode)
	if err == nil {
		return originalURL, nil
	}

	// If not in cache or cache error, query database
	if err != redis.Nil {
		s.logger.Warnf("Cache error: %v", err)
	}

	urlRecord, err := s.urlRepo.GetByShortCode(shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}
	if urlRecord == nil {
		return "", fmt.Errorf("URL not found")
	}

	// Cache the result
	if err := s.cache.Set(shortCode, urlRecord.OriginalURL); err != nil {
		s.logger.Warnf("Failed to cache URL mapping: %v", err)
	}

	return urlRecord.OriginalURL, nil
}

// GetURLStats retrieves statistics for a URL
func (s *URLService) GetURLStats(shortCode string) (*models.URLStats, error) {
	stats, err := s.urlRepo.GetStats(shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL stats: %w", err)
	}
	if stats == nil {
		return nil, fmt.Errorf("URL not found")
	}
	return stats, nil
}

// encodeBase62 converts an integer to base62 string
func (s *URLService) encodeBase62(num int64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	var result strings.Builder
	for num > 0 {
		result.WriteByte(base62Chars[num%62])
		num /= 62
	}

	// Reverse the string
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// validateURL validates and checks if URL is safe
func (s *URLService) validateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("malformed URL")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS URLs are allowed")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	// Basic security check for malicious URLs
	maliciousPatterns := []string{
		"javascript:",
		"data:",
		"file:",
		"ftp:",
	}

	lowerURL := strings.ToLower(rawURL)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(lowerURL, pattern) {
			return fmt.Errorf("potentially malicious URL detected")
		}
	}

	return nil
}

// validateCustomAlias validates custom alias format
func (s *URLService) validateCustomAlias(alias string) error {
	if len(alias) < 3 || len(alias) > 20 {
		return fmt.Errorf("custom alias must be between 3 and 20 characters")
	}

	// Allow alphanumeric characters, hyphens, and underscores
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", alias)
	if err != nil {
		return fmt.Errorf("regex error")
	}
	if !matched {
		return fmt.Errorf("custom alias can only contain letters, numbers, hyphens, and underscores")
	}

	// Reserved words
	reserved := []string{"api", "health", "admin", "www", "app", "short", "url"}
	for _, word := range reserved {
		if strings.ToLower(alias) == word {
			return fmt.Errorf("custom alias cannot be a reserved word")
		}
	}

	return nil
}

// normalizeURL normalizes the URL format
func (s *URLService) normalizeURL(rawURL string) string {
	parsedURL, _ := url.Parse(rawURL)

	// Ensure scheme is present
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	// Remove trailing slash for consistency
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")

	return parsedURL.String()
}

// HealthCheck verifies database connectivity
func (s *URLService) HealthCheck() error {
	// Test database connectivity with a simple query
	_, err := s.urlRepo.HealthCheck()
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// CacheHealthCheck verifies cache connectivity
func (s *URLService) CacheHealthCheck() error {
	// Test cache connectivity
	if err := s.cache.Ping(); err != nil {
		return fmt.Errorf("cache health check failed: %w", err)
	}
	return nil
}
