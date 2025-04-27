package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type IDB interface {
	Close()
}

// DB struct holds the database connection pool
type DB struct {
	SQL *sql.DB
}

// NewDB initializes and returns a new DB instance with the connection pool
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// Ping the database to verify the connection is established
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{SQL: db}, nil
}

// Close closes the database connection pool
func (d *DB) Close() {
	if d.SQL != nil {
		d.SQL.Close()
	}
}

// EnsureDatabaseExists connects to a default database (like 'postgres')
// and executes CREATE DATABASE IF NOT EXISTS for the target database.
// connectionString must be a URL that can connect to the server with create privileges,
// the database name part of this string will be ignored and replaced,
// typically connect to the 'postgres' or an empty database.
func EnsureDatabaseExists(dbName string, connectionString string) error {
	// Parse the connection string URL
	u, err := url.Parse(connectionString)
	if err != nil {
		return fmt.Errorf("invalid connection string URL: %w", err)
	}

	u.Path = "/postgres" // Connect to the default postgres database

	// Reconstruct the connection string for the default database
	connectURL := u.String()
	log.Printf("Attempting to connect to default DB with the `postgres` account")

	// Connect to the default database
	db, err := sql.Open("pgx", connectURL)
	if err != nil {
		return fmt.Errorf("unable to connect to default database (%s): %w", connectURL, err)
	}
	defer db.Close()

	// Ping to verify the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("ping to default database failed: %w", err)
	}
	log.Println("Successfully connected to default database.")

	createDBSQL := fmt.Sprintf("CREATE DATABASE %s", dbName)

	log.Printf("Executing: %s", createDBSQL)
	_, err = db.Exec(createDBSQL)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("Database '%s' already exists, skipping creation.", dbName)
			return nil // Database already exists, no error
		}
		return fmt.Errorf("failed to execute CREATE DATABASE: %w", err)
	}

	log.Printf("Database '%s' ensured to exist (created or already present).", dbName)

	return nil // Success
}
