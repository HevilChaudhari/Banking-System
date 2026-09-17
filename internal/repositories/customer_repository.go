package repositories

import (
	"errors"
	"sync"

	"banking-system/internal/models"
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

func NewInMemoryCustomerRepository() *InMemoryCustomerRepository {
	return &InMemoryCustomerRepository{
		customers:      make(map[int]models.Customer),
		nextCustomerID: 0,
	}
}

func (repository *InMemoryCustomerRepository) Create(customer models.Customer) (models.Customer, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	for _, cus := range repository.customers {
		if cus.Email == customer.Email {
			return models.Customer{}, errors.New("Customer Already Exist")
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
