CREATE TABLE banks (
    id SERIAL PRIMARY KEY, 
    bankName VARCHAR(255) NOT NULL, 
    countryISO2 CHAR(2) NOT NULL, 
    countryName VARCHAR(255) NOT NULL, 
    isHeadquarter BOOLEAN NOT NULL DEFAULT FALSE,
    headquartersId INT DEFAULT NULL, 
    swiftCode VARCHAR(11) NOT NULL UNIQUE, 
    address TEXT NOT NULL, 

    FOREIGN KEY (headquartersId) REFERENCES banks (id) ON DELETE SET NULL 
);