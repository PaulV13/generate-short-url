package db

import (
	"fmt"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DatabaseConnection() *gorm.DB {
	sqlInfo, err := ConnectionStringFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open(sqlInfo), &gorm.Config{})
	if err != nil {
		log.Fatal("error connecting to database")
	}

	return db
}

func ConnectionStringFromEnv() (string, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL != "" {
		return databaseURL, nil
	}

	sslMode := strings.TrimSpace(os.Getenv("DATABASE_SSLMODE"))
	if sslMode == "" {
		sslMode = "disable"
	}

	portValue := strings.TrimSpace(os.Getenv("DATABASE_PORT"))
	if portValue == "" {
		return "", errors.New("DATABASE_PORT is required when DATABASE_URL is not set")
	}

	port, err := strconv.Atoi(portValue)
	if err != nil {
		return "", errors.New("DATABASE_PORT must be a valid integer")
	}

	host := strings.TrimSpace(os.Getenv("DATABASE_HOST"))
	user := strings.TrimSpace(os.Getenv("DATABASE_USERNAME"))
	password := os.Getenv("DATABASE_PASSWORD")
	dbName := strings.TrimSpace(os.Getenv("DATABASE_NAME"))

	if host == "" || user == "" || dbName == "" {
		return "", errors.New("DATABASE_HOST, DATABASE_USERNAME and DATABASE_NAME are required when DATABASE_URL is not set")
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		user,
		password,
		dbName,
		sslMode,
	), nil
}

func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}
