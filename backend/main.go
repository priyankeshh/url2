package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/priyankeshh/url-shortener/backend/handlers"
	"github.com/priyankeshh/url-shortener/backend/store"
	"github.com/priyankeshh/url-shortener/backend/workers"
)

//go:embed docs/openapi.yaml
//go:embed docs/404.html
var docsFS embed.FS

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "HTTP server port")
	host := flag.String("host", "http://localhost:8080", "Host for generated URLs")
	dbURL := flag.String("db-url", "", "PostgreSQL connection URL")
	workerCount := flag.Int("workers", 4, "Number of URL processor workers")
	flag.Parse()

	// Override with environment variables if set
	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", port)
	}
	if envHost := os.Getenv("HOST"); envHost != "" {
		*host = envHost
	}
	if envDB := os.Getenv("DATABASE_URL"); envDB != "" {
		*dbURL = envDB
	}
	if envWorkers := os.Getenv("WORKER_COUNT"); envWorkers != "" {
		fmt.Sscanf(envWorkers, "%d", workerCount)
	}

	// Create URL store
	var urlStore store.URLStore

	// Try to use PostgreSQL if connection URL is provided
	connectionURL := *dbURL

	// Check for DATABASE_URL environment variable (used by many hosting platforms)
	if envDBURL := os.Getenv("DATABASE_URL"); envDBURL != "" {
		connectionURL = envDBURL
	}

	if connectionURL != "" {
		log.Printf("Using PostgreSQL database")

		postgresStore, err := store.NewPostgresURLStore(connectionURL)
		if err != nil {
			log.Printf("Failed to create PostgreSQL store: %v", err)
			log.Println("Falling back to in-memory store")
		} else {
			defer postgresStore.Close()
			urlStore = postgresStore
		}
	}

	// Use in-memory store if PostgreSQL is not available
	if urlStore == nil {
		log.Println("Using in-memory URL store")
		urlStore = store.NewInMemoryURLStore()
	}

	// Create URL processor
	log.Printf("Starting URL processor with %d workers", *workerCount)
	urlProcessor := workers.NewURLProcessor(*workerCount)
	defer urlProcessor.Stop()

	// Start a goroutine to process URL validation results
	go func() {
		for result := range urlProcessor.GetResults() {
			if result.Error != nil {
				log.Printf("URL processing error for %s: %v", result.URL, result.Error)
			} else {
				log.Printf("URL %s processed: status=%d, content-type=%s, time=%s",
					result.URL, result.StatusCode, result.ContentType, result.ProcessTime)
			}
		}
	}()

	// Create URL handler
	urlHandler := handlers.NewURLHandler(urlStore, *host)

	// Set URL processor for the handler
	urlHandler.SetURLProcessor(urlProcessor)

	// Create router
	mux := http.NewServeMux()

	// Register API endpoints
	mux.HandleFunc("/api/shorten", urlHandler.ShortenHandler)
	mux.HandleFunc("/api/urls", urlHandler.GetUserURLsHandler)
	mux.HandleFunc("/api/metrics", handlers.GetMetricsHandler)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
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

	// Serve static files from the static directory
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip API and redirect paths
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/r/") {
			return
		}

		// Check if the file exists in the static directory
		path := "static" + r.URL.Path
		_, err := os.Stat(path)

		// If the file doesn't exist, serve the index.html file
		if os.IsNotExist(err) {
			http.ServeFile(w, r, "static/index.html")
			return
		}

		// Otherwise, serve the file
		fs.ServeHTTP(w, r)
	}))

	// Add middleware (metrics, logging, CORS)
	handler := handlers.MetricsMiddleware(handlers.LoggingMiddleware(handlers.CORSMiddleware(mux)))

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", server.Addr)
		log.Printf("API documentation available at %s/api/docs", *host)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
