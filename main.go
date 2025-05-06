package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"url-shortener/handlers"
	"url-shortener/store"
)

//go:embed docs/openapi.yaml
//go:embed docs/404.html
var docsFS embed.FS

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "HTTP server port")
	host := flag.String("host", "http://localhost:8080", "Host for generated URLs")
	flag.Parse()

	// Override with environment variables if set
	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", port)
	}
	if envHost := os.Getenv("HOST"); envHost != "" {
		*host = envHost
	}

	// Create URL store
	urlStore := store.NewInMemoryURLStore()

	// Create URL handler
	urlHandler := handlers.NewURLHandler(urlStore, *host)

	// Create router
	mux := http.NewServeMux()

	// Register API endpoints
	mux.HandleFunc("/api/shorten", urlHandler.ShortenHandler)
	mux.HandleFunc("/r/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the URL exists
		code := strings.TrimPrefix(r.URL.Path, "/r/")
		if _, err := urlStore.Get(code); err == store.ErrCodeNotFound {
			// Serve custom 404 page
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			notFoundHTML, err := docsFS.ReadFile("docs/404.html")
			if err != nil {
				http.Error(w, "URL not found", http.StatusNotFound)
				return
			}
			w.Write(notFoundHTML)
			return
		}

		// Handle the redirect
		urlHandler.RedirectHandler(w, r)
	})

	// Serve OpenAPI documentation
	mux.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		openAPISpec, err := docsFS.ReadFile("docs/openapi.yaml")
		if err != nil {
			http.Error(w, "Documentation not available", http.StatusInternalServerError)
			return
		}
		w.Write(openAPISpec)
	})

	// Add middleware
	handler := handlers.LoggingMiddleware(handlers.CORSMiddleware(mux))

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)
	log.Printf("API documentation available at %s/api/docs", *host)
	log.Fatal(http.ListenAndServe(addr, handler))
}
