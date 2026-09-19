package services

import (
	"banking-system/internal/models"
	"banking-system/internal/repositories"
	validation "banking-system/internal/validations"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AccountService interface {
	CreateAccount(account models.Account) (models.Account, error)
	GetAccountByNumber(accountNumber string) (models.Account, error)
	GetAccountsByCustomerID(customerID int) ([]models.Account, error)
	UpdateAccountStatus(accountNumber string, status models.AccountStatus) (models.Account, error)
}

type InMemoryAccountService struct {
	accountRepository  repositories.AccountRepository
	customerRepository repositories.CustomerRepository
}

func NewInMemoryAccountService(
	accountRepo repositories.AccountRepository,
	customerRepo repositories.CustomerRepository,
) *InMemoryAccountService {
	return &InMemoryAccountService{
		accountRepository:  accountRepo,
		customerRepository: customerRepo,
	}
}

func (service *InMemoryAccountService) CreateAccount(account models.Account) (models.Account, error) {

	customer, err := service.customerRepository.FindByID(account.CustomerID)
	if err != nil {
		return models.Account{}, errors.New("customer not found")
	}

	if customer.Status == models.CustomerStatusBlocked || customer.Status == models.CustomerStatusClosed {
		return models.Account{}, errors.New("customer is either blocked or closed")
	}

	if !validation.IsValidAccountType(account.AccountType) {
		return models.Account{}, errors.New("account type is invalid")
	}

	if !validation.IsValidCurrency(account.Currency) {
		return models.Account{}, errors.New("account currency is invalid")
	}

	if !validation.IsValidBalance(account.Balance) {
		return models.Account{}, errors.New("account balance is invalid")
	}

	account.Status = models.AccountStatusActive
	account.CreatedAt = time.Now()
	account.AccountNumber = fmt.Sprintf("%d", time.Now().UnixNano())

	return service.accountRepository.Create(account)
}

func (service *InMemoryAccountService) GetAccountByNumber(accountNumber string) (models.Account, error) {

	accountNumber = strings.TrimSpace(accountNumber)
	if accountNumber == "" {
		return models.Account{}, errors.New("account number cannot be empty")
	}
	if len(accountNumber) < 10 {
		return models.Account{}, errors.New("account number is invalid")
	}

	account, err := service.accountRepository.FindByAccountNumber(accountNumber)

	if err != nil {
		return models.Account{}, err
	}

	return account, nil

}

func (service *InMemoryAccountService) GetAccountsByCustomerID(customerID int) ([]models.Account, error) {

	if customerID <= 0 {
		return []models.Account{}, errors.New("customerID is invalid")
	}

	customer, err := service.customerRepository.FindByID(customerID)

	if err != nil {
		return []models.Account{}, err
	}

	if customer.Status == models.CustomerStatusBlocked || customer.Status == models.CustomerStatusClosed {
		return []models.Account{}, errors.New("customer is either blocked or closed")
	}

	return service.accountRepository.FindByCustomerID(customerID)
}

func (service *InMemoryAccountService) UpdateAccountStatus(accountNumber string, status models.AccountStatus) (models.Account, error) {
	accountNumber = strings.TrimSpace(accountNumber)
	if accountNumber == "" {
		return models.Account{}, errors.New("account number cannot be empty")
	}

	if !validation.IsValidAccountStatus(status) {
		return models.Account{}, errors.New("status is invalid")
	}

	account, err := service.accountRepository.FindByAccountNumber(accountNumber)
	if err != nil {
		return models.Account{}, err
	}

	if account.Status == models.AccountStatusClosed {
		return models.Account{}, errors.New("cannot change status of a closed account")
	}

	account.Status = status

	return service.accountRepository.Update(account)
}
