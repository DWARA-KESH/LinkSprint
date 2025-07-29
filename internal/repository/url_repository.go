package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/<YOUR-GITHUB-USERNAME>/LinkSprint/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type URLRepository struct {
	conn *pgx.Conn
}

func NewURLRepository(conn *pgx.Conn) *URLRepository {
	return &URLRepository{conn: conn}
}

// SaveURL inserts a new URL mapping into the database
func (r *URLRepository) SaveURL(ctx context.Context, url model.URL) error {
	query := `INSERT INTO urls (short_code, original_url, custom_slug) VALUES ($1, $2, $3)`

	_, err := r.conn.Exec(ctx, query,
		url.ShortCode,
		url.OriginalURL,
		url.CustomSlug,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("custom slug '%s' already exists", *url.CustomSlug)
		}
		log.Printf("Failed to save URL to DB: %v", err)
		return fmt.Errorf("failed to save URL: %w", err)
	}
	log.Printf("URL saved to DB: %s -> %s", url.ShortCode, url.OriginalURL)
	return nil
}

// GetURL retrieves a URL record by its short code or custom slug
func (r *URLRepository) GetURL(ctx context.Context, code string) (*model.URL, error) {
	var url model.URL

	err := r.conn.QueryRow(ctx, `
		SELECT short_code, original_url, click_count, custom_slug
		FROM urls
		WHERE short_code = $1
	`, code).Scan(
		&url.ShortCode,
		&url.OriginalURL,
		&url.ClickCount,
		&url.CustomSlug,
	)

	if err == pgx.ErrNoRows {
		err = r.conn.QueryRow(ctx, `
			SELECT short_code, original_url, click_count, custom_slug
			FROM urls
			WHERE custom_slug = $1
		`, code).Scan(
			&url.ShortCode,
			&url.OriginalURL,
			&url.ClickCount,
			&url.CustomSlug,
		)
	}

	if err == pgx.ErrNoRows {
		log.Printf("URL with code '%s' not found in DB.", code)
		return nil, fmt.Errorf("URL with code '%s' not found", code)
	}
	if err != nil {
		log.Printf("Failed to get URL from DB for code '%s': %v", code, err)
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	log.Printf("URL retrieved from DB: %s -> %s", url.ShortCode, url.OriginalURL)
	return &url, nil
}

// IncrementClickCount
func (r *URLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	_, err := r.conn.Exec(ctx, `
		UPDATE urls
		SET click_count = click_count + 1
		WHERE short_code = $1
	`, shortCode)
	if err != nil {
		log.Printf("Failed to increment click count for %s: %v", shortCode, err)
		return fmt.Errorf("failed to increment click count for %s: %w", shortCode, err)
	}
	log.Printf("Click count incremented for %s", shortCode)
	return nil
}

// InitDB accepts the connection string and returns an error
func InitDB(connStr string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	return conn, nil
}
