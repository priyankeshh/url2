package workers

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type URLProcessResult struct {
	URL         string
	StatusCode  int
	Title       string
	ContentType string
	Error       error
	ProcessTime time.Duration
}

type URLProcessor struct {
	workerCount int
	client      *http.Client
	jobs        chan string
	results     chan URLProcessResult
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

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

	processor.startWorkers()

	return processor
}

func (p *URLProcessor) startWorkers() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	go func() {
		p.wg.Wait()
		close(p.results)
	}()
}

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

			result := p.processURL(urlString)

			select {
			case p.results <- result:
			case <-p.ctx.Done():
				log.Printf("URL processor worker %d stopping due to context cancellation", id)
				return
			}
		}
	}
}

func (p *URLProcessor) processURL(urlString string) URLProcessResult {
	startTime := time.Now()
	result := URLProcessResult{
		URL: urlString,
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		result.Error = err
		result.ProcessTime = time.Since(startTime)
		return result
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
		urlString = parsedURL.String()
	}

	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, urlString, nil)
	if err != nil {
		result.Error = err
		result.ProcessTime = time.Since(startTime)
		return result
	}

	req.Header.Set("User-Agent", "URLShortener/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		result.Error = err
		result.ProcessTime = time.Since(startTime)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.ContentType = resp.Header.Get("Content-Type")
	result.ProcessTime = time.Since(startTime)

	return result
}

func (p *URLProcessor) ProcessURL(urlString string) {
	select {
	case p.jobs <- urlString:
	case <-p.ctx.Done():
	}
}

func (p *URLProcessor) GetResults() <-chan URLProcessResult {
	return p.results
}

func (p *URLProcessor) Stop() {
	p.cancel()
	close(p.jobs)
}
