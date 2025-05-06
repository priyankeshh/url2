# URL Shortener Backend

This is the backend for the URL shortener application.

## Running with Docker Compose

To run the entire application (frontend and backend) with Docker Compose:

1. Navigate to the root directory of the project:
   ```bash
   cd ..
   ```

2. Run Docker Compose:
   ```bash
   docker-compose up --build
   ```

3. Access the application at http://localhost:3000

## Running Locally

To run the backend locally:

1. Make sure you have Go installed (version 1.21 or later)

2. Run the backend:
   ```bash
   go run main.go
   ```

3. The backend will be available at http://localhost:8080

## API Endpoints

- `POST /api/shorten` - Shorten a URL
- `GET /r/{code}` - Redirect to the original URL
- `GET /api/docs` - View API documentation
