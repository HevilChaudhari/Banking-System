package repositories

import (
	"banking-system/internal/models"
	"errors"
	"sync"
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
