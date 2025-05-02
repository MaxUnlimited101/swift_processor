package tests

import (
	"fmt"
	"testing"

	"maxunlimited.com/swift_processor/internal/models"
	"maxunlimited.com/swift_processor/internal/services"
)

type MockDB struct {
	// Swift code to Bank mapping
	Map map[string]*models.Bank
}

func (m *MockDB) GetBankBySwiftCode(swiftCode string) (*models.Bank, error) {
	// Mock implementation of GetBankBySwiftCode
	if bank, exists := m.Map[swiftCode]; exists {
		return bank, nil
	}
	return nil, nil
}

func (m *MockDB) GetBanksByCountry(countryISO2 string) ([]*models.Bank, error) {
	// Mock implementation of GetBanksByCountry
	var banks []*models.Bank
	for _, bank := range m.Map {
		if bank.CountryISO2 == countryISO2 {
			banks = append(banks, bank)
		}
	}
	return banks, nil
}

func (m *MockDB) GetCountryNameByISO2(countryISO2 string) (string, error) {
	// Mock implementation of GetCountryNameByISO2
	for _, bank := range m.Map {
		if bank.CountryISO2 == countryISO2 {
			return bank.CountryName, nil
		}
	}
	return "", nil
}

func (m *MockDB) CreateBank(bank *models.Bank) error {
	// Mock implementation of CreateBank
	if _, exists := m.Map[bank.SwiftCode]; exists {
		return fmt.Errorf("Duplicate entry") // Simulate duplicate entry
	}
	m.Map[bank.SwiftCode] = bank
	return nil
}

func (m *MockDB) UpdateBankHeadquartersId(bank *models.Bank) error {
	// Mock implementation of UpdateBankHeadquartersId
	if existingBank, exists := m.Map[bank.SwiftCode]; exists {
		existingBank.HeadquartersId = bank.HeadquartersId
		return nil
	}
	return fmt.Errorf("Bank not found")
}

func (m *MockDB) DeleteBankBySwiftCode(swiftCode string) error {
	// Mock implementation of DeleteBankBySwiftCode
	if _, exists := m.Map[swiftCode]; exists {
		delete(m.Map, swiftCode)
		return nil
	}
	return fmt.Errorf("Bank not found")
}

func (m *MockDB) Close() {
	// Mock implementation of Close with no return value
}

func TestGetBankBySwiftCode(t *testing.T) {
	branches := make([]*models.Bank, 0)
	branches = append(branches, &models.Bank{
		BankName:      "Branch 1",
		CountryISO2:   "US",
		CountryName:   "United States",
		IsHeadquarter: false,
		SwiftCode:     "SWIFT456",
		Address:       "456 Branch St",
	})

	mockDB := &MockDB{
		Map: map[string]*models.Bank{
			"SWIFT123": {
				BankName:      "Test Bank",
				CountryISO2:   "US",
				CountryName:   "United States",
				IsHeadquarter: true,
				SwiftCode:     "SWIFT123",
				Address:       "123 Test St",
				Branches:      branches,
			},
			"SWIFT456": {
				BankName:      "Branch 1",
				CountryISO2:   "US",
				CountryName:   "United States",
				IsHeadquarter: false,
				SwiftCode:     "SWIFT456",
				Address:       "456 Branch St",
			},
		},
	}

	service := services.NewBankService(mockDB)

	t.Run("Get existing headquarter bank by swift code", func(t *testing.T) {
		result, err := service.GetBankBySwiftCode("SWIFT123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatalf("Expected a result, got nil")
		}
	})

	t.Run("Get existing branch bank by swift code", func(t *testing.T) {
		result, err := service.GetBankBySwiftCode("SWIFT456")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatalf("Expected a result, got nil")
		}

		// check if reslut type is correct
		branchResult, ok := result.(*models.BankBranchDTO)
		if !ok {
			t.Fatalf("Expected result to be of type BankBranchDTO, got %T", result)
		}

		if branchResult.BankName != "Branch 1" {
			t.Fatalf("Expected Branch 1, got %s", branchResult.BankName)
		}
	})

	t.Run("Get non-existing bank by swift code", func(t *testing.T) {
		result, err := service.GetBankBySwiftCode("NONEXISTENT")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != nil {
			t.Fatalf("Expected nil, got %v", result)
		}
	})
}

