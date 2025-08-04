package main

import (
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
