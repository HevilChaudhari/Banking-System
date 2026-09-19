package main

import (
	"fmt"
	"net/http"

	"banking-system/internal/handlers"
	"banking-system/internal/repositories"
	"banking-system/internal/services"
)

func main() {

	customerRepository := repositories.NewInMemoryCustomerRepository()
	customerService := services.NewInMemoryCustomerService(customerRepository)
	customerHandler := handlers.NewCustomerHandler(customerService)

	http.HandleFunc("POST /customers", customerHandler.CreateCustomer)
	http.HandleFunc("GET /customers/{id}", customerHandler.GetCustomerByID)
	http.HandleFunc("GET /customers", customerHandler.GetCustomerByEmail)
	http.HandleFunc("PATCH /customers/{id}", customerHandler.UpdateCustomer)

	accountRepository := repositories.NewAccountRepository()
	accountService := services.NewInMemoryAccountService(accountRepository, customerRepository)
	accountHandler := handlers.NewAccountHandler(accountService)

	http.HandleFunc("POST /accounts", accountHandler.CreateAccount)
	http.HandleFunc("GET /accounts/{accountNumber}", accountHandler.GetAccountByNumber)
	http.HandleFunc("GET /customers/{id}/accounts", accountHandler.GetAccountsByCustomerID)
	http.HandleFunc("PATCH /accounts/{accountNumber}/status", accountHandler.UpdateAccountStatus)

	fmt.Println("Server Started on Port :8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Server Error ", err)
	}
}
