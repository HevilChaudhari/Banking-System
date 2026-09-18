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

	fmt.Println("Server Started on Port :8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Server Error ", err)
	}
}
