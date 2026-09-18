package handlers

import (
	"banking-system/internal/models"
	"banking-system/internal/services"
	"encoding/json"
	"net/http"
	"strconv"
)

type CustomerHandler struct {
	customerService services.CustomerService
}

func NewCustomerHandler(customerService services.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

func (handler *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {

	var customer models.Customer

	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	createdCustomer, err := handler.customerService.CreateCustomer(customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdCustomer)

}

func (handler *CustomerHandler) GetCustomerByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerIDText := r.PathValue("id")

	customerID, err := strconv.Atoi(customerIDText)
	if err != nil || customerID <= 0 {
		http.Error(w, "customer id must be a positive number", http.StatusBadRequest)
		return
	}

	customer, err := handler.customerService.GetCustomerByID(customerID)
	if err != nil {
		http.Error(w, "customer not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func (handler *CustomerHandler) GetCustomerByEmail(
	w http.ResponseWriter,
	r *http.Request,
) {
	email := r.URL.Query().Get("email")

	customer, err := handler.customerService.GetCustomerByEmail(email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func (handler *CustomerHandler) UpdateCustomer(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerIDText := r.PathValue("id")

	customerID, err := strconv.Atoi(customerIDText)
	if err != nil || customerID <= 0 {
		http.Error(w, "customer id must be a positive number", http.StatusBadRequest)
		return
	}

	var customer models.Customer

	err = json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	customer.CustomerID = customerID

	updatedCustomer, err := handler.customerService.UpdateCustomer(customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCustomer)
}
