package store

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// Common errors
var (
	ErrCodeNotFound = errors.New("code not found")
	ErrInvalidURL   = errors.New("invalid URL")
	ErrAliasInUse   = errors.New("custom alias is already in use")
	ErrInvalidAlias = errors.New("invalid alias: must be 3-20 alphanumeric characters")
)

// URLEntry represents a stored URL with metadata
type URLEntry struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// URLStore defines the interface for URL storage
type URLStore interface {
	// Set stores a URL and returns a unique code
	Set(url string) (string, error)

	// SetWithOptions stores a URL with optional custom alias and user ID
	SetWithOptions(url, customAlias, userID string) (string, error)

	// Get retrieves a URL by its code
	Get(code string) (string, error)

	// GetByUser retrieves all URLs for a specific user
	GetByUser(userID string) ([]URLEntry, error)

	// Stats returns the number of URLs in the store
	Stats() int
}

// InMemoryURLStore implements URLStore with an in-memory map
type InMemoryURLStore struct {
	urls     map[string]URLEntry
	userURLs map[string][]string // Maps user IDs to their URL codes
	mutex    sync.RWMutex
}

// NewInMemoryURLStore creates a new in-memory URL store
func NewInMemoryURLStore() *InMemoryURLStore {
	return &InMemoryURLStore{
		urls:     make(map[string]URLEntry),
		userURLs: make(map[string][]string),
	}
}

// generateCode creates a random short code for URLs
func generateCode() (string, error) {
	// Generate 6 bytes of random data
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode to base64 and remove any non-alphanumeric characters
	code := base64.URLEncoding.EncodeToString(b)
	// Trim to 8 characters
	if len(code) > 8 {
		code = code[:8]
	}

	return code, nil
}

// Set stores a URL and returns a unique code
func (s *InMemoryURLStore) Set(url string) (string, error) {
	return s.SetWithOptions(url, "", "anonymous")
}

// SetWithOptions stores a URL with optional custom alias and user ID
func (s *InMemoryURLStore) SetWithOptions(url, customAlias, userID string) (string, error) {
	if url == "" {
		return "", ErrInvalidURL
	}

	// If no user ID is provided, use a default
	if userID == "" {
		userID = "anonymous"
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var code string
	var err error

	// Use custom alias if provided
	if customAlias != "" {
		// Check if alias already exists
		if _, exists := s.urls[customAlias]; exists {
			return "", ErrAliasInUse
		}
		code = customAlias
	} else {
		// Generate a unique code
		code, err = generateCode()
		if err != nil {
			return "", err
		}

		// Ensure the code is unique
		for {
			if _, exists := s.urls[code]; !exists {
				break
			}
			code, err = generateCode()
			if err != nil {
				return "", err
			}
		}
	}

	// Store the URL with metadata
	s.urls[code] = URLEntry{
		Code:      code,
		URL:       url,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	// Add to user's URLs
	s.userURLs[userID] = append(s.userURLs[userID], code)

	return code, nil
}

// Get retrieves a URL by its code
func (s *InMemoryURLStore) Get(code string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, exists := s.urls[code]
	if !exists {
		return "", ErrCodeNotFound
	}

	return entry.URL, nil
}

// GetByUser retrieves all URLs for a specific user
func (s *InMemoryURLStore) GetByUser(userID string) ([]URLEntry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	codes, exists := s.userURLs[userID]
	if !exists {
		return []URLEntry{}, nil
	}

	entries := make([]URLEntry, 0, len(codes))
	for _, code := range codes {
		if entry, ok := s.urls[code]; ok {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// Stats returns the number of URLs in the store
func (s *InMemoryURLStore) Stats() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.urls)
}
