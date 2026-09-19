package validation

import (
	"banking-system/internal/models"
	"strings"
)

func IsValidTransactionType(transactionType models.TransactionType) bool {

	return transactionType == models.TransactionTypeDeposit ||
		transactionType == models.TransactionTypeWithdrawal ||
		transactionType == models.TransactionTypeTransfer
}

func IsValidTransactionStatus(status models.TransactionStatus) bool {
	return status == models.TransactionStatusPending ||
		status == models.TransactionStatusCompleted ||
		status == models.TransactionStatusReversed ||
		status == models.TransactionStatusFailed
}

func IsValidAmount(amount int64) bool {
	return amount > 0
}

func IsValidTransferAccounts(source string, destination string) bool {
	source = strings.TrimSpace(source)
	destination = strings.TrimSpace(destination)

	// Neither can be empty, and they cannot be the same account
	if source == "" || destination == "" {
		return false
	}

	return source != destination
}
