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
)

// URLEntry represents a stored URL with metadata
type URLEntry struct {
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// URLStore defines the interface for URL storage
type URLStore interface {
	// Set stores a URL and returns a unique code
	Set(url string) (string, error)
	
	// Get retrieves a URL by its code
	Get(code string) (string, error)
	
	// Stats returns the number of URLs in the store
	Stats() int
}

// InMemoryURLStore implements URLStore with an in-memory map
type InMemoryURLStore struct {
	urls  map[string]URLEntry
	mutex sync.RWMutex
}

// NewInMemoryURLStore creates a new in-memory URL store
func NewInMemoryURLStore() *InMemoryURLStore {
	return &InMemoryURLStore{
		urls: make(map[string]URLEntry),
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
	if url == "" {
		return "", ErrInvalidURL
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Generate a unique code
	code, err := generateCode()
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
	
	// Store the URL with metadata
	s.urls[code] = URLEntry{
		URL:       url,
		CreatedAt: time.Now(),
	}
	
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

// Stats returns the number of URLs in the store
func (s *InMemoryURLStore) Stats() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return len(s.urls)
}
