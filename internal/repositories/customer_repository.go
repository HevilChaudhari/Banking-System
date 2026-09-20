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

type CustomerRepository interface {
	Create(customer models.Customer) (models.Customer, error)
	FindByID(customerID int) (models.Customer, error)
	FindByEmail(email string) (models.Customer, error)
	Update(customer models.Customer) (models.Customer, error)
}

type InMemoryCustomerRepository struct {
	mutex          sync.RWMutex
	customers      map[int]models.Customer
	nextCustomerID int
}

type PostgresCustomerRepository struct {
	pool *pgxpool.Pool
}

func NewInMemoryCustomerRepository() *InMemoryCustomerRepository {
	return &InMemoryCustomerRepository{
		customers:      make(map[int]models.Customer),
		nextCustomerID: 0,
	}
}

func NewPostgresCustomerRepository(pool *pgxpool.Pool) *PostgresCustomerRepository {
	return &PostgresCustomerRepository{
		pool: pool,
	}
}

func (repository *InMemoryCustomerRepository) Create(customer models.Customer) (models.Customer, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	for _, cus := range repository.customers {
		if cus.Email == customer.Email {
			return models.Customer{}, errors.New("customer already exist")
		}
	}

	repository.nextCustomerID++

	customer.CustomerID = repository.nextCustomerID
	repository.customers[customer.CustomerID] = customer

	return customer, nil
}

func (repository *InMemoryCustomerRepository) FindByID(customerID int) (models.Customer, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	customer, exist := repository.customers[customerID]

	if !exist {
		return models.Customer{}, errors.New("customer not found")
	}

	return customer, nil

}

func (repository *InMemoryCustomerRepository) FindByEmail(email string) (models.Customer, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	for _, customer := range repository.customers {
		if customer.Email == email {
			return customer, nil
		}
	}

	return models.Customer{}, errors.New("customer not found")
}

func (repository *InMemoryCustomerRepository) Update(customer models.Customer) (models.Customer, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	existingCustomer, exists := repository.customers[customer.CustomerID]
	if !exists {
		return models.Customer{}, errors.New("customer not found")
	}

	for _, otherCustomer := range repository.customers {
		if otherCustomer.Email == customer.Email &&
			otherCustomer.CustomerID != customer.CustomerID {
			return models.Customer{}, errors.New("email already exists")
		}
	}

	customer.CreatedAt = existingCustomer.CreatedAt
	repository.customers[customer.CustomerID] = customer

	return customer, nil
}

func (repo *PostgresCustomerRepository) Create(customer models.Customer) (models.Customer, error) {
	query := `
		INSERT INTO customers (
			first_name, last_name, email, phone_number,
			date_of_birth, address, status, kyc_status, password_hash, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING customer_id, created_at;
	`

	now := time.Now()
	err := repo.pool.QueryRow(
		context.Background(),
		query,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.PhoneNumber,
		customer.DateOfBirth,
		customer.Address,
		customer.Status,
		customer.KYCStatus,
		customer.PasswordHash,
		now,
	).Scan(&customer.CustomerID, &customer.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.Customer{}, errors.New("customer with this email already exists")
		}
		return models.Customer{}, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

func (repo *PostgresCustomerRepository) FindByID(customerID int) (models.Customer, error) {
	query := `
		SELECT 
			customer_id, first_name, last_name, email, phone_number,
			date_of_birth, address, status, kyc_status, password_hash, created_at
		FROM customers
		WHERE customer_id = $1;
	`

	var customer models.Customer
	err := repo.pool.QueryRow(context.Background(), query, customerID).Scan(
		&customer.CustomerID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.PhoneNumber,
		&customer.DateOfBirth,
		&customer.Address,
		&customer.Status,
		&customer.KYCStatus,
		&customer.PasswordHash,
		&customer.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Customer{}, errors.New("customer not found")
		}
		return models.Customer{}, fmt.Errorf("failed to find customer by id: %w", err)
	}

	return customer, nil
}

func (repo *PostgresCustomerRepository) FindByEmail(email string) (models.Customer, error) {
	query := `
		SELECT 
			customer_id, first_name, last_name, email, phone_number,
			date_of_birth, address, status, kyc_status, password_hash, created_at
		FROM customers
		WHERE email = $1;
	`

	var customer models.Customer
	err := repo.pool.QueryRow(context.Background(), query, email).Scan(
		&customer.CustomerID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.PhoneNumber,
		&customer.DateOfBirth,
		&customer.Address,
		&customer.Status,
		&customer.KYCStatus,
		&customer.PasswordHash,
		&customer.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Customer{}, errors.New("customer not found")
		}
		return models.Customer{}, fmt.Errorf("failed to find customer by email: %w", err)
	}

	return customer, nil
}

func (repo *PostgresCustomerRepository) Update(customer models.Customer) (models.Customer, error) {
	query := `
		UPDATE customers
		SET 
			first_name = $1,
			last_name = $2,
			email = $3,
			phone_number = $4,
			date_of_birth = $5,
			address = $6,
			status = $7,
			kyc_status = $8
		WHERE customer_id = $9;
	`

	cmdTag, err := repo.pool.Exec(
		context.Background(),
		query,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.PhoneNumber,
		customer.DateOfBirth,
		customer.Address,
		customer.Status,
		customer.KYCStatus,
		customer.CustomerID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.Customer{}, errors.New("email already in use by another customer")
		}
		return models.Customer{}, fmt.Errorf("failed to update customer: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return models.Customer{}, errors.New("customer not found")
	}

	return customer, nil
}
