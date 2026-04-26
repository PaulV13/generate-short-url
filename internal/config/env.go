package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var requiredEnvs = []string{
	"PORT",
	"BASE_URL",
	"CODE_LENGTH",
}

var requiredDatabaseEnvs = []string{
	"DATABASE_HOST",
	"DATABASE_PORT",
	"DATABASE_USERNAME",
	"DATABASE_PASSWORD",
	"DATABASE_NAME",
}

func ValidateRequiredEnv() error {
	missing := make([]string, 0)

	for _, key := range requiredEnvs {
		value, exists := os.LookupEnv(key)
		if !exists || strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}

	if strings.TrimSpace(os.Getenv("DATABASE_URL")) == "" {
		for _, key := range requiredDatabaseEnvs {
			value, exists := os.LookupEnv(key)
			if !exists || strings.TrimSpace(value) == "" {
				missing = append(missing, key)
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if strings.TrimSpace(os.Getenv("DATABASE_URL")) == "" {
		if _, err := strconv.Atoi(os.Getenv("DATABASE_PORT")); err != nil {
			return fmt.Errorf("DATABASE_PORT must be a valid integer")
		}
	}

	codeLength, err := strconv.Atoi(os.Getenv("CODE_LENGTH"))
	if err != nil || codeLength <= 0 {
		return fmt.Errorf("CODE_LENGTH must be a positive integer")
	}

	return nil
}
