package repositories

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"banking-system/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository interface {
	Create(transaction models.Transaction) error
	FindByID(transactionID string) (models.Transaction, error)
	FindByAccountNumber(accountNumber string) ([]models.Transaction, error)
}

type InMemoryTransactionRepository struct {
	mutex        sync.RWMutex
	transactions map[string]models.Transaction
}

func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions: make(map[string]models.Transaction),
	}
}

func (repository *InMemoryTransactionRepository) Create(transaction models.Transaction) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if _, exists := repository.transactions[transaction.TransactionID]; exists {
		return errors.New("transaction already exists")
	}

	repository.transactions[transaction.TransactionID] = transaction
	return nil
}

func (repository *InMemoryTransactionRepository) FindByID(transactionID string) (models.Transaction, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	transaction, exists := repository.transactions[transactionID]
	if !exists {
		return models.Transaction{}, errors.New("transaction not found")
	}

	return transaction, nil
}

func (repository *InMemoryTransactionRepository) FindByAccountNumber(accountNumber string) ([]models.Transaction, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	var result []models.Transaction
	for _, tx := range repository.transactions {
		if tx.SourceAccountNumber == accountNumber || tx.DestinationAccountNumber == accountNumber {
			result = append(result, tx)
		}
	}

	return result, nil
}

// PostgresTransactionRepository implements TransactionRepository using PostgreSQL
type PostgresTransactionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTransactionRepository(pool *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{
		pool: pool,
	}
}

func (repository *PostgresTransactionRepository) Create(transaction models.Transaction) error {
	query := `
		INSERT INTO transactions (
			transaction_id, source_account_number, destination_account_number,
			transaction_type, transaction_status, amount, currency, description,
			created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`

	// Store empty strings as NULL in PostgreSQL so Foreign Key constraints are satisfied
	var sourceAcc *string
	if transaction.SourceAccountNumber != "" {
		sourceAcc = &transaction.SourceAccountNumber
	}

	var destAcc *string
	if transaction.DestinationAccountNumber != "" {
		destAcc = &transaction.DestinationAccountNumber
	}

	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now()
	}
	if transaction.CompletedAt.IsZero() {
		transaction.CompletedAt = time.Now()
	}

	_, err := repository.pool.Exec(
		context.Background(),
		query,
		transaction.TransactionID,
		sourceAcc,
		destAcc,
		transaction.Type,
		transaction.Status,
		transaction.Amount,
		transaction.Currency,
		transaction.Description,
		transaction.CreatedAt,
		transaction.CompletedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("transaction already exists")
		}
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (repository *PostgresTransactionRepository) FindByID(transactionID string) (models.Transaction, error) {
	query := `
		SELECT 
			transaction_id, source_account_number, destination_account_number,
			transaction_type, transaction_status, amount, currency, description,
			created_at, completed_at
		FROM transactions
		WHERE transaction_id = $1;
	`

	var tx models.Transaction
	var sourceAcc, destAcc *string

	err := repository.pool.QueryRow(context.Background(), query, transactionID).Scan(
		&tx.TransactionID,
		&sourceAcc,
		&destAcc,
		&tx.Type,
		&tx.Status,
		&tx.Amount,
		&tx.Currency,
		&tx.Description,
		&tx.CreatedAt,
		&tx.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Transaction{}, errors.New("transaction not found")
		}
		return models.Transaction{}, fmt.Errorf("failed to find transaction: %w", err)
	}

	if sourceAcc != nil {
		tx.SourceAccountNumber = *sourceAcc
	}
	if destAcc != nil {
		tx.DestinationAccountNumber = *destAcc
	}

	return tx, nil
}

func (repository *PostgresTransactionRepository) FindByAccountNumber(accountNumber string) ([]models.Transaction, error) {
	query := `
		SELECT 
			transaction_id, source_account_number, destination_account_number,
			transaction_type, transaction_status, amount, currency, description,
			created_at, completed_at
		FROM transactions
		WHERE source_account_number = $1 OR destination_account_number = $1
		ORDER BY created_at DESC;
	`

	rows, err := repository.pool.Query(context.Background(), query, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		var sourceAcc, destAcc *string

		err := rows.Scan(
			&tx.TransactionID,
			&sourceAcc,
			&destAcc,
			&tx.Type,
			&tx.Status,
			&tx.Amount,
			&tx.Currency,
			&tx.Description,
			&tx.CreatedAt,
			&tx.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		if sourceAcc != nil {
			tx.SourceAccountNumber = *sourceAcc
		}
		if destAcc != nil {
			tx.DestinationAccountNumber = *destAcc
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}
