package configs

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func Connect(config *Config) (*sql.DB, error) {
	db, err := sql.Open(config.Database.Driver, config.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	// Read schema file.
	// Assuming the application is run from the root directory.
	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute schema.
	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

func Seed(db *sql.DB) error {
	// Check if catalog table is empty.
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM catalog").Scan(&count); err != nil {
		return fmt.Errorf("failed to check catalog table: %w", err)
	}

	if count > 0 {
		return nil // Already seeded
	}

	// Read seed file.
	seed, err := os.ReadFile("database/seed.sql")
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	// Execute seed.
	if _, err := db.Exec(string(seed)); err != nil {
		return fmt.Errorf("failed to execute seed: %w", err)
	}

	return nil
}
