package main

import (
	"log"
	"os"
	"strings"

	"generate-short-url/internal/db"
	"generate-short-url/internal/migrations"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [up|down]")
	}

	direction := strings.TrimSpace(strings.ToLower(os.Args[1]))
	if direction != "up" && direction != "down" {
		log.Fatal("usage: migrate [up|down]")
	}

	connectionString, err := db.ConnectionStringFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	migrationsDir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	if err := migrations.Run(connectionString, migrationsDir, direction); err != nil {
		log.Fatal(err)
	}

	log.Printf("migrations %s completed", direction)
}
