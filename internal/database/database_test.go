package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"maxunlimited.com/swift_processor/internal/models"
)

func TestGetBankBySwiftCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	database := &DB{SQL: db}

	t.Run("Headquaters Bank found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "bankName", "countryISO2", "countryName", "isHeadquarter", "headquartersId", "swiftCode", "address"}).
			AddRow(1, "Test Bank", "US", "United States", true, nil, "SWIFT111XXX", "123 Test St").
			AddRow(2, "Test Bank branch", "PL", "Poland", false, 1, "SWIFT111345", "123 Poland St")
		mock.ExpectQuery("SELECT id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address FROM banks WHERE swiftCode = \\$1").
			WithArgs("SWIFT123").
			WillReturnRows(rows)
		mock.ExpectQuery("SELECT id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address FROM banks WHERE headquartersId = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "bankName", "countryISO2", "countryName", "isHeadquarter", "headquartersId", "swiftCode", "address"}).
				AddRow(2, "Test Bank branch", "PL", "Poland", false, 1, "SWIFT111345", "123 Poland St"))

		bank, err := database.GetBankBySwiftCode("SWIFT123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if bank == nil || bank.BankName != "Test Bank" {
			t.Fatalf("Expected bank 'Test Bank', got %v", bank)
		}
	})

	t.Run("Branch Bank found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "bankName", "countryISO2", "countryName", "isHeadquarter", "headquartersId", "swiftCode", "address"}).
			AddRow(2, "Test Bank branch", "PL", "Poland", false, 1, "SWIFT111345", "123 Poland St")
		mock.ExpectQuery("SELECT id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address FROM banks WHERE swiftCode = \\$1").
			WithArgs("SWIFT111345").
			WillReturnRows(rows)

		bank, err := database.GetBankBySwiftCode("SWIFT111345")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if bank == nil || bank.BankName != "Test Bank branch" {
			t.Fatalf("Expected bank 'Test Bank branch', got %v", bank)
		}
		if bank.HeadquartersId.Int64 != 1 {
			t.Fatalf("Expected headquarters ID 1, got %d", bank.HeadquartersId.Int64)
		}
	})

	t.Run("Bank not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address FROM banks WHERE swiftCode = \\$1").
			WithArgs("NONEXISTENT").
			WillReturnError(sql.ErrNoRows)

		bank, err := database.GetBankBySwiftCode("NONEXISTENT")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if bank != nil {
			t.Fatalf("Expected nil, got %v", bank)
		}
	})
}

func TestGetBanksByCountry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	database := &DB{SQL: db}

	t.Run("Banks found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "bankName", "countryISO2", "countryName", "isHeadquarter", "headquartersId", "swiftCode", "address"}).
			AddRow(1, "Bank 1", "US", "United States", true, nil, "SWIFT123", "123 Test St").
			AddRow(2, "Bank 2", "US", "United States", false, 1, "SWIFT456", "456 Test St")
		mock.ExpectQuery("SELECT id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address FROM banks WHERE countryISO2 = \\$1").
			WithArgs("US").
			WillReturnRows(rows)

		banks, err := database.GetBanksByCountry("US")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(banks) != 2 {
			t.Fatalf("Expected 2 banks, got %d", len(banks))
		}
	})

	t.Run("No banks found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address FROM banks WHERE countryISO2 = \\$1").
			WithArgs("XX").
			WillReturnRows(sqlmock.NewRows(nil))

		banks, err := database.GetBanksByCountry("XX")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(banks) != 0 {
			t.Fatalf("Expected 0 banks, got %d", len(banks))
		}
	})
}

