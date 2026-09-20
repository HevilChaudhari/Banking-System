package handlers

import (
	"encoding/json"
	"net/http"

	"banking-system/internal/services"
)

type TransactionHandler struct {
	transactionService services.TransactionService
}

func NewTransactionHandler(transactionService services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

type DepositRequest struct {
	AccountNumber string `json:"accountNumber"`
	Amount        int64  `json:"amount"`
	Description   string `json:"description"`
}

type WithdrawRequest struct {
	AccountNumber string `json:"accountNumber"`
	Amount        int64  `json:"amount"`
	Description   string `json:"description"`
}

type TransferRequest struct {
	SourceAccountNumber      string `json:"sourceAccountNumber"`
	DestinationAccountNumber string `json:"destinationAccountNumber"`
	Amount                   int64  `json:"amount"`
	Description              string `json:"description"`
}

func (h *TransactionHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.transactionService.Deposit(req.AccountNumber, req.Amount, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

func (h *TransactionHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.transactionService.Withdraw(req.AccountNumber, req.Amount, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.transactionService.Transfer(req.SourceAccountNumber, req.DestinationAccountNumber, req.Amount, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

func (h *TransactionHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	transactionID := r.PathValue("id")
	if transactionID == "" {
		http.Error(w, "transaction id is required", http.StatusBadRequest)
		return
	}

	tx, err := h.transactionService.GetTransactionByID(transactionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func (h *TransactionHandler) GetTransactionsByAccountNumber(w http.ResponseWriter, r *http.Request) {
	accountNumber := r.PathValue("accountNumber")
	if accountNumber == "" {
		http.Error(w, "account number is required", http.StatusBadRequest)
		return
	}

	transactions, err := h.transactionService.GetTransactionsByAccountNumber(accountNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}
