package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"maxunlimited.com/swift_processor/internal/handlers"
	"maxunlimited.com/swift_processor/internal/models"
)

type MockBankService struct {
	banks map[string]*models.Bank
}

func (m *MockBankService) GetBankBySwiftCode(swiftCode string) (any, error) {
	bank, exists := m.banks[swiftCode]
	if !exists {
		return nil, nil
	}
	return bank, nil
}

func (m *MockBankService) CreateBank(bank *models.BankBranchDTO) error {
	if _, exists := m.banks[bank.SwiftCode]; exists {
		return fmt.Errorf("bank already exists")
	}
	m.banks[bank.SwiftCode] = &models.Bank{
		BankName:      bank.BankName,
		CountryISO2:   bank.CountryISO2,
		CountryName:   bank.CountryName,
		IsHeadquarter: bank.IsHeadquarter,
		SwiftCode:     bank.SwiftCode,
		Address:       bank.Address,
	}
	return nil
}

func (m *MockBankService) DeleteBankBySwiftCode(swiftCode string) error {
	if _, exists := m.banks[swiftCode]; !exists {
		return fmt.Errorf("bank not found")
	}
	delete(m.banks, swiftCode)
	return nil
}

func (m *MockBankService) GetBanksByCountry(countryISO2 string) (*models.CountrySpecificBankDTO, error) {
	b := &models.CountrySpecificBankDTO{
		CountryISO2: "",
		CountryName: "",
		SwiftCodes:  make([]models.BankDTOWithNoCountry, 0),
	}

	for _, bank := range m.banks {
		if bank.CountryISO2 == countryISO2 {
			c := models.BankDTOWithNoCountry{
				Address:       bank.Address,
				BankName:      bank.BankName,
				IsHeadquarter: bank.IsHeadquarter,
				SwiftCode:     bank.SwiftCode,
			}
			b.SwiftCodes = append(b.SwiftCodes, c)
		}
	}
	return b, nil
}

func TestGetBankBySwiftCodeHandler(t *testing.T) {
	mockService := &MockBankService{
		banks: map[string]*models.Bank{
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
	handler := handlers.NewHandler(mockService)
	r := mux.NewRouter()
	r.HandleFunc("/v1/swift-codes/{swift-code}", handler.GetBankBySwiftCode).Methods("GET")

	t.Run("Bank found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/swift-codes/SWIFT123", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Expected status OK, got %v", resp.Code)
		}
	})

	t.Run("Bank not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/swift-codes/NONEXISTENT", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusNotFound {
			t.Fatalf("Expected status Not Found, got %v", resp.Code)
		}
	})
}

func TestCreateBankHandler(t *testing.T) {
	mockService := &MockBankService{
		banks: make(map[string]*models.Bank),
	}
	handler := handlers.NewHandler(mockService)
	r := mux.NewRouter()
	r.HandleFunc("/v1/swift-codes", handler.CreateBank).Methods("POST")

	t.Run("Create bank successfully", func(t *testing.T) {
		bank := &models.BankBranchDTO{
			BankName:      "New Bank",
			CountryISO2:   "US",
			CountryName:   "United States",
			IsHeadquarter: true,
			SwiftCode:     "SWIFT456",
			Address:       "456 New St",
		}
		body, _ := json.Marshal(bank)
		req := httptest.NewRequest("POST", "/v1/swift-codes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusCreated {
			t.Fatalf("Expected status Created, got %v", resp.Code)
		}
	})

	t.Run("Create duplicate bank", func(t *testing.T) {
		bank := &models.BankBranchDTO{
			BankName:      "Duplicate Bank",
			CountryISO2:   "US",
			CountryName:   "United States",
			IsHeadquarter: true,
			SwiftCode:     "SWIFT456",
			Address:       "456 New St",
		}
		body, _ := json.Marshal(bank)
		req := httptest.NewRequest("POST", "/v1/swift-codes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusInternalServerError {
			t.Fatalf("Expected status Internal Server Error, got %v", resp.Code)
		}
	})
}