func TestCreateBank(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	database := &DB{SQL: db}

	t.Run("Create headquarter bank", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO banks \\(bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6, \\$7\\) RETURNING id").
			WithArgs("Test Bank", "US", "United States", true, nil, "SWIFT123", "123 Test St").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		bank := &models.Bank{
			BankName:      "Test Bank",
			CountryISO2:   "US",
			CountryName:   "United States",
			IsHeadquarter: true,
			SwiftCode:     "SWIFT123",
			Address:       "123 Test St",
		}
		err := database.CreateBank(bank)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if bank.Id != 1 {
			t.Fatalf("Expected bank ID to be 1, got %d", bank.Id)
		}
	})

	t.Run("Create branch bank", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO banks \\(bankName, countryISO2, countryName, isHeadquarter, headquartersId, swiftCode, address\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6, \\$7\\) RETURNING id").
			WithArgs("Branch Bank", "US", "United States", false, 1, "SWIFT456", "456 Test St").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

		bank := &models.Bank{
			BankName:      "Branch Bank",
			CountryISO2:   "US",
			CountryName:   "United States",
			IsHeadquarter: false,
			HeadquartersId: sql.NullInt64{
				Int64: 1,
				Valid: true,
			},
			SwiftCode: "SWIFT456",
			Address:   "456 Test St",
		}
		err := database.CreateBank(bank)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if bank.Id != 2 {
			t.Fatalf("Expected bank ID to be 2, got %d", bank.Id)
		}
	})
}

func TestDeleteBankBySwiftCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	database := &DB{SQL: db}

	t.Run("Delete existing bank", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM banks WHERE swiftcode = \\$1").
			WithArgs("SWIFT123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := database.DeleteBankBySwiftCode("SWIFT123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	t.Run("Delete non-existing bank", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM banks WHERE swiftcode = \\$1").
			WithArgs("NONEXISTENT").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := database.DeleteBankBySwiftCode("NONEXISTENT")
		if err == nil || !strings.Contains(err.Error(), "no bank found") {
			t.Fatalf("Expected [no bank found], got: %v", err)
		}
	})
}

func TestGetCountryNameByISO2(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	database := &DB{SQL: db}

	t.Run("Country found", func(t *testing.T) {
		mock.ExpectQuery("SELECT countryName FROM banks WHERE countryISO2 = \\$1").
			WithArgs("US").
			WillReturnRows(sqlmock.NewRows([]string{"countryName"}).AddRow("United States"))

		countryName, err := database.GetCountryNameByISO2("US")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if countryName != "United States" {
			t.Fatalf("Expected 'United States', got %s", countryName)
		}
	})

	t.Run("Country not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT countryName FROM banks WHERE countryISO2 = \\$1").
			WithArgs("XX").
			WillReturnError(sql.ErrNoRows)

		countryName, err := database.GetCountryNameByISO2("XX")
		if err == nil {
			t.Fatalf("Expected an error, got %v", err)
		}
		if countryName != "" {
			t.Fatalf("Expected empty string, got %s", countryName)
		}
	})
}

func TestUpdateBankHeadquartersId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	database := &DB{SQL: db}

	t.Run("Update headquarters ID", func(t *testing.T) {
		bank := &models.Bank{
			SwiftCode:      "abcABC123",
			HeadquartersId: sql.NullInt64{Int64: 1, Valid: true},
		}
		mock.ExpectExec("UPDATE banks SET headquartersId = \\$1 WHERE swiftCode = \\$2").
			WithArgs(bank.HeadquartersId, bank.SwiftCode).
			WillReturnResult(sqlmock.NewResult(1, 1))

		rows, err := database.UpdateBankHeadquartersId(bank)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if rows != 1 {
			t.Fatalf("Expected 1 row affected, got %d", rows)
		}
	})

	t.Run("Update non-existing bank", func(t *testing.T) {
		mock.ExpectExec("UPDATE banks SET headquartersId = \\$1 WHERE swiftCode = \\$2").
			WithArgs(sql.NullInt64{Int64: 1, Valid: true}, "abcABC123").
			WillReturnResult(sqlmock.NewResult(0, 0))

		bank := &models.Bank{
			SwiftCode:      "abcABC123",
			HeadquartersId: sql.NullInt64{Int64: 1, Valid: true},
		}
		rows, err := database.UpdateBankHeadquartersId(bank)
		if err != nil {
			t.Fatalf("Expected not error, got: %v", err)
		}
		if rows != 0 {
			t.Fatalf("Expected 0 rows affected, got %d", rows)
		}
	})
}
