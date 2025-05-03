package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	expected := &Config{
		PostgresConnectionString: "postgres://user:password@localhost:5432/dbname",
		ServerPort:               "3000",
		InitialCSVDataPath:       "/path/to/csv",
	}

	// create a temporary .env file for testing
	err := os.WriteFile(".env", []byte("POSTGRES_CONNECTION_STRING=postgres://user:password@localhost:5432/dbname\nPORT=3000\nINITIAL_CSV_DATA_PATH=/path/to/csv"), 0644)
	if err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}
	// defer cleanup of the .env file
	defer func() {
		err := os.Remove(".env")
		if err != nil {
			t.Fatalf("failed to remove .env file: %v", err)
		}
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("expected %+v, got %+v", expected, cfg)
	}
}
