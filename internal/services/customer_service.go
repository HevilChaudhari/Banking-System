package services

import (
	"banking-system/internal/models"
	"banking-system/internal/repositories"
	validation "banking-system/internal/validations"
	"errors"
	"strings"
	"time"
)

type CustomerService interface {
	CreateCustomer(customer models.Customer) (models.Customer, error)
	GetCustomerByID(customerID int) (models.Customer, error)
	GetCustomerByEmail(email string) (models.Customer, error)
	UpdateCustomer(customer models.Customer) (models.Customer, error)
}

type InMemoryCustomerService struct {
	customerRepository repositories.CustomerRepository
}

func NewInMemoryCustomerService(
	customerRepository repositories.CustomerRepository,
) *InMemoryCustomerService {
	return &InMemoryCustomerService{
		customerRepository: customerRepository,
	}
}

func (service *InMemoryCustomerService) CreateCustomer(customer models.Customer) (models.Customer, error) {

	customer.FirstName = strings.TrimSpace(customer.FirstName)
	customer.LastName = strings.TrimSpace(customer.LastName)
	customer.Email = strings.TrimSpace(customer.Email)
	customer.PhoneNumber = strings.TrimSpace(customer.PhoneNumber)
	customer.Address = strings.TrimSpace(customer.Address)

	if customer.FirstName == "" || customer.LastName == "" {
		return models.Customer{}, errors.New("firstname or lastname cannot be empty")
	}

	if !validation.IsValidEmail(customer.Email) {
		return models.Customer{}, errors.New("email is invalid")
	}

	if !validation.IsValidPhoneNumber(customer.PhoneNumber) {
		return models.Customer{}, errors.New("phone number is invalid")
	}

	if !validation.IsValidDateOfBirth(customer.DateOfBirth) {
		return models.Customer{}, errors.New("customer must be at least 18 years old")
	}

	customer.CustomerID = 0
	customer.CreatedAt = time.Now()
	customer.KYCStatus = models.KYCStatusNotStarted
	customer.Status = models.CustomerStatusPendingVerification

	return service.customerRepository.Create(customer)
}

func (service *InMemoryCustomerService) GetCustomerByID(customerID int) (models.Customer, error) {
	if customerID <= 0 {
		return models.Customer{}, errors.New("customer id must be greater than zero")
	}

	return service.customerRepository.FindByID(customerID)
}

func (service *InMemoryCustomerService) GetCustomerByEmail(email string) (models.Customer, error) {

	email = strings.TrimSpace(email)

	if !validation.IsValidEmail(email) {
		return models.Customer{}, errors.New("invalid email")
	}

	return service.customerRepository.FindByEmail(email)
}

func (service *InMemoryCustomerService) UpdateCustomer(customer models.Customer) (models.Customer, error) {

	if customer.CustomerID <= 0 {
		return models.Customer{}, errors.New("customer id must be greater than zero")
	}

	customer.FirstName = strings.TrimSpace(customer.FirstName)
	customer.LastName = strings.TrimSpace(customer.LastName)
	customer.Email = strings.TrimSpace(customer.Email)
	customer.PhoneNumber = strings.TrimSpace(customer.PhoneNumber)
	customer.Address = strings.TrimSpace(customer.Address)

	if customer.FirstName == "" || customer.LastName == "" {
		return models.Customer{}, errors.New("firstname or lastname cannot be empty")
	}

	if !validation.IsValidEmail(customer.Email) {
		return models.Customer{}, errors.New("email is invalid")
	}

	if !validation.IsValidPhoneNumber(customer.PhoneNumber) {
		return models.Customer{}, errors.New("phone number is invalid")
	}

	if !validation.IsValidDateOfBirth(customer.DateOfBirth) {
		return models.Customer{}, errors.New("customer must be at least 18 years old")
	}

	existingCustomer, err := service.GetCustomerByID(customer.CustomerID)
	if err != nil {
		return models.Customer{}, err
	}

	customer.CreatedAt = existingCustomer.CreatedAt
	customer.KYCStatus = existingCustomer.KYCStatus
	customer.Status = existingCustomer.Status

	return service.customerRepository.Update(customer)
}
