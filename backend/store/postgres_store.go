package store

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"

	_ "github.com/lib/pq"
)

type PostgresURLStore struct {
	db *sql.DB
}

func NewPostgresURLStore(connStr string) (*PostgresURLStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	store := &PostgresURLStore{
		db: db,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return store, nil
}

func (s *PostgresURLStore) initSchema() error {
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

	_, err = s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id)
	`)
	if err != nil {
		return err
	}

	log.Println("Database schema initialized successfully")
	return nil
}

func (s *PostgresURLStore) Close() error {
	return s.db.Close()
}

func validateAlias(alias string) error {
	if len(alias) < 3 || len(alias) > 20 {
		return ErrInvalidAlias
	}

	match, err := regexp.MatchString("^[a-zA-Z0-9]+$", alias)
	if err != nil || !match {
		return ErrInvalidAlias
	}

	return nil
}

func (s *PostgresURLStore) Set(url string) (string, error) {
	return s.SetWithOptions(url, "", "")
}

func (s *PostgresURLStore) SetWithOptions(url, customAlias, userID string) (string, error) {
	if url == "" {
		return "", ErrInvalidURL
	}

	if userID == "" {
		userID = "anonymous"
	}

	var code string
	var err error

	if customAlias != "" {
		if err := validateAlias(customAlias); err != nil {
			return "", err
		}

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
		code, err = generateCode()
		if err != nil {
			return "", err
		}

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

	_, err = s.db.Exec(
		"INSERT INTO urls (code, url, user_id, created_at) VALUES ($1, $2, $3, $4)",
		code, url, userID, time.Now(),
	)
	if err != nil {
		return "", err
	}

	return code, nil
}

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

func (s *PostgresURLStore) Stats() int {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		return 0
	}

	return count
}
