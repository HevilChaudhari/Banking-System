package services

import (
	"banking-system/internal/models"
	"banking-system/internal/repositories"
	validation "banking-system/internal/validations"
	"errors"
	"time"
)

type CustomerService interface {
	CreateCustomer(customer models.Customer) (models.Customer, error)
	GetCustomerByID(customerID int) (models.Customer, error)
	GetCustomerByEmail(email string) (models.Customer, error)
	UpdateCustomer(customer models.Customer) (models.Customer, error)
}

type InMemoryCustomerService struct {
	customerRepository *repositories.InMemoryCustomerRepository
}

func (service *InMemoryCustomerService) CreateCustomer(customer models.Customer) (models.Customer, error) {

	if customer.FirstName == "" || customer.LastName == "" {
		return models.Customer{}, errors.New("Firstname or Lastname cannot be empty")
	}

	if !validation.IsValidEmail(customer.Email) {
		return models.Customer{}, errors.New("Email is invalid")
	}

	if !validation.IsValidPhoneNumber(customer.PhoneNumber) {
		return models.Customer{}, errors.New("Phone number is invalid")
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
	return models.Customer{}, nil
}

func (service *InMemoryCustomerService) GetCustomerByEmail(email string) (models.Customer, error) {
	return models.Customer{}, nil

}

func (service *InMemoryCustomerService) UpdateCustomer(customer models.Customer) (models.Customer, error) {
	return models.Customer{}, nil
}
