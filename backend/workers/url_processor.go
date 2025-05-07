package workers

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// URLProcessResult represents the result of processing a URL
type URLProcessResult struct {
	URL         string
	StatusCode  int
	Title       string
	ContentType string
	Error       error
	ProcessTime time.Duration
}

// URLProcessor handles concurrent URL processing
type URLProcessor struct {
	workerCount int
	client      *http.Client
	jobs        chan string
	results     chan URLProcessResult
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewURLProcessor creates a new URL processor with the specified number of workers
func NewURLProcessor(workerCount int) *URLProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	
	processor := &URLProcessor{
		workerCount: workerCount,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		jobs:    make(chan string, workerCount*2),
		results: make(chan URLProcessResult, workerCount*2),
		ctx:     ctx,
		cancel:  cancel,
	}
	
	// Start workers
	processor.startWorkers()
	
	return processor
}

// startWorkers starts the worker goroutines
func (p *URLProcessor) startWorkers() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	
	// Start a goroutine to close the results channel when all workers are done
	go func() {
		p.wg.Wait()
		close(p.results)
	}()
}

// worker processes URLs from the jobs channel
func (p *URLProcessor) worker(id int) {
	defer p.wg.Done()
	
	log.Printf("URL processor worker %d started", id)
	
	for {
		select {
		case <-p.ctx.Done():
			log.Printf("URL processor worker %d stopping due to context cancellation", id)
			return
		case urlString, ok := <-p.jobs:
			if !ok {
				log.Printf("URL processor worker %d stopping due to closed jobs channel", id)
				return
			}
			
			// Process the URL
			result := p.processURL(urlString)
			
			// Send the result
			select {
			case p.results <- result:
				// Result sent successfully
			case <-p.ctx.Done():
				log.Printf("URL processor worker %d stopping due to context cancellation", id)
				return
			}
		}
	}
}

// processURL validates and processes a URL
func (p *URLProcessor) processURL(urlString string) URLProcessResult {
	startTime := time.Now()
	result := URLProcessResult{
		URL: urlString,
	}
	
	// Parse the URL
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		result.Error = err
		result.ProcessTime = time.Since(startTime)
		return result
	}
	
	// Ensure the URL has a scheme
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
		urlString = parsedURL.String()
	}
	
	// Create a context with timeout for the request
	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()
	
	// Create a request
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, urlString, nil)
	if err != nil {
		result.Error = err
		result.ProcessTime = time.Since(startTime)
		return result
	}
	
	// Set a user agent
	req.Header.Set("User-Agent", "URLShortener/1.0")
	
	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		result.Error = err
		result.ProcessTime = time.Since(startTime)
		return result
	}
	defer resp.Body.Close()
	
	// Record the status code and content type
	result.StatusCode = resp.StatusCode
	result.ContentType = resp.Header.Get("Content-Type")
	
	// Record the processing time
	result.ProcessTime = time.Since(startTime)
	
	return result
}

// ProcessURL submits a URL for processing
func (p *URLProcessor) ProcessURL(urlString string) {
	select {
	case p.jobs <- urlString:
		// URL submitted successfully
	case <-p.ctx.Done():
		// Context cancelled
	}
}

// GetResults returns the results channel
func (p *URLProcessor) GetResults() <-chan URLProcessResult {
	return p.results
}

// Stop stops the URL processor
func (p *URLProcessor) Stop() {
	p.cancel()
	close(p.jobs)
}
