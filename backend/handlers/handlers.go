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

type URLHandler struct {
	store        store.URLStore
	host         string
	urlProcessor *workers.URLProcessor
}

type ShortenRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

type ShortenResponse struct {
	Code string `json:"code"`
	URL  string `json:"url,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewURLHandler(store store.URLStore, host string) *URLHandler {
	return &URLHandler{
		store: store,
		host:  host,
	}
}

func (h *URLHandler) SetURLProcessor(processor *workers.URLProcessor) {
	h.urlProcessor = processor
}

func (h *URLHandler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		sendJSONError(w, "URL is required", http.StatusBadRequest)
		return
	}

	userID := getUserID(w, r)

	code, err := h.store.SetWithOptions(req.URL, req.Alias, userID)
	if err != nil {
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

	shortURL := fmt.Sprintf("%s/r/%s", h.host, code)

	go func() {
		log.Printf("Shortened URL: %s -> %s (user: %s)", req.URL, shortURL, userID)
	}()

	if h.urlProcessor != nil {
		h.urlProcessor.ProcessURL(req.URL)
	}

	resp := ShortenResponse{
		Code: code,
		URL:  shortURL,
	}
	sendJSONResponse(w, resp, http.StatusCreated)
}

func (h *URLHandler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Path[len("/r/"):]
	if code == "" {
		http.Error(w, "Code is required", http.StatusBadRequest)
		return
	}

	url, err := h.store.Get(code)
	if err != nil {
		if err == store.ErrCodeNotFound {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	go func() {
		logRedirect(r.Context(), code)
	}()

	http.Redirect(w, r, url, http.StatusFound)
}

func logRedirect(_ context.Context, code string) {
	logEntry := struct {
		Code      string    `json:"code"`
		Timestamp time.Time `json:"time"`
	}{
		Code:      code,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Error marshaling log entry: %v", err)
		return
	}

	log.Println(string(jsonData))
}

func sendJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	resp := ErrorResponse{
		Error: message,
	}
	sendJSONResponse(w, resp, statusCode)
}

func getUserID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("user_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	userID := uuid.NewString()

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 365,
		SameSite: http.SameSiteLaxMode,
	})

	return userID
}

func (h *URLHandler) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(w, r)

	entries, err := h.store.GetByUser(userID)
	if err != nil {
		sendJSONError(w, "Failed to get URLs", http.StatusInternalServerError)
		return
	}

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

	sendJSONResponse(w, userURLs, http.StatusOK)
}
