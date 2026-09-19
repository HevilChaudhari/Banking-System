package services

import (
	"testing"
	"time"

	"banking-system/internal/models"
	"banking-system/internal/repositories"
)

func setupAccountTest(t *testing.T) (*InMemoryAccountService, repositories.CustomerRepository, repositories.AccountRepository, models.Customer) {
	customerRepo := repositories.NewInMemoryCustomerRepository()
	accountRepo := repositories.NewAccountRepository()
	service := NewInMemoryAccountService(accountRepo, customerRepo)

	customer, err := customerRepo.Create(models.Customer{
		FirstName:   "Aarav",
		LastName:    "Sharma",
		Email:       "aarav.sharma@example.com",
		PhoneNumber: "+919876543210",
		DateOfBirth: time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC),
		Address:     "Mumbai, Maharashtra",
		Status:      models.CustomerStatusActive,
		KYCStatus:   models.KYCStatusVerified,
	})
	if err != nil {
		t.Fatalf("failed to create test customer: %v", err)
	}

	return service, customerRepo, accountRepo, customer
}

func TestCreateAccount_Success(t *testing.T) {
	service, _, _, customer := setupAccountTest(t)

	account := models.Account{
		CustomerID:  customer.CustomerID,
		AccountType: models.AccountTypeSavings,
		Currency:    models.CurrencyType_INR,
		Balance:     5000,
	}

	created, err := service.CreateAccount(account)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(created.AccountNumber) < 10 {
		t.Errorf("expected account number length >= 10, got: %s", created.AccountNumber)
	}

	if created.CustomerID != customer.CustomerID {
		t.Errorf("expected customer ID %d, got: %d", customer.CustomerID, created.CustomerID)
	}

	if created.AccountType != models.AccountTypeSavings {
		t.Errorf("expected account type savings, got: %s", created.AccountType)
	}

	if created.Currency != models.CurrencyType_INR {
		t.Errorf("expected currency INR, got: %s", created.Currency)
	}

	if created.Balance != 5000 {
		t.Errorf("expected balance 5000, got: %d", created.Balance)
	}

	if created.Status != models.AccountStatusActive {
		t.Errorf("expected status active, got: %s", created.Status)
	}

	if created.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}
}

func TestCreateAccount_CustomerErrors(t *testing.T) {
	service, customerRepo, _, _ := setupAccountTest(t)

	t.Run("fails when customer does not exist", func(t *testing.T) {
		acc := models.Account{
			CustomerID:  9999,
			AccountType: models.AccountTypeSavings,
			Currency:    models.CurrencyType_INR,
			Balance:     1000,
		}

		_, err := service.CreateAccount(acc)
		if err == nil {
			t.Fatal("expected error for non-existent customer, got nil")
		}
		if err.Error() != "customer not found" {
			t.Errorf("expected 'customer not found', got: %v", err)
		}
	})

	t.Run("fails when customer is blocked", func(t *testing.T) {
		blockedCustomer, err := customerRepo.Create(models.Customer{
			FirstName:   "Blocked",
			LastName:    "User",
			Email:       "blocked@example.com",
			PhoneNumber: "+919876543211",
			DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:      models.CustomerStatusBlocked,
		})
		if err != nil {
			t.Fatalf("failed to create blocked customer: %v", err)
		}

		acc := models.Account{
			CustomerID:  blockedCustomer.CustomerID,
			AccountType: models.AccountTypeSavings,
			Currency:    models.CurrencyType_INR,
			Balance:     1000,
		}

		_, err = service.CreateAccount(acc)
		if err == nil {
			t.Fatal("expected error for blocked customer, got nil")
		}
		if err.Error() != "customer is either blocked or closed" {
			t.Errorf("expected 'customer is either blocked or closed', got: %v", err)
		}
	})

	t.Run("fails when customer is closed", func(t *testing.T) {
		closedCustomer, err := customerRepo.Create(models.Customer{
			FirstName:   "Closed",
			LastName:    "User",
			Email:       "closed@example.com",
			PhoneNumber: "+919876543212",
			DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:      models.CustomerStatusClosed,
		})
		if err != nil {
			t.Fatalf("failed to create closed customer: %v", err)
		}

		acc := models.Account{
			CustomerID:  closedCustomer.CustomerID,
			AccountType: models.AccountTypeCurrent,
			Currency:    models.CurrencyType_INR,
			Balance:     1000,
		}

		_, err = service.CreateAccount(acc)
		if err == nil {
			t.Fatal("expected error for closed customer, got nil")
		}
		if err.Error() != "customer is either blocked or closed" {
			t.Errorf("expected 'customer is either blocked or closed', got: %v", err)
		}
	})
}

