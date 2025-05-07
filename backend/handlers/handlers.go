package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/priyankeshh/url-shortener/backend/store"
	"github.com/priyankeshh/url-shortener/backend/workers"
)

// URLHandler handles URL shortening and redirection
type URLHandler struct {
	store        store.URLStore
	host         string
	urlProcessor *workers.URLProcessor
}

// ShortenRequest represents the request body for shortening a URL
type ShortenRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"` // Optional custom alias
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

// SetURLProcessor sets the URL processor for the handler
func (h *URLHandler) SetURLProcessor(processor *workers.URLProcessor) {
	h.urlProcessor = processor
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

	// Get or create user ID from cookie
	userID := getUserID(w, r)

	// Store URL with options and get code
	code, err := h.store.SetWithOptions(req.URL, req.Alias, userID)
	if err != nil {
		// Handle specific errors
		switch err {
		case store.ErrAliasInUse:
			sendJSONError(w, "Custom alias is already in use", http.StatusConflict)
		case store.ErrInvalidAlias:
			sendJSONError(w, "Invalid alias: must be 3-20 alphanumeric characters", http.StatusBadRequest)
		default:
			sendJSONError(w, fmt.Sprintf("Failed to shorten URL: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Create short URL
	shortURL := fmt.Sprintf("%s/r/%s", h.host, code)

	// Log the operation (non-blocking)
	go func() {
		log.Printf("Shortened URL: %s -> %s (user: %s)", req.URL, shortURL, userID)
	}()

	// Submit URL for processing in the background if processor is available
	if h.urlProcessor != nil {
		h.urlProcessor.ProcessURL(req.URL)
	}

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

// getUserID gets the user ID from a cookie or creates a new one
func getUserID(w http.ResponseWriter, r *http.Request) string {
	// Try to get the user ID from the cookie
	cookie, err := r.Cookie("user_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Generate a new user ID
	userID := uuid.NewString()

	// Set the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 365, // 1 year
		SameSite: http.SameSiteLaxMode,
	})

	return userID
}

// GetUserURLsHandler handles GET /api/urls requests
func (h *URLHandler) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from cookie
	userID := getUserID(w, r)

	// Get URLs for the user
	entries, err := h.store.GetByUser(userID)
	if err != nil {
		sendJSONError(w, "Failed to get URLs", http.StatusInternalServerError)
		return
	}

	// Transform entries to include full URLs
	type UserURL struct {
		Code        string    `json:"code"`
		ShortURL    string    `json:"short_url"`
		OriginalURL string    `json:"original_url"`
		CreatedAt   time.Time `json:"created_at"`
	}

	userURLs := make([]UserURL, 0, len(entries))
	for _, entry := range entries {
		userURLs = append(userURLs, UserURL{
			Code:        entry.Code,
			ShortURL:    fmt.Sprintf("%s/r/%s", h.host, entry.Code),
			OriginalURL: entry.URL,
			CreatedAt:   entry.CreatedAt,
		})
	}

	// Return response
	sendJSONResponse(w, userURLs, http.StatusOK)
}
