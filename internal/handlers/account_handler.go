package handlers

import (
	"banking-system/internal/models"
	"banking-system/internal/services"
	"encoding/json"
	"net/http"
	"strconv"
)

type AccountHandler struct {
	accountService services.AccountService
}

func NewAccountHandler(accountService services.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

func (handler *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var account models.Account

	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	createdAccount, err := handler.accountService.CreateAccount(account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAccount)
}

func (handler *AccountHandler) GetAccountByNumber(w http.ResponseWriter, r *http.Request) {
	accountNumber := r.PathValue("accountNumber")

	account, err := handler.accountService.GetAccountByNumber(accountNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (handler *AccountHandler) GetAccountsByCustomerID(w http.ResponseWriter, r *http.Request) {
	idText := r.PathValue("id")

	customerID, err := strconv.Atoi(idText)
	if err != nil || customerID <= 0 {
		http.Error(w, "customer id must be a positive number", http.StatusBadRequest)
		return
	}

	accounts, err := handler.accountService.GetAccountsByCustomerID(customerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

type UpdateAccountStatusRequest struct {
	Status models.AccountStatus `json:"status"`
}

func (handler *AccountHandler) UpdateAccountStatus(w http.ResponseWriter, r *http.Request) {
	accountNumber := r.PathValue("accountNumber")

	var req UpdateAccountStatusRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updatedAccount, err := handler.accountService.UpdateAccountStatus(accountNumber, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedAccount)
}
