package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

func cleanInput(cmdArgs string) []string {
	var sanitizedText = strings.Trim(strings.ToLower(cmdArgs), " ")
	return strings.Split(sanitizedText, " ")
}

func validNoteInputFormat(note string) bool {
	// only letters, numbers, commas, spaces or dashes
	pattern := `^[a-zA-Z0-9,' '-]+$`
	matched, err := regexp.MatchString(pattern, note)
	if err != nil {
		return false
	}

	return matched
}

func normalizeTransactionType(t string) string {
	switch t {

	case "expense", "expenses":
		return "Expenses"

	case "investment", "investments":
		return "Investments"

	case "income":
		return "Income"

	default:
		return t // unknown type returned as is
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

// TODO: function to validate there are no ID is not used already in data json
