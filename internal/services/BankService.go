package services

import (
	"database/sql"
	"fmt"

	"maxunlimited.com/swift_processor/internal/database"
	"maxunlimited.com/swift_processor/internal/models"
)

type IBankService interface {
	// GetBankBySwiftCode retrieves a bank by its SWIFT code.
	// If the bank is a headquarter, it returns the headquarter and its branches.
	// If the bank is not a headquarter, it returns the bank details.
	// If the bank is not found, it returns nil.
	// If an error occurs, it returns the error.
	GetBankBySwiftCode(swiftCode string) (any, error)

	// GetBanksByCountry retrieves all banks in a specific country by its ISO2 code.
	// If no banks are found, it returns an empty slice.
	// If an error occurs, it returns the error.
	GetBanksByCountry(countryISO2 string) (*models.CountrySpecificBankDTO, error)

	// CreateBank creates a new bank entry in the database.
	// If the bank already exists, it returns an error.
	// If an error occurs, it returns the error.
	CreateBank(bank *models.BankBranchDTO) error

	// DeleteBankBySwiftCode deletes a bank entry by its SWIFT code.
	// If the bank is not found, it returns nil.
	// If an error occurs, it returns the error.
	DeleteBankBySwiftCode(swiftCode string) error
}

type BankService struct {
	DB database.IDB
}

func NewBankService(db database.IDB) *BankService {
	return &BankService{
		DB: db,
	}
}

// GetBankBySwiftCode retrieves a bank by its SWIFT code.
// If the bank is a headquarter, it returns the headquarter and its branches.
// If the bank is not a headquarter, it returns the bank details.
// If the bank is not found, it returns nil.
// If an error occurs, it returns the error.
func (s *BankService) GetBankBySwiftCode(swiftCode string) (any, error) {
	bank, err := s.DB.GetBankBySwiftCode(swiftCode)
	if err != nil {
		return nil, err
	}
	if bank == nil {
		return nil, nil
	}
	if bank.IsHeadquarter {
		branchesDTO := make([]models.BankBranchDTO, len(bank.Branches))
		for i, branch := range bank.Branches {
			branchesDTO[i] = models.BankBranchDTO{
				Address:       branch.Address,
				BankName:      branch.BankName,
				CountryISO2:   branch.CountryISO2,
				CountryName:   branch.CountryName,
				IsHeadquarter: branch.IsHeadquarter,
				SwiftCode:     branch.SwiftCode,
			}
		}

		bankDTO := models.BankHequarterDTO{
			BankName:      bank.BankName,
			CountryISO2:   bank.CountryISO2,
			CountryName:   bank.CountryName,
			IsHeadquarter: bank.IsHeadquarter,
			SwiftCode:     bank.SwiftCode,
			Address:       bank.Address,
			Branches:      branchesDTO,
		}
		return &bankDTO, nil
	}
	// not a headquarter
	bankDTO := models.BankBranchDTO{
		Address:       bank.Address,
		BankName:      bank.BankName,
		CountryISO2:   bank.CountryISO2,
		CountryName:   bank.CountryName,
		IsHeadquarter: bank.IsHeadquarter,
		SwiftCode:     bank.SwiftCode,
	}
	return &bankDTO, nil
}

// GetBanksByCountry retrieves all banks in a specific country by its ISO2 code.
// If no banks are found, it returns an empty slice.
// If an error occurs, it returns the error.
// If the country is not found, it returns nil.
func (s *BankService) GetBanksByCountry(countryISO2 string) (*models.CountrySpecificBankDTO, error) {
	banks, err := s.DB.GetBanksByCountry(countryISO2)
	if err != nil {
		return nil, err
	}
	if len(banks) == 0 {
		countryName, err := s.DB.GetCountryNameByISO2(countryISO2)
		if err != nil {
			r := models.CountrySpecificBankDTO{
				CountryISO2: countryISO2,
				CountryName: countryName,
				SwiftCodes:  []models.BankDTOWithNoCountry{},
			}
			return &r, nil
		}
		return nil, err
	}
	swiftCodes := make([]models.BankDTOWithNoCountry, len(banks))
	for i, bank := range banks {
		swiftCodes[i] = models.BankDTOWithNoCountry{
			Address:       bank.Address,
			BankName:      bank.BankName,
			IsHeadquarter: bank.IsHeadquarter,
			SwiftCode:     bank.SwiftCode,
		}
	}
	r := models.CountrySpecificBankDTO{
		CountryISO2: countryISO2,
		CountryName: banks[0].CountryName,
		SwiftCodes:  swiftCodes,
	}
	return &r, nil
}

// CreateBank creates a new bank entry in the database.
// If the bank already exists, it returns an error.
// If an error occurs, it returns the error.
func (s *BankService) CreateBank(bank *models.BankBranchDTO) error {
	newbank := &models.Bank{
		BankName:      bank.BankName,
		CountryISO2:   bank.CountryISO2,
		CountryName:   bank.CountryName,
		IsHeadquarter: bank.IsHeadquarter,
		SwiftCode:     bank.SwiftCode,
		Address:       bank.Address,
		Branches:      make([]*models.Bank, 0),
	}
	if !newbank.IsHeadquarter {
		headquarterSwiftCode := bank.SwiftCode[:8] + "XXX"
		headquarter, err := s.DB.GetBankBySwiftCode(headquarterSwiftCode)
		if err != nil {
			newbank.HeadquartersId = sql.NullInt64{
				Int64: headquarter.Id,
				Valid: true,
			}
		}
		if headquarter == nil {
			return fmt.Errorf("headquarter not found for branch with SWIFT code: %s", bank.SwiftCode)
		}
	}
	err := s.DB.CreateBank(newbank)
	if err != nil {
		return err
	}
	return nil
}

// DeleteBankBySwiftCode deletes a bank entry by its SWIFT code.
// If the bank is not found, it returns nil.
// If an error occurs, it returns the error.
func (s *BankService) DeleteBankBySwiftCode(swiftCode string) error {
	err := s.DB.DeleteBankBySwiftCode(swiftCode)
	if err != nil {
		return err
	}
	return nil
}
