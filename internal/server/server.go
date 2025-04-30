package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"maxunlimited.com/swift_processor/internal/handlers"
)

// NewRouter creates and configures the mux router
func NewRouter(handler *handlers.Handler) *mux.Router {
	r := mux.NewRouter()

	r.Path("/v1/swift-codes/{swift-code}").Methods("GET").HandlerFunc(handler.GetBankBySwiftCode)
	r.Path("/v1/swift-codes/country/{countryISO2code}").Methods("GET").HandlerFunc(handler.GetBanksByCountry)
	r.Path("/v1/swift-codes").Methods("POST").HandlerFunc(handler.CreateBank)
	r.Path("/v1/swift-codes/{swift-code}").Methods("DELETE").HandlerFunc(handler.DeleteBankBySwiftCode)

	// Add a simple health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return r
}

// NewServer creates a custom http.Server
func NewServer(addr string, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return srv
}
