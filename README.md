# URL Shortener

A modern URL shortener application with a React frontend and Go backend.

## Features

- Shorten long URLs to easy-to-share links
- Copy shortened URLs to clipboard with a single click
- View history of previously shortened URLs
- Responsive design that works on desktop and mobile
- Fast and lightweight Go backend

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
- In-memory URL storage with thread-safety

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

- `POST /api/shorten` - Shorten a URL
- `GET /r/{code}` - Redirect to the original URL

## License

This project is licensed under the MIT License - see the LICENSE file for details.
