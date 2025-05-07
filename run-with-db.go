package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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
		fmt.Printf("Set environment variable: %s\n", key)
	}

	return scanner.Err()
}

func main() {
	// Load environment variables from .env file
	if err := loadEnv(".env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// Check if DATABASE_URL is set
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	fmt.Printf("DATABASE_URL is set to: %s\n", dbURL)

	// Run the application
	cmd := exec.Command("go", "run", "backend/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	fmt.Println("Starting the application...")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run the application: %v", err)
	}
}
