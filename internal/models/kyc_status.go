package models

type KYCStatus string

const (
	KYCStatusNotStarted KYCStatus = "not_started"
	KYCStatusPending    KYCStatus = "pending"
	KYCStatusVerified   KYCStatus = "verified"
	KYCStatusRejected   KYCStatus = "rejected"
	KYCStatusExpired    KYCStatus = "expired"
)
