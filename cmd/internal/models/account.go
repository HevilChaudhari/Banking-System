package models

import "time"

type Account struct {
	AccountNumber string        `json:"accountNumber"`
	CustomerID    int           `json:"customerId"`
	AccountType   string        `json:"accountType"`
	Balance       int64         `json:"balance"`
	Currency      CurrencyType  `json:"currency"`
	Status        AccountStatus `json:"status"`
	CreatedAt     time.Time     `json:"createdAt"`
}
