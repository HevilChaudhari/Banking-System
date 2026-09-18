package services

import (
	"testing"
	"time"

	"banking-system/internal/models"
	"banking-system/internal/repositories"
)

func getCustomerData() models.Customer {
	return models.Customer{
		FirstName:   "Aarav",
		LastName:    "Sharma",
		Email:       "aarav.sharma@example.com",
		PhoneNumber: "+919876543210",
		DateOfBirth: time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC),
		Address:     "Mumbai, Maharashtra",
	}
}

func TestCreateCustomer_Success(t *testing.T) {
	repository := repositories.NewInMemoryCustomerRepository()
	service := NewInMemoryCustomerService(repository)

	customer := getCustomerData()
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

	if createdCustomer.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}
}

func TestCreateCustomer_ValidationErrors(t *testing.T) {
	repository := repositories.NewInMemoryCustomerRepository()
	service := NewInMemoryCustomerService(repository)

	testCases := []struct {
		name        string
		modify      func(c *models.Customer)
		expectedErr string
	}{
		{
			name: "empty first name",
			modify: func(c *models.Customer) {
				c.FirstName = "   "
			},
			expectedErr: "firstname or lastname cannot be empty",
		},
		{
			name: "empty last name",
			modify: func(c *models.Customer) {
				c.LastName = ""
			},
			expectedErr: "firstname or lastname cannot be empty",
		},
		{
			name: "invalid email format",
			modify: func(c *models.Customer) {
				c.Email = "invalid-email-string"
			},
			expectedErr: "email is invalid",
		},
		{
			name: "invalid phone number without plus prefix",
			modify: func(c *models.Customer) {
				c.PhoneNumber = "9876543210"
			},
			expectedErr: "phone number is invalid",
		},
		{
			name: "invalid phone number with bad digits",
			modify: func(c *models.Customer) {
				c.PhoneNumber = "+91123"
			},
			expectedErr: "phone number is invalid",
		},
		{
			name: "underage customer (under 18)",
			modify: func(c *models.Customer) {
				c.DateOfBirth = time.Now().AddDate(-17, 0, 0)
			},
			expectedErr: "customer must be at least 18 years old",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cust := getCustomerData()
			tc.modify(&cust)

			_, err := service.CreateCustomer(cust)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.expectedErr)
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("expected error %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestCreateCustomer_DuplicateEmail(t *testing.T) {
	repository := repositories.NewInMemoryCustomerRepository()
	service := NewInMemoryCustomerService(repository)

	first := getCustomerData()
	_, err := service.CreateCustomer(first)
	if err != nil {
		t.Fatalf("expected first creation to succeed, got: %v", err)
	}

	second := getCustomerData()
	second.FirstName = "Different"
	second.LastName = "Person"
	second.PhoneNumber = "+919999988888"

	_, err = service.CreateCustomer(second)
	if err == nil {
		t.Fatal("expected duplicate email error, got nil")
	}

	expectedErr := "customer already exist"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestGetCustomerByID(t *testing.T) {
	repository := repositories.NewInMemoryCustomerRepository()
	service := NewInMemoryCustomerService(repository)

	t.Run("fails when customer ID is less than or equal to zero", func(t *testing.T) {
		_, err := service.GetCustomerByID(0)
		if err == nil {
			t.Fatal("expected error for ID 0, got nil")
		}
		if err.Error() != "customer id must be greater than zero" {
			t.Errorf("unexpected error: %v", err)
		}

		_, err = service.GetCustomerByID(-5)
		if err == nil {
			t.Fatal("expected error for ID -5, got nil")
		}
	})

	t.Run("fails when customer is not found", func(t *testing.T) {
		_, err := service.GetCustomerByID(999)
		if err == nil {
			t.Fatal("expected error for non-existent ID, got nil")
		}
		if err.Error() != "customer not found" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("successfully retrieves created customer", func(t *testing.T) {
		created, err := service.CreateCustomer(getCustomerData())
		if err != nil {
			t.Fatalf("failed to create customer: %v", err)
		}

		found, err := service.GetCustomerByID(created.CustomerID)
		if err != nil {
			t.Fatalf("expected to find customer, got error: %v", err)
		}

		if found.CustomerID != created.CustomerID {
			t.Errorf("expected ID %d, got %d", created.CustomerID, found.CustomerID)
		}
		if found.Email != created.Email {
			t.Errorf("expected email %s, got %s", created.Email, found.Email)
		}
	})
}

func TestGetCustomerByEmail(t *testing.T) {
	repository := repositories.NewInMemoryCustomerRepository()
	service := NewInMemoryCustomerService(repository)

	t.Run("fails when email is invalid", func(t *testing.T) {
		_, err := service.GetCustomerByEmail("bad-email")
		if err == nil {
			t.Fatal("expected error for invalid email, got nil")
		}
		if err.Error() != "invalid email" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("fails when email is not found", func(t *testing.T) {
		_, err := service.GetCustomerByEmail("nonexistent@example.com")
		if err == nil {
			t.Fatal("expected error for not found email, got nil")
		}
		if err.Error() != "customer not found" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("successfully retrieves customer by email", func(t *testing.T) {
		created, err := service.CreateCustomer(getCustomerData())
		if err != nil {
			t.Fatalf("failed to create customer: %v", err)
		}

		found, err := service.GetCustomerByEmail(created.Email)
		if err != nil {
			t.Fatalf("expected to find customer by email, got error: %v", err)
		}

		if found.CustomerID != created.CustomerID {
			t.Errorf("expected ID %d, got %d", created.CustomerID, found.CustomerID)
		}
	})
}

func TestUpdateCustomer(t *testing.T) {
	repository := repositories.NewInMemoryCustomerRepository()
	service := NewInMemoryCustomerService(repository)

	created, err := service.CreateCustomer(getCustomerData())
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	t.Run("fails when ID is invalid", func(t *testing.T) {
		update := created
		update.CustomerID = 0
		_, err := service.UpdateCustomer(update)
		if err == nil {
			t.Fatal("expected error for ID 0, got nil")
		}
	})

	t.Run("fails when customer does not exist", func(t *testing.T) {
		update := created
		update.CustomerID = 9999
		_, err := service.UpdateCustomer(update)
		if err == nil {
			t.Fatal("expected error for non-existent customer, got nil")
		}
		if err.Error() != "customer not found" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("fails on invalid validation", func(t *testing.T) {
		update := created
		update.Email = "invalid-email"
		_, err := service.UpdateCustomer(update)
		if err == nil {
			t.Fatal("expected error for invalid email on update, got nil")
		}
	})

	t.Run("fails when updating to another customer's email", func(t *testing.T) {
		secondCust := models.Customer{
			FirstName:   "Rohan",
			LastName:    "Verma",
			Email:       "rohan.verma@example.com",
			PhoneNumber: "+919123456780",
			DateOfBirth: time.Date(1992, 1, 10, 0, 0, 0, 0, time.UTC),
			Address:     "Delhi, India",
		}
		secondCreated, err := service.CreateCustomer(secondCust)
		if err != nil {
			t.Fatalf("failed to create second customer: %v", err)
		}

		conflictUpdate := secondCreated
		conflictUpdate.Email = created.Email

		_, err = service.UpdateCustomer(conflictUpdate)
		if err == nil {
			t.Fatal("expected error when updating to already taken email, got nil")
		}
		if err.Error() != "email already exists" {
			t.Errorf("expected 'email already exists', got: %v", err)
		}
	})

	t.Run("successfully updates customer and preserves invariants", func(t *testing.T) {
		update := created
		update.FirstName = "Aarav Updated"
		update.LastName = "Sharma Updated"
		update.Address = "Pune, Maharashtra"
		update.KYCStatus = models.KYCStatusVerified
		update.Status = models.CustomerStatusActive

		updated, err := service.UpdateCustomer(update)
		if err != nil {
			t.Fatalf("expected update to succeed, got: %v", err)
		}

		if updated.FirstName != "Aarav Updated" || updated.Address != "Pune, Maharashtra" {
			t.Errorf("fields were not updated properly")
		}

		if updated.CreatedAt != created.CreatedAt {
			t.Errorf("expected CreatedAt to be preserved")
		}
		if updated.KYCStatus != models.KYCStatusNotStarted {
			t.Errorf("expected KYCStatus to be preserved as not_started, got: %s", updated.KYCStatus)
		}
		if updated.Status != models.CustomerStatusPendingVerification {
			t.Errorf("expected Status to be preserved as pending_verification, got: %s", updated.Status)
		}
	})
}
