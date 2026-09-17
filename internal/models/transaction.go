package models

import "time"

type Transaction struct {
	TransactionID            string            `json:"transactionId"`
	SourceAccountNumber      string            `json:"sourceAccountNumber"`
	DestinationAccountNumber string            `json:"destinationAccountNumber"`
	Type                     TransactionType   `json:"transactionType"`
	Status                   TransactionStatus `json:"transactionStatus"`
	Amount                   int64             `json:"amount"`
	Currency                 CurrencyType      `json:"currency"`
	Description              string            `json:"description"`
	CreatedAt                time.Time         `json:"createdAt"`
	CompletedAt              time.Time         `json:"completedAt"`
}
