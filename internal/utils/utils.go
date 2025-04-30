package utils

import (
	"encoding/csv"
	"fmt"
	"io"
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

		err = db.CreateBank(&bank)
		if err != nil {
			return fmt.Errorf("failed to insert bank data into database: %w", err)
		}

		bankMapBySwiftCode[bank.SwiftCode] = &bank
	}

	for _, bank := range bankMapBySwiftCode {
		if !bank.IsHeadquarter {
			// Find the headquarter bank and set HeadquartersId
			headquarter, exists := bankMapBySwiftCode[bank.SwiftCode[:len(bank.SwiftCode)-3]+"XXX"]
			if exists {
				bank.HeadquartersId = headquarter.Id
				err := db.UpdateBankHeadquartersId(bank)
				if err != nil {
					return fmt.Errorf("failed to update bank headquarters ID: %w", err)
				}
			} else {
				return fmt.Errorf("headquarter not found for bank with swift code: %s", bank.SwiftCode)
			}
		}
	}

	return nil
}