func TestGetBanksByCountry(t *testing.T) {
	mockDB := &MockDB{
		Map: map[string]*models.Bank{
			"SWIFTXXX": {
				BankName:      "Test Bank",
				CountryISO2:   "US",
				CountryName:   "United States",
				IsHeadquarter: true,
				SwiftCode:     "SWIFT123",
				Address:       "123 Test St",
			},
			"SWIFT456": {
				BankName:      "Another Bank",
				CountryISO2:   "US",
				CountryName:   "United States",
				IsHeadquarter: false,
				SwiftCode:     "SWIFT456",
				Address:       "456 Another St",
			},
			"ABCXXX": {
				BankName:      "ABC Bank",
				CountryISO2:   "PL",
				CountryName:   "Poland",
				IsHeadquarter: true,
				SwiftCode:     "ABCXXX",
				Address:       "123 St St",
			},
		},
	}

	service := services.NewBankService(mockDB)

	t.Run("Get banks by existing country", func(t *testing.T) {
		result, err := service.GetBanksByCountry("US")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil || len(result.SwiftCodes) != 2 {
			t.Fatalf("Expected 2 banks, got %v", result)
		}
	})

	t.Run("Get banks by non-existing country", func(t *testing.T) {
		result, err := service.GetBanksByCountry("XX")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != nil && len(result.SwiftCodes) != 0 {
			t.Fatalf("Expected 0 banks, got %v", result)
		}
	})
}

func TestCreateBank(t *testing.T) {
	mockDB := &MockDB{
		Map: make(map[string]*models.Bank),
	}

	service := services.NewBankService(mockDB)

	t.Run("Create a new bank", func(t *testing.T) {
		bank := &models.BankBranchDTO{
			BankName:      "New Bank",
			CountryISO2:   "US",
			CountryName:   "United States",
			IsHeadquarter: true,
			SwiftCode:     "SWIFTXXX",
			Address:       "789 New St",
		}
		err := service.CreateBank(bank)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if _, exists := mockDB.Map["SWIFTXXX"]; !exists {
			t.Fatalf("Expected bank to be created, but it was not")
		}
	})

	t.Run("Create a duplicate bank", func(t *testing.T) {
		bank := &models.BankBranchDTO{
			BankName:      "Duplicate Bank",
			CountryISO2:   "PL",
			CountryName:   "Poland",
			IsHeadquarter: false,
			SwiftCode:     "SWIFTXXX",
			Address:       "789 Duplicate St",
		}
		err := service.CreateBank(bank)
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})

	t.Run("Create a bank with invalid headquaters", func(t *testing.T) {
		bank := &models.BankBranchDTO{
			BankName:      "No Headquaters Bank",
			CountryISO2:   "PL",
			CountryName:   "Poland",
			IsHeadquarter: false,
			SwiftCode:     "ADS123",
			Address:       "789 Duplicate St",
		}
		err := service.CreateBank(bank)
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
}

func TestDeleteBankBySwiftCode(t *testing.T) {
	mockDB := &MockDB{
		Map: map[string]*models.Bank{
			"SWIFT123": {
				BankName:      "Test Bank",
				CountryISO2:   "US",
				CountryName:   "United States",
				IsHeadquarter: true,
				SwiftCode:     "SWIFT123",
				Address:       "123 Test St",
			},
		},
	}

	service := services.NewBankService(mockDB)

	t.Run("Delete existing bank", func(t *testing.T) {
		err := service.DeleteBankBySwiftCode("SWIFT123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if _, exists := mockDB.Map["SWIFT123"]; exists {
			t.Fatalf("Expected bank to be deleted, but it still exists")
		}
	})

	t.Run("Delete non-existing bank", func(t *testing.T) {
		err := service.DeleteBankBySwiftCode("NONEXISTENT")
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})
}
