# URL Shortener Frontend

This is the frontend for the URL shortener application.

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

To run the frontend locally:

1. Make sure you have Node.js installed (version 16 or later)

2. Install dependencies:
   ```bash
   npm install
   ```

3. Run the development server:
   ```bash
   npm run dev
   ```

4. The frontend will be available at http://localhost:3000

## Building for Production

To build the frontend for production:

1. Install dependencies:
   ```bash
   npm install
   ```

2. Build the application:
   ```bash
   npm run build
   ```

3. The built files will be in the `dist` directory
