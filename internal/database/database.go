package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"maxunlimited.com/swift_processor/internal/models"
)

type IDB interface {

	// Bank methods
	GetBankBySwiftCode(swiftCode string) (*models.BankHequarterDTO, error)
	GetBanksByCountry(countryISO2 string) ([]models.Bank, error)
	CreateBank(bank *models.BankBranchDTO) error
	DeleteBankBySwiftCode(swiftCode string) error
	GetCountryNameByISO2(countryISO2 string) (string, error)

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

// --- Database Operation Methods (repository layer) ---

// GetBankBySwiftCode retrieves a bank by its SWIFT code.
// Returns a Bank struct if found, or nil if not found.
// Returns an error if the query fails.
func (d *DB) GetBankBySwiftCode(swiftCode string) (*models.Bank, error) {
	query := `SELECT id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address
			  FROM banks WHERE swiftCode = $1`
	row := d.SQL.QueryRow(query, swiftCode)

	var bank models.Bank
	err := row.Scan(&bank.Id, &bank.BankName, &bank.CountryISO2, &bank.CountryName,
		&bank.IsHeadquarter, &bank.HeadquartersId, &bank.SwiftCode, &bank.Address)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No bank found with the given SWIFT code
		}
		return nil, fmt.Errorf("failed to retrieve bank by SWIFT code: %w", err)
	}

	// If the bank is a headquarter, retrieve its branches
	if bank.IsHeadquarter {
		query = `SELECT id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address
				 FROM banks WHERE headquartersId = $1`
		rows, err := d.SQL.Query(query, bank.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve branches for headquarter: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var branch models.Bank
			err := rows.Scan(&branch.Id, &branch.BankName, &branch.CountryISO2,
				&branch.CountryName, &branch.IsHeadquarter,
				&branch.HeadquartersId, &branch.SwiftCode,
				&branch.Address)
			if err != nil {
				return nil, fmt.Errorf("failed to scan branch row: %w", err)
			}
			bank.Branches = append(bank.Branches, branch)
		}

		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over branch rows: %w", err)
		}
	}
	return &bank, nil
}

// GetBanksByCountry retrieves all banks in a specific country by its ISO2 code.
// Returns a slice of Bank structs.
// Returns an error if the query fails.
func (d *DB) GetBanksByCountry(countryISO2 string) ([]models.Bank, error) {
	query := `SELECT id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address
				 FROM banks WHERE countryISO2 = $1`
	rows, err := d.SQL.Query(query, countryISO2)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve banks by country: %w", err)
	}
	defer rows.Close()

	var banks []models.Bank
	for rows.Next() {
		var bank models.Bank
		err := rows.Scan(&bank.Id, &bank.BankName, &bank.CountryISO2,
			&bank.CountryName, &bank.IsHeadquarter, &bank.HeadquartersId,
			&bank.SwiftCode, &bank.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bank row: %w", err)
		}
		banks = append(banks, bank)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over bank rows: %w", err)
	}

	return banks, nil
}

// CreateBank inserts a new bank into the database.
// Returns an error if the insertion fails.
// Returns nil if the bank was successfully created.
// The bank.Id field will be populated with the new bank's ID.
func (d *DB) CreateBank(bank *models.Bank) error {
	query := `INSERT INTO banks (id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	err := d.SQL.QueryRow(query,
		bank.BankName,
		bank.CountryISO2,
		bank.CountryName,
		bank.IsHeadquarter,
		bank.HeadquartersId,
		bank.SwiftCode,
		bank.Address).Scan(&bank.Id)
	if err != nil {
		return fmt.Errorf("failed to create bank: %w", err)
	}

	return nil
}

// DeleteBankBySwiftCode deletes a bank by its SWIFT code.
// Returns an error if the bank does not exist or if the deletion fails.
// Returns nil if the bank was successfully deleted.
func (d *DB) DeleteBankBySwiftCode(swiftCode string) error {
	query := `DELETE FROM banks WHERE swift_code = $1`
	result, err := d.SQL.Exec(query, swiftCode)
	if err != nil {
		return fmt.Errorf("failed to delete bank by SWIFT code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no bank found with SWIFT code: %s", swiftCode)
	}

	return nil
}

// GetCountryNameByISO2 retrieves the country name by its ISO2 code.
// Returns the country name as a string.
// Returns an error if the query fails.
func (d *DB) GetCountryNameByISO2(countryISO2 string) (string, error) {
	query := `SELECT countryName FROM banks WHERE countryISO2 = $1`
	row := d.SQL.QueryRow(query, countryISO2)

	var countryName string
	err := row.Scan(&countryName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no country found with ISO2 code: %s", countryISO2)
		}
		return "", fmt.Errorf("failed to retrieve country name by ISO2 code: %w", err)
	}

	return countryName, nil
}
