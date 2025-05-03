package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"maxunlimited.com/swift_processor/internal/models"
	"maxunlimited.com/swift_processor/internal/services"
)

// SendJsonResponse sends a JSON response with the given status code and data
// It sets the Content-Type header to application/json and encodes the data into JSON format
// If encoding fails, it logs the error and sends an internal server error response
func SendJsonResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type IHandler interface {
	GetBankBySwiftCode(w http.ResponseWriter, r *http.Request)
	GetBanksByCountry(w http.ResponseWriter, r *http.Request)
	CreateBank(w http.ResponseWriter, r *http.Request)
	DeleteBankBySwiftCode(w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	BankService services.IBankService
}

func NewHandler(bankService services.IBankService) *Handler {
	return &Handler{
		BankService: bankService,
	}
}

func (h *Handler) GetBankBySwiftCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swiftCode := vars["swift-code"]
	if swiftCode == "" {
		http.Error(w, "Missing swiftCode parameter", http.StatusBadRequest)
		return
	}

	bank, err := h.BankService.GetBankBySwiftCode(swiftCode)
	if err != nil {
		http.Error(w, "Error fetching bank data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if bank == nil {
		http.Error(w, "Bank not found", http.StatusNotFound)
		return
	}

	SendJsonResponse(w, http.StatusOK, bank)
}

func (h *Handler) GetBanksByCountry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	countryISO2 := vars["countryISO2code"]
	if countryISO2 == "" {
		http.Error(w, "Missing countryISO2 parameter", http.StatusBadRequest)
		return
	}

	banks, err := h.BankService.GetBanksByCountry(countryISO2)
	if err != nil {
		http.Error(w, "Error fetching banks data", http.StatusInternalServerError)
		return
	}

	if banks == nil {
		http.Error(w, "No such country found in DB", http.StatusNotFound)
	}

	SendJsonResponse(w, http.StatusOK, banks)
}

func (h *Handler) CreateBank(w http.ResponseWriter, r *http.Request) {
	var bank models.BankBranchDTO
	if err := json.NewDecoder(r.Body).Decode(&bank); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.BankService.CreateBank(&bank); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Bank with this SWIFT code already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Error creating bank", http.StatusInternalServerError)
		return
	}

	SendJsonResponse(w, http.StatusCreated, map[string]string{"message": "Bank created successfully"})
}

func (h *Handler) DeleteBankBySwiftCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swiftCode := vars["swift-code"]
	if swiftCode == "" || len(swiftCode) == 0 {
		http.Error(w, "Missing swiftCode parameter", http.StatusBadRequest)
		return
	}

	if err := h.BankService.DeleteBankBySwiftCode(swiftCode); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Bank not found: "+err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Error deleting bank: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
