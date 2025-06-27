package repository

import (
	"database/sql"

	"github.com/alexnthnz/url-shortener/internal/models"
)

type URLRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRepository {
	return &URLRepository{db: db}
}

// Create stores a new URL mapping in the database
func (r *URLRepository) Create(url *models.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, custom_alias, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.db.QueryRow(
		query,
		url.ShortCode,
		url.OriginalURL,
		url.CustomAlias,
		url.ExpiresAt,
	).Scan(&url.ID, &url.CreatedAt)
}

// GetByShortCode retrieves a URL by its short code
func (r *URLRepository) GetByShortCode(shortCode string) (*models.URL, error) {
	url := &models.URL{}
	query := `
		SELECT id, short_code, original_url, custom_alias, created_at, expires_at
		FROM urls
		WHERE short_code = $1`

	err := r.db.QueryRow(query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CustomAlias,
		&url.CreatedAt,
		&url.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return url, err
}

// Exists checks if a short code already exists
func (r *URLRepository) Exists(shortCode string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`
	err := r.db.QueryRow(query, shortCode).Scan(&exists)
	return exists, err
}

// GetNextID returns the next sequential ID for generating short codes
func (r *URLRepository) GetNextID() (int64, error) {
	var nextID int64
	// Use atomic sequence to prevent race conditions in concurrent environments
	query := `SELECT nextval('url_id_sequence')`
	err := r.db.QueryRow(query).Scan(&nextID)
	return nextID, err
}

// GetStats retrieves statistics for a URL
func (r *URLRepository) GetStats(shortCode string) (*models.URLStats, error) {
	stats := &models.URLStats{}
	query := `
		SELECT 
			u.short_code,
			u.original_url,
			u.created_at,
			COALESCE(COUNT(a.id), 0) as click_count
		FROM urls u
		LEFT JOIN analytics a ON u.short_code = a.short_code
		WHERE u.short_code = $1
		GROUP BY u.short_code, u.original_url, u.created_at`

	err := r.db.QueryRow(query, shortCode).Scan(
		&stats.ShortCode,
		&stats.OriginalURL,
		&stats.CreatedAt,
		&stats.ClickCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return stats, err
}

// HealthCheck performs a simple database connectivity test
func (r *URLRepository) HealthCheck() (bool, error) {
	// Simple query to test database connectivity
	var result int
	query := `SELECT 1`
	err := r.db.QueryRow(query).Scan(&result)
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
