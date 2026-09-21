package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"banking-system/internal/models"
	"banking-system/internal/repositories"
	validation "banking-system/internal/validations"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTransactionService implements TransactionService using PostgreSQL ACID transactions
// and row-level pessimistic locking (FOR UPDATE) for concurrent-safe balance updates.
type PostgresTransactionService struct {
	pool   *pgxpool.Pool
	txRepo repositories.TransactionRepository
}

func NewPostgresTransactionService(
	pool *pgxpool.Pool,
	txRepo repositories.TransactionRepository,
) *PostgresTransactionService {
	return &PostgresTransactionService{
		pool:   pool,
		txRepo: txRepo,
	}
}

func (s *PostgresTransactionService) Deposit(accountNumber string, amount int64, description string) (models.Transaction, error) {
	if !validation.IsValidAmount(amount) {
		return models.Transaction{}, errors.New("invalid deposit amount, amount must be greater than zero")
	}

	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Acquire exclusive row lock on the account
	selectQuery := `
		SELECT balance, currency, status 
		FROM accounts 
		WHERE account_number = $1 
		FOR UPDATE;
	`
	var currentBalance int64
	var currency models.CurrencyType
	var status models.AccountStatus

	err = tx.QueryRow(ctx, selectQuery, accountNumber).Scan(&currentBalance, &currency, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Transaction{}, errors.New("account not found")
		}
		return models.Transaction{}, fmt.Errorf("failed to fetch account: %w", err)
	}

	if status != models.AccountStatusActive {
		return models.Transaction{}, errors.New("cannot deposit to inactive or closed account")
	}

	newBalance := currentBalance + amount
	updateQuery := `
		UPDATE accounts 
		SET balance = $1 
		WHERE account_number = $2;
	`
	_, err = tx.Exec(ctx, updateQuery, newBalance, accountNumber)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to update account balance: %w", err)
	}

	now := time.Now()
	txn := models.Transaction{
		TransactionID:            generateTransactionID(),
		DestinationAccountNumber: accountNumber,
		Type:                     models.TransactionTypeDeposit,
		Status:                   models.TransactionStatusCompleted,
		Amount:                   amount,
		Currency:                 currency,
		Description:              description,
		CreatedAt:                now,
		CompletedAt:              now,
	}

	insertTxQuery := `
		INSERT INTO transactions (
			transaction_id, source_account_number, destination_account_number,
			transaction_type, transaction_status, amount, currency, description,
			created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`
	var destAcc *string = &accountNumber
	_, err = tx.Exec(
		ctx,
		insertTxQuery,
		txn.TransactionID,
		nil, // source_account_number is NULL for deposits
		destAcc,
		txn.Type,
		txn.Status,
		txn.Amount,
		txn.Currency,
		txn.Description,
		txn.CreatedAt,
		txn.CompletedAt,
	)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to record transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return txn, nil
}

func (s *PostgresTransactionService) Withdraw(accountNumber string, amount int64, description string) (models.Transaction, error) {
	if !validation.IsValidAmount(amount) {
		return models.Transaction{}, errors.New("invalid withdrawal amount, amount must be greater than zero")
	}

	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Acquire exclusive row lock on the account
	selectQuery := `
		SELECT balance, currency, status 
		FROM accounts 
		WHERE account_number = $1 
		FOR UPDATE;
	`
	var currentBalance int64
	var currency models.CurrencyType
	var status models.AccountStatus

	err = tx.QueryRow(ctx, selectQuery, accountNumber).Scan(&currentBalance, &currency, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Transaction{}, errors.New("account not found")
		}
		return models.Transaction{}, fmt.Errorf("failed to fetch account: %w", err)
	}

	if status != models.AccountStatusActive {
		return models.Transaction{}, errors.New("cannot withdraw from inactive or closed account")
	}

	if currentBalance < amount {
		return models.Transaction{}, errors.New("insufficient balance")
	}

	newBalance := currentBalance - amount
	updateQuery := `
		UPDATE accounts 
		SET balance = $1 
		WHERE account_number = $2;
	`
	_, err = tx.Exec(ctx, updateQuery, newBalance, accountNumber)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to update account balance: %w", err)
	}

	now := time.Now()
	txn := models.Transaction{
		TransactionID:       generateTransactionID(),
		SourceAccountNumber: accountNumber,
		Type:                models.TransactionTypeWithdrawal,
		Status:              models.TransactionStatusCompleted,
		Amount:              amount,
		Currency:            currency,
		Description:         description,
		CreatedAt:           now,
		CompletedAt:         now,
	}

	insertTxQuery := `
		INSERT INTO transactions (
			transaction_id, source_account_number, destination_account_number,
			transaction_type, transaction_status, amount, currency, description,
			created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`
	var sourceAcc *string = &accountNumber
	_, err = tx.Exec(
		ctx,
		insertTxQuery,
		txn.TransactionID,
		sourceAcc,
		nil, // destination_account_number is NULL for withdrawals
		txn.Type,
		txn.Status,
		txn.Amount,
		txn.Currency,
		txn.Description,
		txn.CreatedAt,
		txn.CompletedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return models.Transaction{}, errors.New("insufficient balance")
		}
		return models.Transaction{}, fmt.Errorf("failed to record transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return txn, nil
}

