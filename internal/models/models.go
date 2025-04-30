package models

type Bank struct {
	Id             int64  `json:"id"`
	BankName       string `json:"bankName"`
	CountryISO2    string `json:"countryISO2"`
	CountryName    string `json:"countryName"`
	IsHeadquarter  bool   `json:"isHeadquarter"`
	HeadquartersId int64  `json:"headquartersId"`
	SwiftCode      string `json:"swiftCode"`
	Address        string `json:"address"`
	Branches       []Bank `json:"branches"`
}

type BankHequarterDTO struct {
	BankName      string          `json:"bankName"`
	CountryISO2   string          `json:"countryISO2"`
	CountryName   string          `json:"countryName"`
	IsHeadquarter bool            `json:"isHeadquarter"`
	SwiftCode     string          `json:"swiftCode"`
	Address       string          `json:"address"`
	Branches      []BankBranchDTO `json:"branches"`
}

type BankBranchDTO struct {
	Address       string `json:"address"`
	BankName      string `json:"bankName"`
	CountryISO2   string `json:"countryISO2"`
	CountryName   string `json:"countryName"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
}

type BankDTOWithNoCountry struct {
	Address       string `json:"address"`
	BankName      string `json:"bankName"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
}

type CountrySpecificBankDTO struct {
	CountryISO2 string                 `json:"countryISO2"`
	CountryName string                 `json:"countryName"`
	SwiftCodes  []BankDTOWithNoCountry `json:"swiftCodes"`
}
