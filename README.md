# Go URL Shortener

A feature-rich URL shortener built with Go, showcasing Go's powerful features and concurrency patterns.

## Features

- **Custom URL Aliases**: Create memorable short URLs with custom aliases
- **Persistent Storage**: SQLite database ensures URLs persist across server restarts
- **User-specific History**: Browser-based tracking of shortened URLs
- **Concurrent URL Processing**: Background validation and analysis of URLs
- **Metrics and Monitoring**: Real-time statistics on API usage

## Project Structure

```
url-shortener/
├── backend/                # Go backend
│   ├── docs/               # OpenAPI spec and static assets
│   ├── handlers/           # HTTP handlers
│   ├── store/              # URL store implementation
│   ├── Dockerfile          # Backend Docker configuration
│   └── main.go             # Entry point
├── frontend/               # React frontend
│   ├── src/                # Source code
│   │   ├── components/     # React components
│   │   ├── hooks/          # Custom React hooks
│   │   ├── utils/          # Utility functions
│   │   └── types/          # TypeScript type definitions
│   ├── Dockerfile          # Frontend Docker configuration
│   └── nginx.conf          # Nginx configuration for production
├── docker-compose.yml      # Docker Compose configuration
└── Makefile                # Build and run commands
```

## Tech Stack

### Frontend
- React with TypeScript
- Tailwind CSS for styling
- Vite for fast development and building

### Backend
- Go (Golang)
- Standard library HTTP server
- SQLite database for persistent storage
- Goroutines and channels for concurrency
- Context for request cancellation and timeouts

## Getting Started

### Prerequisites

- Node.js (v16+)
- Go (v1.18+)
- Docker and Docker Compose (optional, for containerized deployment)

### Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/url-shortener.git
   cd url-shortener
   ```

2. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   ```

3. Start the Go backend:
   ```bash
   cd backend
   go run main.go
   ```

4. In a separate terminal, start the frontend development server:
   ```bash
   cd frontend
   npm run dev
   ```

5. Open your browser and navigate to http://localhost:3000

### Using Make

The project includes a Makefile with common commands:

```bash
# Start both backend and frontend for development
make dev

# Run tests
make test

# Build the backend
make build

# See all available commands
make help
```

## Deployment

### Using Docker Compose

1. Build and start the containers:
   ```bash
   docker-compose up --build
   ```

2. Access the application at http://localhost:3000

### Manual Deployment

#### Backend

1. Build the Go binary:
   ```bash
   cd backend
   go build -o url-shortener
   ```

2. Run the binary:
   ```bash
   ./url-shortener
   ```

#### Frontend

1. Build the frontend:
   ```bash
   cd frontend
   npm run build
   ```

2. Serve the static files from the `frontend/dist` directory using a web server like Nginx.

## Configuration

### Environment Variables

Create a `.env` file based on `.env.example`:

```
# Frontend environment variables
VITE_API_URL=http://localhost:8080

# Backend environment variables
PORT=8080
HOST=http://localhost:8080
```

## API Documentation

The API documentation is available at `/api/docs` when the server is running.

### Endpoints

- `POST /api/shorten` - Shorten a URL (with optional custom alias)
- `GET /api/urls` - Get user's shortened URLs
- `GET /r/{code}` - Redirect to the original URL
- `GET /api/metrics` - View server metrics
- `GET /api/docs` - API documentation

## Go-Specific Features

This project showcases many Go-specific features and patterns:

### 1. Concurrency with Goroutines and Channels

- **Worker Pool Pattern**: The URL processor uses a pool of goroutines to process URLs concurrently
- **Fan-out/Fan-in**: Results from multiple workers are collected in a single channel
- **Context for Cancellation**: Graceful shutdown of workers using context
- **Synchronization with WaitGroups**: Coordinating multiple goroutines

```go
// Example of worker pool pattern
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
```

### 2. Middleware Chaining

- **HTTP Handler Wrapping**: Middleware functions that wrap handlers
- **Context Propagation**: Request-scoped values passed through context
- **Functional Composition**: Combining multiple middleware functions

```go
// Example of middleware chaining
handler := handlers.MetricsMiddleware(
    handlers.LoggingMiddleware(
        handlers.CORSMiddleware(mux)
    )
)
```

### 3. Interface-based Design

- **Store Interface**: Abstraction for different storage backends
- **Dependency Injection**: Components receive their dependencies
- **Testability**: Easy to mock dependencies for testing

```go
// Example of interface-based design
type URLStore interface {
    Set(url string) (string, error)
    SetWithOptions(url, customAlias, userID string) (string, error)
    Get(code string) (string, error)
    GetByUser(userID string) ([]URLEntry, error)
    Stats() int
}
```

### 4. Go's Standard Library

- **net/http**: Built-in HTTP server with no external dependencies
- **context**: Request cancellation and value propagation
- **sync**: Mutex and WaitGroup for synchronization
- **embed**: Embedding static files in the binary
- **database/sql**: Database-agnostic SQL interface

### 5. Error Handling

- **Error as Values**: Explicit error handling
- **Custom Error Types**: Domain-specific errors
- **Error Wrapping**: Preserving error context

```go
// Example of custom errors
var (
    ErrCodeNotFound = errors.New("code not found")
    ErrInvalidURL   = errors.New("invalid URL")
    ErrAliasInUse   = errors.New("custom alias is already in use")
)
```

### 6. Graceful Shutdown

- **Signal Handling**: Catching OS signals for graceful shutdown
- **Context Timeout**: Setting deadlines for shutdown operations
- **Resource Cleanup**: Proper closing of database connections

```go
// Example of graceful shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
log.Println("Shutting down server...")

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    log.Fatalf("Server forced to shutdown: %v", err)
}
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
