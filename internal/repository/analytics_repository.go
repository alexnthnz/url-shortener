package repository

import (
	"database/sql"

	"github.com/alexnthnz/url-shortener/internal/models"
)

type AnalyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// RecordClick stores a click event for analytics
func (r *AnalyticsRepository) RecordClick(analytics *models.Analytics) error {
	query := `
		INSERT INTO analytics (short_code, ip_address, user_agent)
		VALUES ($1, $2, $3)
		RETURNING id, clicked_at`

	return r.db.QueryRow(
		query,
		analytics.ShortCode,
		analytics.IPAddress,
		analytics.UserAgent,
	).Scan(&analytics.ID, &analytics.ClickedAt)
}

// GetClickCount returns the total click count for a short code
func (r *AnalyticsRepository) GetClickCount(shortCode string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM analytics WHERE short_code = $1`
	err := r.db.QueryRow(query, shortCode).Scan(&count)
	return count, err
}
