package db

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DatabaseConnection() *gorm.DB {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	sslMode := strings.TrimSpace(os.Getenv("DATABASE_SSLMODE"))
	if sslMode == "" {
		sslMode = "disable"
	}

	sqlInfo := databaseURL
	if sqlInfo == "" {
		port, err := strconv.Atoi(os.Getenv("DATABASE_PORT"))
		if err != nil {
			log.Fatal("DATABASE_PORT must be a valid integer")
		}

		var (
			host     = os.Getenv("DATABASE_HOST")
			user     = os.Getenv("DATABASE_USERNAME")
			password = os.Getenv("DATABASE_PASSWORD")
			dbName   = os.Getenv("DATABASE_NAME")
		)

		sqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslMode)
	}

	db, err := gorm.Open(postgres.Open(sqlInfo), &gorm.Config{})
	if err != nil {
		log.Fatal("error connecting to database")
	}

	return db
}

func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}
