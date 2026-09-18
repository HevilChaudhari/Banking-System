package services

import (
	"testing"
	"time"

	"banking-system/internal/models"
	"banking-system/internal/repositories"
)

func TestCreateCustomer(t *testing.T) {
	repository := repositories.NewInMemoryCustomerRepository()
	service := NewInMemoryCustomerService(repository)

	customer := models.Customer{
		FirstName:   "Aarav",
		LastName:    "Sharma",
		Email:       "aarav.sharma@example.com",
		PhoneNumber: "+919876543210",
		DateOfBirth: time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC),
		Address:     "Mumbai, Maharashtra",
	}

	createdCustomer, err := service.CreateCustomer(customer)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if createdCustomer.CustomerID != 1 {
		t.Errorf("expected customer ID 1, got: %d", createdCustomer.CustomerID)
	}

	if createdCustomer.KYCStatus != models.KYCStatusNotStarted {
		t.Errorf("expected KYC status not_started, got: %s", createdCustomer.KYCStatus)
	}

	if createdCustomer.Status != models.CustomerStatusPendingVerification {
		t.Errorf("expected status pending_verification, got: %s", createdCustomer.Status)
	}
}
