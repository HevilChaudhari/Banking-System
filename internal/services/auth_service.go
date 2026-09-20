package services

import (
	"errors"
	"strings"

	"banking-system/internal/auth"
	"banking-system/internal/models"
	"banking-system/internal/repositories"
)

type AuthService interface {
	Register(customer models.Customer, password string) (models.Customer, string, error)
	Login(email, password string) (models.Customer, string, error)
}

type DefaultAuthService struct {
	customerService    CustomerService
	customerRepository repositories.CustomerRepository
}

func NewAuthService(custService CustomerService, custRepo repositories.CustomerRepository) *DefaultAuthService {
	return &DefaultAuthService{
		customerService:    custService,
		customerRepository: custRepo,
	}
}

func (s *DefaultAuthService) Register(customer models.Customer, password string) (models.Customer, string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 8 {
		return models.Customer{}, "", errors.New("password must be at least 8 characters long")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return models.Customer{}, "", errors.New("failed to process password")
	}

	customer.PasswordHash = hash
	createdCustomer, err := s.customerService.CreateCustomer(customer)
	if err != nil {
		return models.Customer{}, "", err
	}

	token, err := auth.GenerateToken(createdCustomer.CustomerID, createdCustomer.Email)
	if err != nil {
		return models.Customer{}, "", errors.New("failed to generate authentication token")
	}

	return createdCustomer, token, nil
}

func (s *DefaultAuthService) Login(email, password string) (models.Customer, string, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if email == "" || password == "" {
		return models.Customer{}, "", errors.New("email and password are required")
	}

	customer, err := s.customerRepository.FindByEmail(email)
	if err != nil {
		return models.Customer{}, "", errors.New("invalid email or password")
	}

	if !auth.CheckPasswordHash(password, customer.PasswordHash) {
		return models.Customer{}, "", errors.New("invalid email or password")
	}

	if customer.Status == models.CustomerStatusBlocked || customer.Status == models.CustomerStatusClosed {
		return models.Customer{}, "", errors.New("customer account is blocked or closed")
	}

	token, err := auth.GenerateToken(customer.CustomerID, customer.Email)
	if err != nil {
		return models.Customer{}, "", errors.New("failed to generate authentication token")
	}

	return customer, token, nil
}
