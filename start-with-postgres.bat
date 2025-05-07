@echo off
echo Starting URL Shortener with PostgreSQL database...
cd backend
set DATABASE_URL=postgresql://neondb_owner:npg_N4umKc3zEyGW@ep-winter-sound-a10irlyk-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require
go run main.go
