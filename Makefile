.PHONY: build run test clean docker-build docker-run

# Go backend
build:
	go build -o url-shortener

run:
	go run main.go

test:
	go test ./...

clean:
	rm -f url-shortener

# Docker
docker-build:
	docker build -t url-shortener .

docker-run:
	docker run -p 8080:8080 url-shortener

# Docker Compose
compose-up:
	docker-compose up --build

compose-down:
	docker-compose down

# Frontend
frontend-install:
	npm install

frontend-dev:
	npm run dev

frontend-build:
	npm run build

# Combined commands
dev: build
	@echo "Starting backend server..."
	@./url-shortener & \
	echo "Starting frontend server..." && \
	npm run dev

# Help
help:
	@echo "Available commands:"
	@echo "  make build            - Build the Go backend"
	@echo "  make run              - Run the Go backend"
	@echo "  make test             - Run tests for the Go backend"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make docker-build     - Build Docker image for the backend"
	@echo "  make docker-run       - Run Docker container for the backend"
	@echo "  make compose-up       - Start all services with Docker Compose"
	@echo "  make compose-down     - Stop all services with Docker Compose"
	@echo "  make frontend-install - Install frontend dependencies"
	@echo "  make frontend-dev     - Start frontend development server"
	@echo "  make frontend-build   - Build frontend for production"
	@echo "  make dev              - Start both backend and frontend for development"
