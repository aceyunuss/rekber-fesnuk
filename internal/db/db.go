package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

func Connect(dbURL string) (*sql.DB, error) {

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection: %w", err)
	}

	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("Failed ping database: %w", err)
	}

	return conn, nil
}
