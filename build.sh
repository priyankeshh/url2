#!/bin/bash

# Build the frontend
cd frontend
npm install
npm run build

# Create static directory in backend if it doesn't exist
mkdir -p ../backend/static

# Copy the frontend build to the backend static directory
cp -r dist/* ../backend/static/

echo "Frontend built and copied to backend/static"
