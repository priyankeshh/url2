# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o url-shortener

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS and SQLite
RUN apk --no-cache add ca-certificates sqlite

# Set working directory
WORKDIR /app

# Create data directory
RUN mkdir -p /app/data

# Copy binary from builder
COPY --from=builder /app/url-shortener .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./url-shortener"]
