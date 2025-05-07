package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// loadEnv loads environment variables from a .env file
func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		// Split by the first equals sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Set environment variable
		os.Setenv(key, value)
	}

	return scanner.Err()
}

func main() {
	// Load environment variables from .env file
	if err := loadEnv(".env"); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	// Get the database URL from the environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	fmt.Printf("Attempting to connect to PostgreSQL database with connection string: %s\n", dbURL)

	// Open PostgreSQL database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL database!")

	// Check if the urls table exists
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'urls'
		)
	`).Scan(&tableExists)
	if err != nil {
		log.Fatalf("Failed to check if table exists: %v", err)
	}

	if tableExists {
		fmt.Println("The 'urls' table exists in the database.")

		// Count the number of URLs in the database
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count)
		if err != nil {
			log.Fatalf("Failed to count URLs: %v", err)
		}
		fmt.Printf("There are %d URLs in the database.\n", count)
	} else {
		fmt.Println("The 'urls' table does not exist in the database.")
	}
}
