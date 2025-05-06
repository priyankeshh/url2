.PHONY: build run test clean docker-build docker-run

# Go backend
build:
	cd backend && go build -o url-shortener

run:
	cd backend && go run main.go

test:
	cd backend && go test ./...

clean:
	rm -f backend/url-shortener

# Docker
docker-build-backend:
	docker build -t url-shortener-backend ./backend

docker-build-frontend:
	docker build -t url-shortener-frontend ./frontend

docker-run-backend:
	docker run -p 8080:8080 url-shortener-backend

docker-run-frontend:
	docker run -p 3000:80 url-shortener-frontend

# Docker Compose
compose-up:
	docker-compose up --build

compose-down:
	docker-compose down

# Frontend
frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

# Combined commands
dev:
	@echo "Starting backend server..."
	@cd backend && go run main.go & \
	echo "Starting frontend server..." && \
	cd frontend && npm run dev

# Help
help:
	@echo "Available commands:"
	@echo "  make build                - Build the Go backend"
	@echo "  make run                  - Run the Go backend"
	@echo "  make test                 - Run tests for the Go backend"
	@echo "  make clean                - Remove build artifacts"
	@echo "  make docker-build-backend - Build Docker image for the backend"
	@echo "  make docker-build-frontend - Build Docker image for the frontend"
	@echo "  make docker-run-backend   - Run Docker container for the backend"
	@echo "  make docker-run-frontend  - Run Docker container for the frontend"
	@echo "  make compose-up           - Start all services with Docker Compose"
	@echo "  make compose-down         - Stop all services with Docker Compose"
	@echo "  make frontend-install     - Install frontend dependencies"
	@echo "  make frontend-dev         - Start frontend development server"
	@echo "  make frontend-build       - Build frontend for production"
	@echo "  make dev                  - Start both backend and frontend for development"
