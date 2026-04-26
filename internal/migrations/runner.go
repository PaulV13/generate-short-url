package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run(dbURL, dir, direction string) error {
	if strings.TrimSpace(dbURL) == "" {
		return fmt.Errorf("database connection string is required")
	}

	if strings.TrimSpace(dir) == "" {
		dir = "migrations"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return err
	}

	files, err := migrationFiles(dir, direction)
	if err != nil {
		return err
	}

	for _, file := range files {
		content, readErr := os.ReadFile(file)
		if readErr != nil {
			return readErr
		}

		if _, execErr := db.Exec(string(content)); execErr != nil {
			return fmt.Errorf("migration failed on %s: %w", filepath.Base(file), execErr)
		}
	}

	return nil
}

func migrationFiles(dir, direction string) ([]string, error) {
	all, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(all))
	for _, file := range all {
		isDown := strings.HasSuffix(file, "_down.sql")
		if direction == "up" && !isDown {
			files = append(files, file)
		}
		if direction == "down" && isDown {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no migration files found for direction %q in %s", direction, dir)
	}

	sort.Strings(files)
	if direction == "down" {
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}
	}

	return files, nil
}
