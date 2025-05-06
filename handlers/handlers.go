package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"url-shortener/store"
)

// URLHandler handles URL shortening and redirection
type URLHandler struct {
	store store.URLStore
	host  string
}

// ShortenRequest represents the request body for shortening a URL
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse represents the response body for a shortened URL
type ShortenResponse struct {
	Code string `json:"code"`
	URL  string `json:"url,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewURLHandler creates a new URL handler
func NewURLHandler(store store.URLStore, host string) *URLHandler {
	return &URLHandler{
		store: store,
		host:  host,
	}
}

// ShortenHandler handles POST /api/shorten requests
func (h *URLHandler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate URL
	if req.URL == "" {
		sendJSONError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Store URL and get code
	code, err := h.store.Set(req.URL)
	if err != nil {
		sendJSONError(w, fmt.Sprintf("Failed to shorten URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Create short URL
	shortURL := fmt.Sprintf("%s/r/%s", h.host, code)

	// Log the operation (non-blocking)
	go func() {
		log.Printf("Shortened URL: %s -> %s", req.URL, shortURL)
	}()

	// Return response
	resp := ShortenResponse{
		Code: code,
		URL:  shortURL,
	}
	sendJSONResponse(w, resp, http.StatusCreated)
}

// RedirectHandler handles GET /r/{code} requests
func (h *URLHandler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract code from path
	code := r.URL.Path[len("/r/"):]
	if code == "" {
		http.Error(w, "Code is required", http.StatusBadRequest)
		return
	}

	// Get URL from store
	url, err := h.store.Get(code)
	if err != nil {
		if err == store.ErrCodeNotFound {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log the redirection (non-blocking)
	go func() {
		logRedirect(r.Context(), code)
	}()

	// Redirect to URL
	http.Redirect(w, r, url, http.StatusFound)
}

// logRedirect logs a redirect event in JSON format
func logRedirect(ctx context.Context, code string) {
	// Create a JSON log entry
	logEntry := struct {
		Code      string    `json:"code"`
		Timestamp time.Time `json:"time"`
	}{
		Code:      code,
		Timestamp: time.Now(),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Error marshaling log entry: %v", err)
		return
	}

	// Log to stdout
	log.Println(string(jsonData))
}

// sendJSONResponse sends a JSON response
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// sendJSONError sends a JSON error response
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	resp := ErrorResponse{
		Error: message,
	}
	sendJSONResponse(w, resp, statusCode)
}
