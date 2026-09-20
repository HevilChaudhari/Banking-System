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

type AccountRepository interface {
	Create(account models.Account) (models.Account, error)
	FindByAccountNumber(accountNumber string) (models.Account, error)
	FindByCustomerID(customerID int) ([]models.Account, error)
	Update(account models.Account) (models.Account, error)
}

type InMemoryAccountRepository struct {
	mutex    sync.RWMutex
	accounts map[string]models.Account
}

func NewAccountRepository() *InMemoryAccountRepository {
	return &InMemoryAccountRepository{
		accounts: make(map[string]models.Account),
	}
}

func (repository *InMemoryAccountRepository) Create(account models.Account) (models.Account, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if _, exist := repository.accounts[account.AccountNumber]; exist {
		return models.Account{}, errors.New("account already exist")
	}

	repository.accounts[account.AccountNumber] = account

	return account, nil
}

func (repository *InMemoryAccountRepository) FindByAccountNumber(accountNumber string) (models.Account, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	account, exist := repository.accounts[accountNumber]

	if !exist {
		return models.Account{}, errors.New("account not found")
	}

	return account, nil
}

func (repository *InMemoryAccountRepository) FindByCustomerID(customerID int) ([]models.Account, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	var accounts []models.Account
	for _, account := range repository.accounts {
		if account.CustomerID == customerID {
			accounts = append(accounts, account)
		}
	}

	return accounts, nil
}

func (repository *InMemoryAccountRepository) Update(account models.Account) (models.Account, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if _, exist := repository.accounts[account.AccountNumber]; !exist {
		return models.Account{}, errors.New("account not found")
	}

	repository.accounts[account.AccountNumber] = account

	return account, nil
}

// PostgresAccountRepository implements AccountRepository using PostgreSQL
type PostgresAccountRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAccountRepository(pool *pgxpool.Pool) *PostgresAccountRepository {
	return &PostgresAccountRepository{
		pool: pool,
	}
}

func (repository *PostgresAccountRepository) Create(account models.Account) (models.Account, error) {
	query := `
		INSERT INTO accounts (
			account_number, customer_id, account_type, balance, currency, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at;
	`

	if account.CreatedAt.IsZero() {
		account.CreatedAt = time.Now()
	}

	err := repository.pool.QueryRow(
		context.Background(),
		query,
		account.AccountNumber,
		account.CustomerID,
		account.AccountType,
		account.Balance,
		account.Currency,
		account.Status,
		account.CreatedAt,
	).Scan(&account.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return models.Account{}, errors.New("account already exist")
			}
			if pgErr.Code == "23503" {
				return models.Account{}, errors.New("customer not found")
			}
		}
		return models.Account{}, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

func (repository *PostgresAccountRepository) FindByAccountNumber(accountNumber string) (models.Account, error) {
	query := `
		SELECT account_number, customer_id, account_type, balance, currency, status, created_at
		FROM accounts
		WHERE account_number = $1;
	`

	var account models.Account
	err := repository.pool.QueryRow(context.Background(), query, accountNumber).Scan(
		&account.AccountNumber,
		&account.CustomerID,
		&account.AccountType,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, errors.New("account not found")
		}
		return models.Account{}, fmt.Errorf("failed to find account: %w", err)
	}

	return account, nil
}

func (repository *PostgresAccountRepository) FindByCustomerID(customerID int) ([]models.Account, error) {
	query := `
		SELECT account_number, customer_id, account_type, balance, currency, status, created_at
		FROM accounts
		WHERE customer_id = $1;
	`

	rows, err := repository.pool.Query(context.Background(), query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.AccountNumber,
			&account.CustomerID,
			&account.AccountType,
			&account.Balance,
			&account.Currency,
			&account.Status,
			&account.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return accounts, nil
}

func (repository *PostgresAccountRepository) Update(account models.Account) (models.Account, error) {
	query := `
		UPDATE accounts
		SET balance = $1, status = $2
		WHERE account_number = $3;
	`

	cmdTag, err := repository.pool.Exec(
		context.Background(),
		query,
		account.Balance,
		account.Status,
		account.AccountNumber,
	)

	if err != nil {
		return models.Account{}, fmt.Errorf("failed to update account: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return models.Account{}, errors.New("account not found")
	}

	return account, nil
}
