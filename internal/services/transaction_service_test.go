package services

import (
	"testing"
	"time"

	"banking-system/internal/models"
	"banking-system/internal/repositories"
)

type transactionTestContext struct {
	service       *InMemoryTransactionService
	customerRepo  repositories.CustomerRepository
	accountRepo   repositories.AccountRepository
	txRepo        repositories.TransactionRepository
	customer      models.Customer
	account1      models.Account
	account2      models.Account
	usdAccount    models.Account
	frozenAccount models.Account
}

func setupTransactionTest(t *testing.T) transactionTestContext {
	customerRepo := repositories.NewInMemoryCustomerRepository()
	accountRepo := repositories.NewAccountRepository()
	txRepo := repositories.NewInMemoryTransactionRepository()
	service := NewInMemoryTransactionService(txRepo, accountRepo)

	customer, err := customerRepo.Create(models.Customer{
		FirstName:   "Rahul",
		LastName:    "Verma",
		Email:       "rahul.verma@example.com",
		PhoneNumber: "+919876543210",
		DateOfBirth: time.Date(1992, 4, 10, 0, 0, 0, 0, time.UTC),
		Address:     "Delhi, India",
		Status:      models.CustomerStatusActive,
		KYCStatus:   models.KYCStatusVerified,
	})
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	acc1, err := accountRepo.Create(models.Account{
		AccountNumber: "ACC-1001",
		CustomerID:    customer.CustomerID,
		AccountType:   models.AccountTypeSavings,
		Currency:      models.CurrencyType_INR,
		Balance:       10000,
		Status:        models.AccountStatusActive,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create account1: %v", err)
	}

	acc2, err := accountRepo.Create(models.Account{
		AccountNumber: "ACC-1002",
		CustomerID:    customer.CustomerID,
		AccountType:   models.AccountTypeCurrent,
		Currency:      models.CurrencyType_INR,
		Balance:       5000,
		Status:        models.AccountStatusActive,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create account2: %v", err)
	}

	usdAcc, err := accountRepo.Create(models.Account{
		AccountNumber: "ACC-USD",
		CustomerID:    customer.CustomerID,
		AccountType:   models.AccountTypeSavings,
		Currency:      models.CurrencyType("USD"),
		Balance:       2000,
		Status:        models.AccountStatusActive,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create usdAccount: %v", err)
	}

	frozenAcc, err := accountRepo.Create(models.Account{
		AccountNumber: "ACC-FROZEN",
		CustomerID:    customer.CustomerID,
		AccountType:   models.AccountTypeSavings,
		Currency:      models.CurrencyType_INR,
		Balance:       1000,
		Status:        models.AccountStatusFrozen,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create frozenAccount: %v", err)
	}

	return transactionTestContext{
		service:       service,
		customerRepo:  customerRepo,
		accountRepo:   accountRepo,
		txRepo:        txRepo,
		customer:      customer,
		account1:      acc1,
		account2:      acc2,
		usdAccount:    usdAcc,
		frozenAccount: frozenAcc,
	}
}

func TestDeposit_Success(t *testing.T) {
	ctx := setupTransactionTest(t)

	tx, err := ctx.service.Deposit(ctx.account1.AccountNumber, 2500, "Salary credit")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if tx.Amount != 2500 {
		t.Errorf("expected amount 2500, got: %d", tx.Amount)
	}
	if tx.Type != models.TransactionTypeDeposit {
		t.Errorf("expected transaction type Deposit, got: %s", tx.Type)
	}
	if tx.Status != models.TransactionStatusCompleted {
		t.Errorf("expected status Completed, got: %s", tx.Status)
	}
	if tx.DestinationAccountNumber != ctx.account1.AccountNumber {
		t.Errorf("expected destination account %s, got: %s", ctx.account1.AccountNumber, tx.DestinationAccountNumber)
	}

	// Verify balance updated
	updatedAcc, _ := ctx.accountRepo.FindByAccountNumber(ctx.account1.AccountNumber)
	if updatedAcc.Balance != 12500 {
		t.Errorf("expected updated balance 12500, got: %d", updatedAcc.Balance)
	}
}

func TestDeposit_InvalidAmount(t *testing.T) {
	ctx := setupTransactionTest(t)

	_, err := ctx.service.Deposit(ctx.account1.AccountNumber, 0, "Zero deposit")
	if err == nil {
		t.Fatal("expected error for zero amount, got nil")
	}

	_, err = ctx.service.Deposit(ctx.account1.AccountNumber, -500, "Negative deposit")
	if err == nil {
		t.Fatal("expected error for negative amount, got nil")
	}
}

func TestDeposit_AccountNotFound(t *testing.T) {
	ctx := setupTransactionTest(t)

	_, err := ctx.service.Deposit("non-existent-account", 1000, "Deposit")
	if err == nil {
		t.Fatal("expected error for non-existent account, got nil")
	}
}

func TestDeposit_FrozenAccount(t *testing.T) {
	ctx := setupTransactionTest(t)

	_, err := ctx.service.Deposit(ctx.frozenAccount.AccountNumber, 1000, "Deposit to frozen")
	if err == nil {
		t.Fatal("expected error when depositing to frozen account, got nil")
	}
}

func TestWithdraw_Success(t *testing.T) {
	ctx := setupTransactionTest(t)

	tx, err := ctx.service.Withdraw(ctx.account1.AccountNumber, 3000, "ATM withdrawal")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if tx.Amount != 3000 {
		t.Errorf("expected amount 3000, got: %d", tx.Amount)
	}
	if tx.Type != models.TransactionTypeWithdrawal {
		t.Errorf("expected transaction type Withdrawal, got: %s", tx.Type)
	}
	if tx.SourceAccountNumber != ctx.account1.AccountNumber {
		t.Errorf("expected source account %s, got: %s", ctx.account1.AccountNumber, tx.SourceAccountNumber)
	}

	updatedAcc, _ := ctx.accountRepo.FindByAccountNumber(ctx.account1.AccountNumber)
	if updatedAcc.Balance != 7000 {
		t.Errorf("expected balance 7000, got: %d", updatedAcc.Balance)
	}
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	ctx := setupTransactionTest(t)

	_, err := ctx.service.Withdraw(ctx.account1.AccountNumber, 50000, "Overdraft attempt")
	if err == nil {
		t.Fatal("expected error for insufficient balance, got nil")
	}

	// Verify balance was unchanged
	acc, _ := ctx.accountRepo.FindByAccountNumber(ctx.account1.AccountNumber)
	if acc.Balance != 10000 {
		t.Errorf("expected balance to remain 10000, got: %d", acc.Balance)
	}
}

func TestWithdraw_InvalidAmount(t *testing.T) {
	ctx := setupTransactionTest(t)

	_, err := ctx.service.Withdraw(ctx.account1.AccountNumber, -100, "Negative withdrawal")
	if err == nil {
		t.Fatal("expected error for negative amount, got nil")
	}
}

func TestWithdraw_FrozenAccount(t *testing.T) {
	ctx := setupTransactionTest(t)

	_, err := ctx.service.Withdraw(ctx.frozenAccount.AccountNumber, 500, "Withdraw from frozen")
	if err == nil {
		t.Fatal("expected error when withdrawing from frozen account, got nil")
	}
}

func TestTransfer_Success(t *testing.T) {
	ctx := setupTransactionTest(t)

	tx, err := ctx.service.Transfer(ctx.account1.AccountNumber, ctx.account2.AccountNumber, 4000, "Rent payment")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if tx.Amount != 4000 {
		t.Errorf("expected amount 4000, got: %d", tx.Amount)
	}
	if tx.Type != models.TransactionTypeTransfer {
		t.Errorf("expected transaction type Transfer, got: %s", tx.Type)
	}
	if tx.SourceAccountNumber != ctx.account1.AccountNumber {
		t.Errorf("expected source account %s, got: %s", ctx.account1.AccountNumber, tx.SourceAccountNumber)
	}
	if tx.DestinationAccountNumber != ctx.account2.AccountNumber {
		t.Errorf("expected dest account %s, got: %s", ctx.account2.AccountNumber, tx.DestinationAccountNumber)
	}

	sourceAcc, _ := ctx.accountRepo.FindByAccountNumber(ctx.account1.AccountNumber)
	destAcc, _ := ctx.accountRepo.FindByAccountNumber(ctx.account2.AccountNumber)

	if sourceAcc.Balance != 6000 {
		t.Errorf("expected source balance 6000, got: %d", sourceAcc.Balance)
	}
	if destAcc.Balance != 9000 {
		t.Errorf("expected destination balance 9000, got: %d", destAcc.Balance)
	}
}

func TestTransfer_SameAccount(t *testing.T) {
	ctx := setupTransactionTest(t)

	_, err := ctx.service.Transfer(ctx.account1.AccountNumber, ctx.account1.AccountNumber, 1000, "Self transfer")
	if err == nil {
		t.Fatal("expected error when source and destination accounts are the same, got nil")
	}
}

func TestTransfer_InsufficientBalance(t *testing.T) {
	ctx := setupTransactionTest(t)

	_, err := ctx.service.Transfer(ctx.account1.AccountNumber, ctx.account2.AccountNumber, 25000, "Excess transfer")
	if err == nil {
		t.Fatal("expected error for insufficient balance, got nil")
	}
}

func TestTransfer_CurrencyMismatch(t *testing.T) {
	ctx := setupTransactionTest(t)

	_, err := ctx.service.Transfer(ctx.account1.AccountNumber, ctx.usdAccount.AccountNumber, 1000, "Cross-currency transfer")
	if err == nil {
		t.Fatal("expected error for currency mismatch, got nil")
	}
}

func TestTransfer_FrozenAccount(t *testing.T) {
	ctx := setupTransactionTest(t)

	// Source frozen
	_, err := ctx.service.Transfer(ctx.frozenAccount.AccountNumber, ctx.account1.AccountNumber, 100, "Transfer from frozen")
	if err == nil {
		t.Fatal("expected error when source account is frozen, got nil")
	}

	// Destination frozen
	_, err = ctx.service.Transfer(ctx.account1.AccountNumber, ctx.frozenAccount.AccountNumber, 100, "Transfer to frozen")
	if err == nil {
		t.Fatal("expected error when destination account is frozen, got nil")
	}
}

func TestGetTransactionByID(t *testing.T) {
	ctx := setupTransactionTest(t)

	createdTx, err := ctx.service.Deposit(ctx.account1.AccountNumber, 1500, "Test deposit")
	if err != nil {
		t.Fatalf("failed to deposit: %v", err)
	}

	fetchedTx, err := ctx.service.GetTransactionByID(createdTx.TransactionID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if fetchedTx.TransactionID != createdTx.TransactionID {
		t.Errorf("expected ID %s, got: %s", createdTx.TransactionID, fetchedTx.TransactionID)
	}
}

func TestGetTransactionsByAccountNumber(t *testing.T) {
	ctx := setupTransactionTest(t)

	ctx.service.Deposit(ctx.account1.AccountNumber, 1000, "Deposit 1")
	ctx.service.Withdraw(ctx.account1.AccountNumber, 500, "Withdraw 1")
	ctx.service.Transfer(ctx.account1.AccountNumber, ctx.account2.AccountNumber, 200, "Transfer 1")

	history, err := ctx.service.GetTransactionsByAccountNumber(ctx.account1.AccountNumber)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("expected 3 transactions for account1, got: %d", len(history))
	}

	// account2 should see the transfer as well
	history2, err := ctx.service.GetTransactionsByAccountNumber(ctx.account2.AccountNumber)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(history2) != 1 {
		t.Errorf("expected 1 transaction for account2, got: %d", len(history2))
	}
}
