package server

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"maxunlimited.com/swift_processor/internal/handlers"
)

func TestNewRouter(t *testing.T) {
	handler := &handlers.Handler{}
	router := NewRouter(handler)

	tests := []struct {
		path    string
		method  string
		handler http.HandlerFunc
	}{
		{"/v1/swift-codes/{swift-code}", "GET", handler.GetBankBySwiftCode},
		{"/v1/swift-codes/country/{countryISO2code}", "GET", handler.GetBanksByCountry},
		{"/v1/swift-codes", "POST", handler.CreateBank},
		{"/v1/swift-codes/{swift-code}", "DELETE", handler.DeleteBankBySwiftCode},
	}

	for _, test := range tests {
		req, err := http.NewRequest(test.method, test.path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		var matchedRoute mux.RouteMatch
		matched := router.Match(req, &matchedRoute)
		if !matched {
			t.Fatalf("Expected route for %s %s not found", test.method, test.path)
			continue
		}

		if reflect.ValueOf(matchedRoute.Handler).Pointer() != reflect.ValueOf(test.handler).Pointer() {
			t.Fatalf("Expected handler %v, got %v", test.handler, matchedRoute.Handler)
		}
	}
}

func TestNewServer(t *testing.T) {
	addr := ":8080"
	handler := http.NotFoundHandler()
	srv := NewServer(addr, handler)

	if srv.Addr != addr {
		t.Fatalf("Expected server address %s, got %s", addr, srv.Addr)
	}

	if reflect.ValueOf(srv.Handler).Pointer() != reflect.ValueOf(handler).Pointer() {
		t.Fatalf("Expected server handler %v, got %v", handler, srv.Handler)
	}

	if srv.ReadTimeout != 10*time.Second {
		t.Fatalf("Expected ReadTimeout 10s, got %v", srv.ReadTimeout)
	}

	if srv.WriteTimeout != 10*time.Second {
		t.Fatalf("Expected WriteTimeout 10s, got %v", srv.WriteTimeout)
	}

	if srv.IdleTimeout != 120*time.Second {
		t.Fatalf("Expected IdleTimeout 120s, got %v", srv.IdleTimeout)
	}
}
