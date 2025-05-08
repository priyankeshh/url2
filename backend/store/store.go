package store

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var (
	ErrCodeNotFound = errors.New("code not found")
	ErrInvalidURL   = errors.New("invalid URL")
	ErrAliasInUse   = errors.New("custom alias is already in use")
	ErrInvalidAlias = errors.New("invalid alias: must be 3-20 alphanumeric characters")
)

type URLEntry struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type URLStore interface {
	Set(url string) (string, error)
	SetWithOptions(url, customAlias, userID string) (string, error)
	Get(code string) (string, error)
	GetByUser(userID string) ([]URLEntry, error)
	Stats() int
}

type InMemoryURLStore struct {
	urls     map[string]URLEntry
	userURLs map[string][]string
	mutex    sync.RWMutex
}

func NewInMemoryURLStore() *InMemoryURLStore {
	return &InMemoryURLStore{
		urls:     make(map[string]URLEntry),
		userURLs: make(map[string][]string),
	}
}

func generateCode() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	code := base64.URLEncoding.EncodeToString(b)
	if len(code) > 8 {
		code = code[:8]
	}

	return code, nil
}

func (s *InMemoryURLStore) Set(url string) (string, error) {
	return s.SetWithOptions(url, "", "anonymous")
}

func (s *InMemoryURLStore) SetWithOptions(url, customAlias, userID string) (string, error) {
	if url == "" {
		return "", ErrInvalidURL
	}

	if userID == "" {
		userID = "anonymous"
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var code string
	var err error

	if customAlias != "" {
		if _, exists := s.urls[customAlias]; exists {
			return "", ErrAliasInUse
		}
		code = customAlias
	} else {
		code, err = generateCode()
		if err != nil {
			return "", err
		}

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

	s.urls[code] = URLEntry{
		Code:      code,
		URL:       url,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	s.userURLs[userID] = append(s.userURLs[userID], code)

	return code, nil
}

func (s *InMemoryURLStore) Get(code string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, exists := s.urls[code]
	if !exists {
		return "", ErrCodeNotFound
	}

	return entry.URL, nil
}

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

func (s *InMemoryURLStore) Stats() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.urls)
}
