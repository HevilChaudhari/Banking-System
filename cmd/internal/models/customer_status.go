package models

type CustomerStatus string

const (
	CustomerStatusPendingVerification CustomerStatus = "pending_verification"
	CustomerStatusActive              CustomerStatus = "active"
	CustomerStatusSuspended           CustomerStatus = "suspended"
	CustomerStatusBlocked             CustomerStatus = "blocked"
	CustomerStatusInactive            CustomerStatus = "inactive"
	CustomerStatusClosed              CustomerStatus = "closed"
)
