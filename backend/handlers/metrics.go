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

type RequestMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AverageLatency     time.Duration
	PathCounts         map[string]int64
	mutex              sync.RWMutex
}

var metrics = &RequestMetrics{
	PathCounts: make(map[string]int64),
}

type contextKey string

const (
	RequestIDKey = contextKey("request_id")
	StartTimeKey = contextKey("start_time")
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		startTime := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, StartTimeKey, startTime)

		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(rw, r.WithContext(ctx))

		duration := time.Since(startTime)

		metrics.mutex.Lock()
		metrics.TotalRequests++
		metrics.PathCounts[r.URL.Path]++

		if rw.statusCode >= 400 {
			metrics.FailedRequests++
		} else {
			metrics.SuccessfulRequests++
		}

		if metrics.AverageLatency == 0 {
			metrics.AverageLatency = duration
		} else {
			metrics.AverageLatency = (metrics.AverageLatency + duration) / 2
		}
		metrics.mutex.Unlock()

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

func GetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics.mutex.RLock()
	defer metrics.mutex.RUnlock()

	successRate := 0.0
	if metrics.TotalRequests > 0 {
		successRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100
	}

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
