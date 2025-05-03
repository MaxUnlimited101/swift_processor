package utils

import (
	"fmt"

	"maxunlimited.com/swift_processor/internal/models"
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

func (m *MockDB) UpdateBankHeadquartersId(bank *models.Bank) (int64, error) {
	// Mock implementation of UpdateBankHeadquartersId
	if existingBank, exists := m.Map[bank.SwiftCode]; exists {
		existingBank.HeadquartersId = bank.HeadquartersId
		return 1, nil
	}
	return 0, fmt.Errorf("Bank not found")
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
