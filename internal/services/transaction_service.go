package services

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"banking-system/internal/models"
	"banking-system/internal/repositories"
	validation "banking-system/internal/validations"
)

type TransactionService interface {
	Deposit(accountNumber string, amount int64, description string) (models.Transaction, error)
	Withdraw(accountNumber string, amount int64, description string) (models.Transaction, error)
	Transfer(sourceAccountNumber, destAccountNumber string, amount int64, description string) (models.Transaction, error)
	GetTransactionByID(transactionID string) (models.Transaction, error)
	GetTransactionsByAccountNumber(accountNumber string) ([]models.Transaction, error)
}

type InMemoryTransactionService struct {
	transactionRepository repositories.TransactionRepository
	accountRepository     repositories.AccountRepository
}

func NewInMemoryTransactionService(
	txRepo repositories.TransactionRepository,
	accountRepo repositories.AccountRepository,
) *InMemoryTransactionService {
	return &InMemoryTransactionService{
		transactionRepository: txRepo,
		accountRepository:     accountRepo,
	}
}

var txCounter uint64

func generateTransactionID() string {
	id := atomic.AddUint64(&txCounter, 1)
	return fmt.Sprintf("TXN-%d-%d", time.Now().UnixNano(), id)
}

func (s *InMemoryTransactionService) Deposit(accountNumber string, amount int64, description string) (models.Transaction, error) {
	if !validation.IsValidAmount(amount) {
		return models.Transaction{}, errors.New("invalid deposit amount, amount must be greater than zero")
	}

	account, err := s.accountRepository.FindByAccountNumber(accountNumber)
	if err != nil {
		return models.Transaction{}, errors.New("account not found")
	}

	if account.Status != models.AccountStatusActive {
		return models.Transaction{}, errors.New("cannot deposit to inactive or closed account")
	}

	account.Balance += amount
	_, err = s.accountRepository.Update(account)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to update account balance: %w", err)
	}

	now := time.Now()
	tx := models.Transaction{
		TransactionID:            generateTransactionID(),
		DestinationAccountNumber: accountNumber,
		Type:                     models.TransactionTypeDeposit,
		Status:                   models.TransactionStatusCompleted,
		Amount:                   amount,
		Currency:                 account.Currency,
		Description:              description,
		CreatedAt:                now,
		CompletedAt:              now,
	}

	if err := s.transactionRepository.Create(tx); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to record transaction: %w", err)
	}

	return tx, nil
}

func (s *InMemoryTransactionService) Withdraw(accountNumber string, amount int64, description string) (models.Transaction, error) {
	if !validation.IsValidAmount(amount) {
		return models.Transaction{}, errors.New("invalid withdrawal amount, amount must be greater than zero")
	}

	account, err := s.accountRepository.FindByAccountNumber(accountNumber)
	if err != nil {
		return models.Transaction{}, errors.New("account not found")
	}

	if account.Status != models.AccountStatusActive {
		return models.Transaction{}, errors.New("cannot withdraw from inactive or closed account")
	}

	if account.Balance < amount {
		return models.Transaction{}, errors.New("insufficient balance")
	}

	account.Balance -= amount
	_, err = s.accountRepository.Update(account)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to update account balance: %w", err)
	}

	now := time.Now()
	tx := models.Transaction{
		TransactionID:       generateTransactionID(),
		SourceAccountNumber: accountNumber,
		Type:                models.TransactionTypeWithdrawal,
		Status:              models.TransactionStatusCompleted,
		Amount:              amount,
		Currency:            account.Currency,
		Description:         description,
		CreatedAt:           now,
		CompletedAt:         now,
	}

	if err := s.transactionRepository.Create(tx); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to record transaction: %w", err)
	}

	return tx, nil
}

func (s *InMemoryTransactionService) Transfer(sourceAccountNumber, destAccountNumber string, amount int64, description string) (models.Transaction, error) {
	if !validation.IsValidTransferAccounts(sourceAccountNumber, destAccountNumber) {
		return models.Transaction{}, errors.New("invalid transfer accounts: source and destination must not be empty and cannot be the same")
	}

	if !validation.IsValidAmount(amount) {
		return models.Transaction{}, errors.New("invalid transfer amount, amount must be greater than zero")
	}

	sourceAccount, err := s.accountRepository.FindByAccountNumber(sourceAccountNumber)
	if err != nil {
		return models.Transaction{}, errors.New("source account not found")
	}

	destAccount, err := s.accountRepository.FindByAccountNumber(destAccountNumber)
	if err != nil {
		return models.Transaction{}, errors.New("destination account not found")
	}

	if sourceAccount.Status != models.AccountStatusActive {
		return models.Transaction{}, errors.New("source account is inactive or closed")
	}

	if destAccount.Status != models.AccountStatusActive {
		return models.Transaction{}, errors.New("destination account is inactive or closed")
	}

	if sourceAccount.Currency != destAccount.Currency {
		return models.Transaction{}, errors.New("currency mismatch between accounts")
	}

	if sourceAccount.Balance < amount {
		return models.Transaction{}, errors.New("insufficient balance in source account")
	}

	sourceAccount.Balance -= amount
	destAccount.Balance += amount

	if _, err := s.accountRepository.Update(sourceAccount); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to debit source account: %w", err)
	}

	if _, err := s.accountRepository.Update(destAccount); err != nil {
		// Rollback source debit if dest credit fails
		sourceAccount.Balance += amount
		_, _ = s.accountRepository.Update(sourceAccount)
		return models.Transaction{}, fmt.Errorf("failed to credit destination account: %w", err)
	}

	now := time.Now()
	tx := models.Transaction{
		TransactionID:            generateTransactionID(),
		SourceAccountNumber:      sourceAccountNumber,
		DestinationAccountNumber: destAccountNumber,
		Type:                     models.TransactionTypeTransfer,
		Status:                   models.TransactionStatusCompleted,
		Amount:                   amount,
		Currency:                 sourceAccount.Currency,
		Description:              description,
		CreatedAt:                now,
		CompletedAt:              now,
	}

	if err := s.transactionRepository.Create(tx); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to record transaction: %w", err)
	}

	return tx, nil
}

func (s *InMemoryTransactionService) GetTransactionByID(transactionID string) (models.Transaction, error) {
	return s.transactionRepository.FindByID(transactionID)
}

func (s *InMemoryTransactionService) GetTransactionsByAccountNumber(accountNumber string) ([]models.Transaction, error) {
	if _, err := s.accountRepository.FindByAccountNumber(accountNumber); err != nil {
		return nil, errors.New("account not found")
	}

	return s.transactionRepository.FindByAccountNumber(accountNumber)
}
