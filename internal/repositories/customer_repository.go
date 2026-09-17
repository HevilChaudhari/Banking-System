package repositories

import (
	"errors"
	"sync"

	"banking-system/cmd/internal/models"
)

type CustomerRepository interface {
	Create(customer models.Customer) (models.Customer, error)
	FindByID(customerID int) (models.Customer, error)
	FindByEmail(email string) (models.Customer, error)
	Update(customer models.Customer) error
}

type InMemoryCustomerRepository struct {
	mutex          sync.RWMutex
	customers      map[int]models.Customer
	nextCustomerID int
}

func (repository *InMemoryCustomerRepository) Create(customer models.Customer) (models.Customer, error) {

	for _, cus := range repository.customers {
		if cus.Email == customer.Email {
			return models.Customer{}, errors.New("Customer Already Exist")
		}
	}

	repository.nextCustomerID++

	return models.Customer{}, nil
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

func (repository *InMemoryCustomerRepository) Update(customer models.Customer) error {
	return errors.New("Not found")
}
