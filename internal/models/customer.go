package models

import "time"

type Customer struct {
	CustomerID  int            `json:"customerId"`
	FirstName   string         `json:"firstName"`
	LastName    string         `json:"lastName"`
	Email       string         `json:"email"`
	PhoneNumber string         `json:"phoneNumber"`
	DateOfBirth time.Time      `json:"dateOfBirth"`
	Address     string         `json:"address"`
	KYCStatus   KYCStatus      `json:"kycStatus"`
	CreatedAt   time.Time      `json:"createdAt"`
	Status      CustomerStatus `json:"status"`
	PasswordHash string         `json:"-"`
}
