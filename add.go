package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	descriptionMaxLength = 40 // chars
)

// add <transaction_type> <amount> <category> <description>
func addTransaction(args []string) (success bool, err error) {
	if len(args) < 4 {
		return false, fmt.Errorf("usage: add <transcation type> <amount> <category> <description>")
	}

	transactionType, err := normalizeTransactionType(args[0])
	if err != nil {
		return false, fmt.Errorf("transaction type error: %w", err)
	}

	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return false, fmt.Errorf("\ninvalid amount: %w\n", err)
	}

	category := args[2]
	if _, ok := allowedTransactionCategories[transactionType][category]; !ok {
		fmt.Printf("\ninvalid transaction category: \"%s\"", category)
		showAllowedCategories(transactionType) // expense, income, investment
		return false, fmt.Errorf("\n\nPlease pick a valid transaction category from the list above.")
	}

	description := strings.Join(args[3:], " ")
	if len(description) > descriptionMaxLength {
		return false, fmt.Errorf("\ndescription should be a maximum of %v characters, provided %v", descriptionMaxLength, len(description))
	}

	if !validDescriptionInputFormat(description) {
		return false, fmt.Errorf("\ninvalid character in description, should contain only letters, numbers, spaces, commas, or dashes")
	}

	// TODO: extend this to support adding transactions for a specific month and not only the current one
	year := strings.ToLower(strconv.Itoa(time.Now().Year()))
	month := strings.ToLower(time.Now().Month().String())

	return handleTransactionAdd(transactionType, amount, category, description, month, year)
}

func handleTransactionAdd(transactionType string, amount float64, category, description, month, year string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	if _, ok := transactions[year]; !ok {
		transactions[year] = make(map[string]map[string][]Transaction)
	}

	if _, ok := transactions[year][month]; !ok {
		transactions[year][month] = make(map[string][]Transaction)
	}

	if _, ok := transactions[year][month][transactionType]; !ok {
		transactions[year][month][transactionType] = []Transaction{}
	}

	var transactionId string
	if transactionId, err = generateTransactionId(); err != nil {
		return false, fmt.Errorf("unable to generate transaction id: %w", err)
	}

	// make sure only unique IDs are used
	for {
		var duplicateIdFound bool
		for txType := range transactions[year][month] {
			for _, t := range transactions[year][month][txType] {
				if transactionId == t.Id {
					duplicateIdFound = true
					break // id is already in use
				}
			}
			if duplicateIdFound {
				break
			}
		}

		if !duplicateIdFound {
			break // id is unique
		}

		if transactionId, err = generateTransactionId(); err != nil {
			return false, fmt.Errorf("unable to generate transaction id: %w", err)
		}
	}

	newTransaction := Transaction{
		Id:          transactionId,
		Amount:      amount,
		Category:    category,
		Description: description,
	}

	transactions[year][month][transactionType] = append(transactions[year][month][transactionType], newTransaction)
	if saveTransactionErr := saveTransactions(transactions); saveTransactionErr != nil {
		return false, fmt.Errorf("Error saving transaction: %w", saveTransactionErr)
	}

	fmt.Printf("\nsuccessfully added %s €%.2f | %s | %s\n", transactionType, amount, category, description)
	return true, nil
}
