package store

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"

	_ "github.com/lib/pq"
)

// PostgresURLStore implements URLStore with a PostgreSQL database
type PostgresURLStore struct {
	db *sql.DB
}

// NewPostgresURLStore creates a new PostgreSQL URL store
func NewPostgresURLStore(connStr string) (*PostgresURLStore, error) {
	// Open PostgreSQL database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create store instance
	store := &PostgresURLStore{
		db: db,
	}

	// Initialize database schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return store, nil
}

// initSchema creates the necessary tables if they don't exist
func (s *PostgresURLStore) initSchema() error {
	// Create URLs table
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			code TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index on user_id for faster lookups
	_, err = s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id)
	`)
	if err != nil {
		return err
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// Close closes the database connection
func (s *PostgresURLStore) Close() error {
	return s.db.Close()
}

// validateAlias checks if an alias is valid
func validateAlias(alias string) error {
	// Check length
	if len(alias) < 3 || len(alias) > 20 {
		return ErrInvalidAlias
	}

	// Check if it's alphanumeric
	match, err := regexp.MatchString("^[a-zA-Z0-9]+$", alias)
	if err != nil || !match {
		return ErrInvalidAlias
	}

	return nil
}

// Set stores a URL and returns a unique code
func (s *PostgresURLStore) Set(url string) (string, error) {
	return s.SetWithOptions(url, "", "")
}

// SetWithOptions stores a URL with optional custom alias and user ID
func (s *PostgresURLStore) SetWithOptions(url, customAlias, userID string) (string, error) {
	if url == "" {
		return "", ErrInvalidURL
	}

	// If no user ID is provided, use a default
	if userID == "" {
		userID = "anonymous"
	}

	var code string
	var err error

	// Use custom alias if provided
	if customAlias != "" {
		// Validate custom alias
		if err := validateAlias(customAlias); err != nil {
			return "", err
		}

		// Check if alias already exists
		var exists bool
		err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE code = $1)", customAlias).Scan(&exists)
		if err != nil {
			return "", err
		}
		if exists {
			return "", ErrAliasInUse
		}

		code = customAlias
	} else {
		// Generate a random code
		code, err = generateCode()
		if err != nil {
			return "", err
		}

		// Ensure the code is unique
		for {
			var exists bool
			err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE code = $1)", code).Scan(&exists)
			if err != nil {
				return "", err
			}
			if !exists {
				break
			}
			code, err = generateCode()
			if err != nil {
				return "", err
			}
		}
	}

	// Insert the URL into the database
	_, err = s.db.Exec(
		"INSERT INTO urls (code, url, user_id, created_at) VALUES ($1, $2, $3, $4)",
		code, url, userID, time.Now(),
	)
	if err != nil {
		return "", err
	}

	return code, nil
}

// Get retrieves a URL by its code
func (s *PostgresURLStore) Get(code string) (string, error) {
	var url string
	err := s.db.QueryRow("SELECT url FROM urls WHERE code = $1", code).Scan(&url)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrCodeNotFound
		}
		return "", err
	}

	return url, nil
}

// GetByUser retrieves all URLs for a specific user
func (s *PostgresURLStore) GetByUser(userID string) ([]URLEntry, error) {
	rows, err := s.db.Query(
		"SELECT code, url, created_at FROM urls WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []URLEntry
	for rows.Next() {
		var code, url string
		var createdAt time.Time
		if err := rows.Scan(&code, &url, &createdAt); err != nil {
			return nil, err
		}
		entries = append(entries, URLEntry{
			Code:      code,
			URL:       url,
			CreatedAt: createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// Stats returns the number of URLs in the store
func (s *PostgresURLStore) Stats() int {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		return 0
	}

	return count
}
