package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RequestMetrics stores metrics about API requests
type RequestMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AverageLatency     time.Duration
	PathCounts         map[string]int64
	mutex              sync.RWMutex
}

// Global metrics instance
var metrics = &RequestMetrics{
	PathCounts: make(map[string]int64),
}

// contextKey is a type for context keys
type contextKey string

const (
	// RequestIDKey is the context key for request IDs
	RequestIDKey = contextKey("request_id")
	// StartTimeKey is the context key for request start times
	StartTimeKey = contextKey("start_time")
)

// MetricsMiddleware collects metrics about requests
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a unique request ID
		requestID := uuid.NewString()

		// Store the start time
		startTime := time.Now()

		// Create a custom response writer to capture the status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Add request ID and start time to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, StartTimeKey, startTime)

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Call the next handler with the updated context
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Calculate request duration
		duration := time.Since(startTime)

		// Update metrics
		metrics.mutex.Lock()
		metrics.TotalRequests++
		metrics.PathCounts[r.URL.Path]++

		if rw.statusCode >= 400 {
			metrics.FailedRequests++
		} else {
			metrics.SuccessfulRequests++
		}

		// Update average latency
		if metrics.AverageLatency == 0 {
			metrics.AverageLatency = duration
		} else {
			// Simple moving average
			metrics.AverageLatency = (metrics.AverageLatency + duration) / 2
		}
		metrics.mutex.Unlock()

		// Log the request with its ID and duration
		log.Printf(
			"[%s] %s %s %d %s",
			requestID,
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
		)
	})
}

// GetMetricsHandler returns the current metrics
func GetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics.mutex.RLock()
	defer metrics.mutex.RUnlock()

	// Calculate success rate
	successRate := 0.0
	if metrics.TotalRequests > 0 {
		successRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100
	}

	// Format the response
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# URL Shortener Metrics\n\n")
	fmt.Fprintf(w, "Total Requests: %d\n", metrics.TotalRequests)
	fmt.Fprintf(w, "Successful Requests: %d\n", metrics.SuccessfulRequests)
	fmt.Fprintf(w, "Failed Requests: %d\n", metrics.FailedRequests)
	fmt.Fprintf(w, "Success Rate: %.2f%%\n", successRate)
	fmt.Fprintf(w, "Average Latency: %s\n\n", metrics.AverageLatency)

	fmt.Fprintf(w, "# Requests by Path\n\n")
	for path, count := range metrics.PathCounts {
		fmt.Fprintf(w, "%s: %d\n", path, count)
	}
}
