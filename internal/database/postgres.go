package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"transjakarta-fleet/internal/config"
)

func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}

func RunMigrations(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS vehicle_locations (
		id SERIAL PRIMARY KEY,
		vehicle_id VARCHAR(50) NOT NULL,
		latitude DOUBLE PRECISION NOT NULL,
		longitude DOUBLE PRECISION NOT NULL,
		timestamp BIGINT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_vehicle_id ON vehicle_locations(vehicle_id);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON vehicle_locations(timestamp);
	CREATE INDEX IF NOT EXISTS idx_vehicle_timestamp ON vehicle_locations(vehicle_id, timestamp DESC);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}
