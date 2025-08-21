package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var monthOrder = map[string]int{
	"january":   1,
	"february":  2,
	"march":     3,
	"april":     4,
	"may":       5,
	"june":      6,
	"july":      7,
	"august":    8,
	"september": 9,
	"october":   10,
	"november":  11,
	"december":  12,
}

var transactionTypeOrder = map[string]int{
	// descending
	"income":     1,
	"expense":    2,
	"investment": 3,
}

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
		return "", fmt.Errorf("error generating transaction id: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

func capitalize(word string) string {
	if len(word) == 0 {
		return ""
	}

	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])

	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}

func listOfAllowedCategories(transactionType string) (categories []string, err error) {
	for c := range allowedTransactionCategories[transactionType] {
		categories = append(categories, c)
	}

	if len(categories) <= 0 {
		return categories, fmt.Errorf("something went wrong with getting list of allowed categories for transaction type %s", transactionType)
	}

	return categories, nil
}

func listOfAllowedTransactionTypes() (categories []string, err error) {
	var transactionTypes []string
	for t := range allowedTransactionTypes {
		transactionTypes = append(transactionTypes, t)
	}

	if len(transactionTypes) <= 0 {
		return transactionTypes, fmt.Errorf("something went wrong with getting list of allowed transaction types")
	}

	return transactionTypes, nil
}

func getListOfTransactoinIds() (transactionIdList []string, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return transactionIdList, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	// build a list of transaction IDs
	for year := range transactions {
		for month := range transactions[year] {
			for txType := range transactions[year][month] {
				for _, tx := range transactions[year][month][txType] {
					transactionIdList = append(transactionIdList, tx.Id)
				}
			}
		}
	}

	return transactionIdList, nil
}

func getTransactionTypeById(txId string) (txType string, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return "", fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	// find what type of transaction is the particular id related to
	for year := range transactions {
		for month := range transactions[year] {
			for txType := range transactions[year][month] {
				for _, tx := range transactions[year][month][txType] {
					if txId == tx.Id {
						return txType, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("transaction ID %s could not be found in transaction list", txId)
}
