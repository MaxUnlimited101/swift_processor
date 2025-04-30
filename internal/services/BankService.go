package services

import (
	"maxunlimited.com/swift_processor/internal/database"
	"maxunlimited.com/swift_processor/internal/models"
)

type IBankService interface {
	// Bank methods
	GetBankBySwiftCode(swiftCode string) (any, error)
	GetBanksByCountry(countryISO2 string) (models.CountrySpecificBankDTO, error)
	CreateBank(bank *models.Bank) error
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

func (s *BankService) GetBankBySwiftCode(swiftCode string) (any, error) {
	bank, err := s.DB.GetBankBySwiftCode(swiftCode)
	if err != nil {
		return nil, err
	}
	if bank == nil {
		return nil, nil
	}
	if bank.IsHeadquarter {
		bankDTO := models.BankHequarterDTO{
			BankName:      bank.BankName,
			CountryISO2:   bank.CountryISO2,
			CountryName:   bank.CountryName,
			IsHeadquarter: bank.IsHeadquarter,
			SwiftCode:     bank.SwiftCode,
			Address:       bank.Address,
			Branches:      bank.Branches,
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

func (s *BankService) CreateBank(bank *models.BankBranchDTO) error {
	err := s.DB.CreateBank(bank)
	if err != nil {
		return err
	}
	return nil
}

func (s *BankService) DeleteBankBySwiftCode(swiftCode string) error {
	err := s.DB.DeleteBankBySwiftCode(swiftCode)
	if err != nil {
		return err
	}
	return nil
}
