package main

import (
	"fmt"
	"log"
	"net/http"

	"banking-system/internal/db"
	"banking-system/internal/handlers"
	"banking-system/internal/middleware"
	"banking-system/internal/repositories"
	"banking-system/internal/services"
)

func main() {

	pool, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	fmt.Println("Connected to PostgreSQL Database successfully")

	customerRepository := repositories.NewPostgresCustomerRepository(pool)
	customerService := services.NewInMemoryCustomerService(customerRepository)
	customerHandler := handlers.NewCustomerHandler(customerService)

	accountRepository := repositories.NewPostgresAccountRepository(pool)
	accountService := services.NewInMemoryAccountService(accountRepository, customerRepository)
	accountHandler := handlers.NewAccountHandler(accountService)

	transactionRepository := repositories.NewPostgresTransactionRepository(pool)
	transactionService := services.NewPostgresTransactionService(pool, transactionRepository)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	authService := services.NewAuthService(customerService, customerRepository)
	authHandler := handlers.NewAuthHandler(authService)

	// Public Authentication Routes
	http.HandleFunc("POST /auth/register", authHandler.Register)
	http.HandleFunc("POST /auth/login", authHandler.Login)

	// Legacy / Public Customer Creation
	http.HandleFunc("POST /customers", customerHandler.CreateCustomer)
	http.HandleFunc("GET /customers", customerHandler.GetCustomerByEmail)

	// Protected Customer Routes (requires Bearer token)
	http.HandleFunc("GET /customers/{id}", middleware.AuthMiddleware(customerHandler.GetCustomerByID))
	http.HandleFunc("PATCH /customers/{id}", middleware.AuthMiddleware(customerHandler.UpdateCustomer))

	// Protected Account Routes (requires Bearer token)
	http.HandleFunc("POST /accounts", middleware.AuthMiddleware(accountHandler.CreateAccount))
	http.HandleFunc("GET /accounts/{accountNumber}", middleware.AuthMiddleware(accountHandler.GetAccountByNumber))
	http.HandleFunc("GET /customers/{id}/accounts", middleware.AuthMiddleware(accountHandler.GetAccountsByCustomerID))
	http.HandleFunc("PATCH /accounts/{accountNumber}/status", middleware.AuthMiddleware(accountHandler.UpdateAccountStatus))

	// Protected Transaction Routes (requires Bearer token)
	http.HandleFunc("POST /transactions/deposit", middleware.AuthMiddleware(transactionHandler.Deposit))
	http.HandleFunc("POST /transactions/withdraw", middleware.AuthMiddleware(transactionHandler.Withdraw))
	http.HandleFunc("POST /transactions/transfer", middleware.AuthMiddleware(transactionHandler.Transfer))
	http.HandleFunc("GET /transactions/{id}", middleware.AuthMiddleware(transactionHandler.GetTransactionByID))
	http.HandleFunc("GET /accounts/{accountNumber}/transactions", middleware.AuthMiddleware(transactionHandler.GetTransactionsByAccountNumber))

	fmt.Println("Server Started on Port :8080")
	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Server Error ", err)
	}
}
