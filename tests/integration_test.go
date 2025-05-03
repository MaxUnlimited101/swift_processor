package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"maxunlimited.com/swift_processor/internal/handlers"
	"maxunlimited.com/swift_processor/internal/models"
	"maxunlimited.com/swift_processor/internal/services"
	"maxunlimited.com/swift_processor/internal/utils"
)

func TestIntegrationHandlersAndServices(t *testing.T) {
	// Mock data
	mockDB := &utils.MockDB{
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

	// Initialize service and handler
	bankService := services.NewBankService(mockDB)
	handler := handlers.NewHandler(bankService)

	// Setup router
	r := mux.NewRouter()
	r.Path("/v1/swift-codes/{swift-code}").Methods("GET").HandlerFunc(handler.GetBankBySwiftCode)
	r.Path("/v1/swift-codes/country/{countryISO2code}").Methods("GET").HandlerFunc(handler.GetBanksByCountry)
	r.Path("/v1/swift-codes").Methods("POST").HandlerFunc(handler.CreateBank)
	r.Path("/v1/swift-codes/{swift-code}").Methods("DELETE").HandlerFunc(handler.DeleteBankBySwiftCode)

	t.Run("Integration Test - GetBankBySwiftCode", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/swift-codes/SWIFT123", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Expected status OK, got %v", resp.Code)
		}
	})

	t.Run("Integration Test - GetBankBySwiftCode with non-existing code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/swift-codes/NONEXISTENT", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("Expected status Not Found, got %v", resp.Code)
		}
	})

	t.Run("Integration Test - GetBankBySwiftCode with empty code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/swift-codes/", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("Expected status Not found, got %v", resp.Code)
		}
	})

	t.Run("Integration Test - CreateBank", func(t *testing.T) {
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

	t.Run("Integration Test - CreateBank with already existing bank", func(t *testing.T) {
		bank := &models.BankBranchDTO{
			BankName:      "Test Bank",
			CountryISO2:   "US",
			CountryName:   "United States",
			IsHeadquarter: true,
			SwiftCode:     "SWIFT123",
			Address:       "123 Test St",
		}
		body, _ := json.Marshal(bank)
		req := httptest.NewRequest("POST", "/v1/swift-codes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		if resp.Code != http.StatusConflict {
			t.Fatalf("Expected status Conflict, got %v", resp.Code)
		}
	})

	t.Run("Integration Test - DeleteBankBySwiftCode", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/swift-codes/SWIFT123", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusNoContent {
			t.Fatalf("Expected status No Content, got %v", resp.Code)
		}
	})

	t.Run("Integration Test - DeleteBankBySwiftCode with non-existing code", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/swift-codes/NONEXISTENT", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("Expected status Bad request, got %v", resp.Body.String())
		}
	})

	t.Run("Integration Test - DeleteBankBySwiftCode with empty code", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/swift-codes/", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("Expected status Not found, got %v", resp.Code)
		}
	})

	t.Run("Integration Test - GetBanksByCountry", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/swift-codes/country/US", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Expected status OK, got %v", resp.Code)
		}
	})

	t.Run("Integration Test - GetBanksByCountry with non-existing country code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/swift-codes/country/XX", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusNotFound {
			t.Fatalf("Expected status NotFound, got %v", resp.Code)
		}
	})

	t.Run("Integration Test - GetBanksByCountry with empty country code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/swift-codes/country/", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusNotFound {
			t.Fatalf("Expected status NotFound, got %v", resp.Code)
		}
	})

	t.Run("Integration Test - GetBanksByCountry with non-existing country code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/swift-codes/country/XX", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusNotFound {
			t.Fatalf("Expected status Not Found, got %v", resp.Code)
		}
	})
}
