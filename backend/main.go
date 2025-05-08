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
	port := flag.Int("port", 8080, "HTTP server port")
	host := flag.String("host", "http://localhost:8080", "Host for generated URLs")
	dbURL := flag.String("db-url", "", "PostgreSQL connection URL")
	workerCount := flag.Int("workers", 4, "Number of URL processor workers")
	flag.Parse()

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

	var urlStore store.URLStore
	connectionURL := *dbURL

	if envDBURL := os.Getenv("DATABASE_URL"); envDBURL != "" {
		connectionURL = envDBURL
		log.Printf("Found DATABASE_URL environment variable")
	} else {
		log.Printf("DATABASE_URL environment variable not found")
	}

	if connectionURL != "" {
		log.Printf("Attempting to connect to PostgreSQL database with connection string: %s", connectionURL)

		postgresStore, err := store.NewPostgresURLStore(connectionURL)
		if err != nil {
			log.Printf("Failed to create PostgreSQL store: %v", err)
			log.Println("Falling back to in-memory store")
		} else {
			log.Printf("Successfully connected to PostgreSQL database")
			defer postgresStore.Close()
			urlStore = postgresStore
		}
	} else {
		log.Printf("No database connection URL provided")
	}

	if urlStore == nil {
		log.Println("Using in-memory URL store")
		urlStore = store.NewInMemoryURLStore()
	}

	log.Printf("Starting URL processor with %d workers", *workerCount)
	urlProcessor := workers.NewURLProcessor(*workerCount)
	defer urlProcessor.Stop()

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

	urlHandler := handlers.NewURLHandler(urlStore, *host)
	urlHandler.SetURLProcessor(urlProcessor)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/shorten", urlHandler.ShortenHandler)
	mux.HandleFunc("/api/urls", urlHandler.GetUserURLsHandler)
	mux.HandleFunc("/api/metrics", handlers.GetMetricsHandler)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/r/", func(w http.ResponseWriter, r *http.Request) {
		code := strings.TrimPrefix(r.URL.Path, "/r/")
		if _, err := urlStore.Get(code); err == store.ErrCodeNotFound {
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

		urlHandler.RedirectHandler(w, r)
	})

	mux.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		openAPISpec, err := docsFS.ReadFile("docs/openapi.yaml")
		if err != nil {
			http.Error(w, "Documentation not available", http.StatusInternalServerError)
			return
		}
		w.Write(openAPISpec)
	})

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/r/") {
			return
		}

		path := "static" + r.URL.Path
		_, err := os.Stat(path)

		if os.IsNotExist(err) {
			http.ServeFile(w, r, "static/index.html")
			return
		}

		fs.ServeHTTP(w, r)
	}))

	handler := handlers.MetricsMiddleware(handlers.LoggingMiddleware(handlers.CORSMiddleware(mux)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on %s", server.Addr)
		log.Printf("API documentation available at %s/api/docs", *host)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
