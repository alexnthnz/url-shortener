package models

import "time"

// URL represents a URL mapping in the database
type URL struct {
	ID          int64      `json:"id" db:"id"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	CustomAlias bool       `json:"custom_alias" db:"custom_alias"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// Analytics represents click analytics for a URL
type Analytics struct {
	ID        int64     `json:"id" db:"id"`
	ShortCode string    `json:"short_code" db:"short_code"`
	ClickedAt time.Time `json:"clicked_at" db:"clicked_at"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
}

// URLStats represents aggregated statistics for a URL
type URLStats struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	ClickCount  int64     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// ShortenRequest represents the request payload for shortening a URL
type ShortenRequest struct {
	URL         string `json:"url" binding:"required,url"`
	CustomAlias string `json:"custom_alias,omitempty"`
}

// ShortenResponse represents the response when creating a short URL
type ShortenResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
