#!/bin/bash

# Exit on error
set -e

# Build the frontend
cd frontend
npm install
npm run build

# Create static directory in backend if it doesn't exist
mkdir -p ../backend/static

# Copy the frontend build to the backend static directory
cp -r dist/* ../backend/static/

# Build the backend
cd ../backend
go build -o app

echo "Build completed successfully"
