package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Parse command line arguments
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	fmt.Printf("Running goose %s...\n", command)

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("✓ Connected to PostgreSQL")

	// Set migration directory
	migrationsDir := "migrations"

	// Run goose command
	if err := goose.Run(command, db, migrationsDir); err != nil {
		log.Fatalf("Goose %s failed: %v", command, err)
	}

	fmt.Printf("✓ Goose %s completed successfully!\n", command)
}
