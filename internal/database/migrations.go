package database

import (
	"log"

	"maxunlimited.com/swift_processor/internal/config"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrateUpAll(config *config.Config) {
	migrationsPath := "file://internal/database/migrations"

	m, err := migrate.New(
		migrationsPath,
		config.PostgresConnectionString,
	)

	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	log.Println("Attempting to run migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		// ErrNoChange means there were no pending migrations
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No migrations to apply.")
	} else {
		log.Println("Migrations applied successfully.")
	}
}

func MigrateDownAll(config *config.Config) {
	migrationsPath := "file://internal/database/migrations"

	m, err := migrate.New(
		migrationsPath,
		config.PostgresConnectionString,
	)

	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	log.Println("Attempting to run migrations...")
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		// ErrNoChange means there were no pending migrations
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No migrations to apply.")
	} else {
		log.Println("Migrations applied successfully.")
	}
}
