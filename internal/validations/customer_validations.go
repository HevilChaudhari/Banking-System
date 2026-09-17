package validation

import (
	"net/mail"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers/v2"
)

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)

	if email == "" {
		return false
	}

	address, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return address.Address == email
}

func IsValidPhoneNumber(phoneNumber string) bool {
	phoneNumber = strings.TrimSpace(phoneNumber)

	// Require international E.164 format, for example: +919876543210
	if !strings.HasPrefix(phoneNumber, "+") {
		return false
	}

	number, err := phonenumbers.Parse(phoneNumber, "")
	if err != nil {
		return false
	}

	return phonenumbers.IsValidNumber(number)
}

func IsValidDateOfBirth(dateOfBirth time.Time) bool {
	if dateOfBirth.IsZero() {
		return false
	}

	today := time.Now()
	adultDate := today.AddDate(-18, 0, 0)

	return !dateOfBirth.After(adultDate)
}
