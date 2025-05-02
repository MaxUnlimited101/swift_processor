package utils

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"maxunlimited.com/swift_processor/internal/database"
	"maxunlimited.com/swift_processor/internal/models"
)

// ProcessCSV reads a CSV file and adds the data to the database
func ProcessCSV(filePath string, db database.IDB) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip the header row
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("failed to read header row: %w", err)
	}

	bankMapBySwiftCode := make(map[string]*models.Bank)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read record: %w", err)
		}

		bank := models.Bank{
			CountryISO2:   record[0],
			SwiftCode:     record[1],
			BankName:      record[3],
			Address:       record[4],
			CountryName:   record[6],
			IsHeadquarter: strings.HasSuffix(record[1], "XXX"),
		}

		if bank.IsHeadquarter {
			bank.HeadquartersId = sql.NullInt64{
				Int64: 0,
				Valid: false,
			}
			err = db.CreateBank(&bank)
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					log.Printf("bank with swift code %s already exists, skipping...", bank.SwiftCode)
					continue
				}
				return fmt.Errorf("failed to insert bank data into database: %w", err)
			}
		}

		bankMapBySwiftCode[bank.SwiftCode] = &bank
	}

	for _, bank := range bankMapBySwiftCode {
		if !bank.IsHeadquarter {
			// Find the headquarter bank and push to the database
			headquarter, exists := bankMapBySwiftCode[bank.SwiftCode[:8]+"XXX"]
			if exists {
				bank.HeadquartersId = sql.NullInt64{
					Int64: headquarter.Id,
					Valid: true,
				}
				err = db.CreateBank(bank)
				if err != nil {
					if strings.Contains(err.Error(), "duplicate key") {
						log.Printf("bank with swift code %s already exists, skipping...", bank.SwiftCode)
						continue
					}
					return fmt.Errorf("failed to insert bank data into database: %w", err)
				}
			} else {
				log.Printf("headquarter not found for bank with swift code: %s, continuing...", bank.SwiftCode)
			}
		}
	}

	return nil
}
