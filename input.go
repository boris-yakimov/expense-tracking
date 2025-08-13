package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

func cleanTerminalInput(cmdArgs string) []string {
	var sanitizedText = strings.Trim(strings.ToLower(cmdArgs), " ")
	return strings.Split(sanitizedText, " ")
}

func validDescriptionInputFormat(description string) bool {
	// only letters, numbers, commas, spaces or dashes
	pattern := `^[a-zA-Z0-9,' '-]+$`
	matched, err := regexp.MatchString(pattern, description)
	if err != nil {
		return false
	}

	return matched
}

func normalizeTransactionType(t string) (string, error) {
	switch t {

	case "expense", "expenses", "Expenses", "Expense":
		return "expense", nil

	case "investment", "investments", "Investments", "Investment":
		return "investment", nil

	case "income", "Income":
		return "income", nil

	default:
		return "", fmt.Errorf("\ninvalid transaction type %s - supported transactions types are income, expense, and investment", t)
	}
}

func generateTransactionId() (id string, err error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err = rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("error generating transaction id: %s", err)
	}

	return hex.EncodeToString(bytes), nil
}

// TODO: function to validate that ID is not used already in data json
