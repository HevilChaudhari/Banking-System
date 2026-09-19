package validation

import (
	"banking-system/internal/models"
)

func IsValidAccountType(accountType models.AccountType) bool {

	return accountType == models.AccountTypeSavings || accountType == models.AccountTypeCurrent

}

func IsValidCurrency(currency models.CurrencyType) bool {

	return currency == models.CurrencyType_INR
}

func IsValidBalance(balance int64) bool {
	return balance >= 0
}

func IsValidAccountStatus(accountStatus models.AccountStatus) bool {
	return accountStatus == models.AccountStatusActive ||
		accountStatus == models.AccountStatusClosed ||
		accountStatus == models.AccountStatusFrozen
}
