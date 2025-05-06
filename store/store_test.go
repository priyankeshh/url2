package store

import (
	"testing"
)

func TestInMemoryURLStore_Set(t *testing.T) {
	store := NewInMemoryURLStore()
	
	// Test valid URL
	url := "https://example.com"
	code, err := store.Set(url)
	if err != nil {
		t.Errorf("Failed to set URL: %v", err)
	}
	if code == "" {
		t.Error("Expected non-empty code")
	}
	
	// Test empty URL
	_, err = store.Set("")
	if err != ErrInvalidURL {
		t.Errorf("Expected ErrInvalidURL, got %v", err)
	}
	
	// Test multiple URLs
	url2 := "https://example.org"
	code2, err := store.Set(url2)
	if err != nil {
		t.Errorf("Failed to set second URL: %v", err)
	}
	if code == code2 {
		t.Error("Expected different codes for different URLs")
	}
}

func TestInMemoryURLStore_Get(t *testing.T) {
	store := NewInMemoryURLStore()
	
	// Set a URL
	url := "https://example.com"
	code, err := store.Set(url)
	if err != nil {
		t.Fatalf("Failed to set URL: %v", err)
	}
	
	// Get the URL
	retrievedURL, err := store.Get(code)
	if err != nil {
		t.Errorf("Failed to get URL: %v", err)
	}
	if retrievedURL != url {
		t.Errorf("Expected URL %s, got %s", url, retrievedURL)
	}
	
	// Test non-existent code
	_, err = store.Get("nonexistent")
	if err != ErrCodeNotFound {
		t.Errorf("Expected ErrCodeNotFound, got %v", err)
	}
}

func TestInMemoryURLStore_Stats(t *testing.T) {
	store := NewInMemoryURLStore()
	
	// Initial stats should be 0
	if stats := store.Stats(); stats != 0 {
		t.Errorf("Expected 0 URLs, got %d", stats)
	}
	
	// Add a URL
	_, err := store.Set("https://example.com")
	if err != nil {
		t.Fatalf("Failed to set URL: %v", err)
	}
	
	// Stats should be 1
	if stats := store.Stats(); stats != 1 {
		t.Errorf("Expected 1 URL, got %d", stats)
	}
	
	// Add another URL
	_, err = store.Set("https://example.org")
	if err != nil {
		t.Fatalf("Failed to set URL: %v", err)
	}
	
	// Stats should be 2
	if stats := store.Stats(); stats != 2 {
		t.Errorf("Expected 2 URLs, got %d", stats)
	}
}