func (s *PostgresTransactionService) Transfer(sourceAccountNumber, destAccountNumber string, amount int64, description string) (models.Transaction, error) {
	if !validation.IsValidTransferAccounts(sourceAccountNumber, destAccountNumber) {
		return models.Transaction{}, errors.New("invalid transfer accounts: source and destination must not be empty and cannot be the same")
	}

	if !validation.IsValidAmount(amount) {
		return models.Transaction{}, errors.New("invalid transfer amount, amount must be greater than zero")
	}

	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Deadlock Prevention:
	// Always lock rows in deterministic (sorted) order by account_number.
	// Using ORDER BY account_number FOR UPDATE ensures PostgreSQL locks rows in the exact same order
	// regardless of whether transfer is A -> B or B -> A.
	selectQuery := `
		SELECT account_number, balance, currency, status 
		FROM accounts 
		WHERE account_number IN ($1, $2) 
		ORDER BY account_number 
		FOR UPDATE;
	`
	rows, err := tx.Query(ctx, selectQuery, sourceAccountNumber, destAccountNumber)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	type accountSnapshot struct {
		balance  int64
		currency models.CurrencyType
		status   models.AccountStatus
	}
	accountsFound := make(map[string]accountSnapshot)

	for rows.Next() {
		var accNum string
		var snap accountSnapshot
		if err := rows.Scan(&accNum, &snap.balance, &snap.currency, &snap.status); err != nil {
			return models.Transaction{}, fmt.Errorf("failed to scan account: %w", err)
		}
		accountsFound[accNum] = snap
	}
	if err := rows.Err(); err != nil {
		return models.Transaction{}, fmt.Errorf("error iterating accounts: %w", err)
	}

	sourceSnap, sourceExists := accountsFound[sourceAccountNumber]
	if !sourceExists {
		return models.Transaction{}, errors.New("source account not found")
	}

	destSnap, destExists := accountsFound[destAccountNumber]
	if !destExists {
		return models.Transaction{}, errors.New("destination account not found")
	}

	if sourceSnap.status != models.AccountStatusActive {
		return models.Transaction{}, errors.New("source account is inactive or closed")
	}

	if destSnap.status != models.AccountStatusActive {
		return models.Transaction{}, errors.New("destination account is inactive or closed")
	}

	if sourceSnap.currency != destSnap.currency {
		return models.Transaction{}, errors.New("currency mismatch between accounts")
	}

	if sourceSnap.balance < amount {
		return models.Transaction{}, errors.New("insufficient balance in source account")
	}

	// Debit source account
	updateQuery := `UPDATE accounts SET balance = $1 WHERE account_number = $2;`
	if _, err := tx.Exec(ctx, updateQuery, sourceSnap.balance-amount, sourceAccountNumber); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to debit source account: %w", err)
	}

	// Credit destination account
	if _, err := tx.Exec(ctx, updateQuery, destSnap.balance+amount, destAccountNumber); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to credit destination account: %w", err)
	}

	now := time.Now()
	txn := models.Transaction{
		TransactionID:            generateTransactionID(),
		SourceAccountNumber:      sourceAccountNumber,
		DestinationAccountNumber: destAccountNumber,
		Type:                     models.TransactionTypeTransfer,
		Status:                   models.TransactionStatusCompleted,
		Amount:                   amount,
		Currency:                 sourceSnap.currency,
		Description:              description,
		CreatedAt:                now,
		CompletedAt:              now,
	}

	insertTxQuery := `
		INSERT INTO transactions (
			transaction_id, source_account_number, destination_account_number,
			transaction_type, transaction_status, amount, currency, description,
			created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`
	sourceAcc := &sourceAccountNumber
	destAcc := &destAccountNumber
	_, err = tx.Exec(
		ctx,
		insertTxQuery,
		txn.TransactionID,
		sourceAcc,
		destAcc,
		txn.Type,
		txn.Status,
		txn.Amount,
		txn.Currency,
		txn.Description,
		txn.CreatedAt,
		txn.CompletedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return models.Transaction{}, errors.New("insufficient balance in source account")
		}
		return models.Transaction{}, fmt.Errorf("failed to record transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return txn, nil
}

func (s *PostgresTransactionService) GetTransactionByID(transactionID string) (models.Transaction, error) {
	return s.txRepo.FindByID(transactionID)
}

func (s *PostgresTransactionService) GetTransactionsByAccountNumber(accountNumber string) ([]models.Transaction, error) {
	return s.txRepo.FindByAccountNumber(accountNumber)
}
