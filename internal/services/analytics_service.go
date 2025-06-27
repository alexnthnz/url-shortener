package services

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alexnthnz/url-shortener/internal/models"
	"github.com/alexnthnz/url-shortener/internal/repository"
	"github.com/sirupsen/logrus"
)

// AnalyticsEvent represents an analytics event to be processed
type AnalyticsEvent struct {
	ShortCode string
	IPAddress string
	UserAgent string
	Timestamp time.Time
}

type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepository
	logger        *logrus.Logger
	eventQueue    chan AnalyticsEvent
	batchSize     int
	flushInterval time.Duration
}

func NewAnalyticsService(analyticsRepo *repository.AnalyticsRepository, logger *logrus.Logger) *AnalyticsService {
	service := &AnalyticsService{
		analyticsRepo: analyticsRepo,
		logger:        logger,
		eventQueue:    make(chan AnalyticsEvent, 10000), // Buffered channel for async processing
		batchSize:     100,
		flushInterval: 5 * time.Second,
	}

	// Start async processor
	go service.processEvents()

	return service
}

// RecordClickAsync queues a click event for async processing (non-blocking)
func (s *AnalyticsService) RecordClickAsync(shortCode, ipAddress, userAgent string) {
	event := AnalyticsEvent{
		ShortCode: shortCode,
		IPAddress: s.sanitizeIPAddress(ipAddress),
		UserAgent: s.sanitizeUserAgent(userAgent),
		Timestamp: time.Now(),
	}

	// Non-blocking send to queue
	select {
	case s.eventQueue <- event:
		// Event queued successfully
	default:
		// Queue is full, log warning but don't block redirect
		s.logger.Warn("Analytics queue full, dropping click event")
	}
}

// RecordClick records a click event for analytics (blocking - for backward compatibility)
func (s *AnalyticsService) RecordClick(shortCode, ipAddress, userAgent string) error {
	// Sanitize inputs
	cleanIP := s.sanitizeIPAddress(ipAddress)
	cleanUserAgent := s.sanitizeUserAgent(userAgent)

	analytics := &models.Analytics{
		ShortCode: shortCode,
		IPAddress: cleanIP,
		UserAgent: cleanUserAgent,
	}

	if err := s.analyticsRepo.RecordClick(analytics); err != nil {
		return fmt.Errorf("failed to record click: %w", err)
	}

	s.logger.Infof("Click recorded for short code: %s", shortCode)
	return nil
}

// processEvents processes analytics events asynchronously in batches
func (s *AnalyticsService) processEvents() {
	batch := make([]*models.Analytics, 0, s.batchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case event := <-s.eventQueue:
			analytics := &models.Analytics{
				ShortCode: event.ShortCode,
				IPAddress: event.IPAddress,
				UserAgent: event.UserAgent,
			}
			batch = append(batch, analytics)

			// Flush batch if it reaches target size
			if len(batch) >= s.batchSize {
				s.flushBatch(batch)
				batch = batch[:0] // Reset slice
			}

		case <-ticker.C:
			// Flush batch on timer
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = batch[:0] // Reset slice
			}
		}
	}
}

// flushBatch processes a batch of analytics events
func (s *AnalyticsService) flushBatch(batch []*models.Analytics) {
	for _, analytics := range batch {
		if err := s.analyticsRepo.RecordClick(analytics); err != nil {
			s.logger.Errorf("Failed to record click in batch: %v", err)
		}
	}
	s.logger.Debugf("Processed analytics batch of %d events", len(batch))
}

// GetClickCount returns the total click count for a short code
func (s *AnalyticsService) GetClickCount(shortCode string) (int64, error) {
	count, err := s.analyticsRepo.GetClickCount(shortCode)
	if err != nil {
		return 0, fmt.Errorf("failed to get click count: %w", err)
	}
	return count, nil
}

// sanitizeIPAddress cleans and validates IP address
func (s *AnalyticsService) sanitizeIPAddress(ipAddress string) string {
	// Handle X-Forwarded-For header (take the first IP)
	if strings.Contains(ipAddress, ",") {
		ipAddress = strings.TrimSpace(strings.Split(ipAddress, ",")[0])
	}

	// Validate IP address
	if net.ParseIP(ipAddress) == nil {
		return "unknown"
	}

	return ipAddress
}

// sanitizeUserAgent cleans user agent string
func (s *AnalyticsService) sanitizeUserAgent(userAgent string) string {
	// Limit length to prevent storage issues
	maxLength := 500
	if len(userAgent) > maxLength {
		userAgent = userAgent[:maxLength]
	}

	// Basic sanitization
	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		return "unknown"
	}

	return userAgent
}