func TestCreateAccount_ValidationErrors(t *testing.T) {
	service, _, _, customer := setupAccountTest(t)

	testCases := []struct {
		name        string
		modify      func(a *models.Account)
		expectedErr string
	}{
		{
			name: "invalid account type",
			modify: func(a *models.Account) {
				a.AccountType = "fixed_deposit"
			},
			expectedErr: "account type is invalid",
		},
		{
			name: "invalid currency",
			modify: func(a *models.Account) {
				a.Currency = "USD"
			},
			expectedErr: "account currency is invalid",
		},
		{
			name: "negative balance",
			modify: func(a *models.Account) {
				a.Balance = -1
			},
			expectedErr: "account balance is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acc := models.Account{
				CustomerID:  customer.CustomerID,
				AccountType: models.AccountTypeSavings,
				Currency:    models.CurrencyType_INR,
				Balance:     0,
			}
			tc.modify(&acc)

			_, err := service.CreateAccount(acc)
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.expectedErr)
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("expected error %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestGetAccountByNumber(t *testing.T) {
	service, _, _, customer := setupAccountTest(t)

	t.Run("fails when account number is empty", func(t *testing.T) {
		_, err := service.GetAccountByNumber("   ")
		if err == nil {
			t.Fatal("expected error for empty account number, got nil")
		}
		if err.Error() != "account number cannot be empty" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("fails when account number length is less than 10", func(t *testing.T) {
		_, err := service.GetAccountByNumber("12345")
		if err == nil {
			t.Fatal("expected error for short account number, got nil")
		}
		if err.Error() != "account number is invalid" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("fails when account is not found", func(t *testing.T) {
		_, err := service.GetAccountByNumber("999999999999")
		if err == nil {
			t.Fatal("expected error for non-existent account, got nil")
		}
		if err.Error() != "account not found" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("successfully retrieves account", func(t *testing.T) {
		created, err := service.CreateAccount(models.Account{
			CustomerID:  customer.CustomerID,
			AccountType: models.AccountTypeSavings,
			Currency:    models.CurrencyType_INR,
			Balance:     2500,
		})
		if err != nil {
			t.Fatalf("failed to create account: %v", err)
		}

		found, err := service.GetAccountByNumber(created.AccountNumber)
		if err != nil {
			t.Fatalf("expected to find account, got error: %v", err)
		}

		if found.AccountNumber != created.AccountNumber {
			t.Errorf("expected account number %s, got %s", created.AccountNumber, found.AccountNumber)
		}
		if found.Balance != 2500 {
			t.Errorf("expected balance 2500, got %d", found.Balance)
		}
	})
}

func TestGetAccountsByCustomerID(t *testing.T) {
	service, _, _, customer := setupAccountTest(t)

	t.Run("fails when customer ID is invalid", func(t *testing.T) {
		_, err := service.GetAccountsByCustomerID(0)
		if err == nil {
			t.Fatal("expected error for customer ID 0, got nil")
		}
		if err.Error() != "customerID is invalid" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("fails when customer is not found", func(t *testing.T) {
		_, err := service.GetAccountsByCustomerID(9999)
		if err == nil {
			t.Fatal("expected error for non-existent customer, got nil")
		}
		if err.Error() != "customer not found" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("returns empty slice when customer has no accounts", func(t *testing.T) {
		accounts, err := service.GetAccountsByCustomerID(customer.CustomerID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(accounts) != 0 {
			t.Errorf("expected 0 accounts, got %d", len(accounts))
		}
	})

	t.Run("returns all accounts for the customer", func(t *testing.T) {
		// Create 2 accounts for customer
		_, err := service.CreateAccount(models.Account{
			CustomerID:  customer.CustomerID,
			AccountType: models.AccountTypeSavings,
			Currency:    models.CurrencyType_INR,
			Balance:     1000,
		})
		if err != nil {
			t.Fatalf("failed to create savings account: %v", err)
		}

		time.Sleep(5 * time.Millisecond)

		_, err = service.CreateAccount(models.Account{
			CustomerID:  customer.CustomerID,
			AccountType: models.AccountTypeCurrent,
			Currency:    models.CurrencyType_INR,
			Balance:     5000,
		})
		if err != nil {
			t.Fatalf("failed to create current account: %v", err)
		}

		accounts, err := service.GetAccountsByCustomerID(customer.CustomerID)
		if err != nil {
			t.Fatalf("failed to get accounts: %v", err)
		}

		if len(accounts) != 2 {
			t.Errorf("expected 2 accounts, got %d", len(accounts))
		}
	})
}

func TestUpdateAccountStatus(t *testing.T) {
	service, _, _, customer := setupAccountTest(t)

	created, err := service.CreateAccount(models.Account{
		CustomerID:  customer.CustomerID,
		AccountType: models.AccountTypeSavings,
		Currency:    models.CurrencyType_INR,
		Balance:     1000,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	t.Run("fails when account number is empty", func(t *testing.T) {
		_, err := service.UpdateAccountStatus("  ", models.AccountStatusFrozen)
		if err == nil {
			t.Fatal("expected error for empty account number, got nil")
		}
		if err.Error() != "account number cannot be empty" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("fails when status is invalid", func(t *testing.T) {
		_, err := service.UpdateAccountStatus(created.AccountNumber, "unknown_status")
		if err == nil {
			t.Fatal("expected error for invalid status, got nil")
		}
		if err.Error() != "status is invalid" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("fails when account is not found", func(t *testing.T) {
		_, err := service.UpdateAccountStatus("999999999999", models.AccountStatusFrozen)
		if err == nil {
			t.Fatal("expected error for non-existent account, got nil")
		}
		if err.Error() != "account not found" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("successfully freezes account", func(t *testing.T) {
		updated, err := service.UpdateAccountStatus(created.AccountNumber, models.AccountStatusFrozen)
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if updated.Status != models.AccountStatusFrozen {
			t.Errorf("expected status frozen, got: %s", updated.Status)
		}
	})

	t.Run("successfully closes account", func(t *testing.T) {
		updated, err := service.UpdateAccountStatus(created.AccountNumber, models.AccountStatusClosed)
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if updated.Status != models.AccountStatusClosed {
			t.Errorf("expected status closed, got: %s", updated.Status)
		}
	})

	t.Run("fails to update already closed account", func(t *testing.T) {
		_, err := service.UpdateAccountStatus(created.AccountNumber, models.AccountStatusActive)
		if err == nil {
			t.Fatal("expected error when updating closed account, got nil")
		}
		if err.Error() != "cannot change status of a closed account" {
			t.Errorf("expected 'cannot change status of a closed account', got: %v", err)
		}
	})
}
