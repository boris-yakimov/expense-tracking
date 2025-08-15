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
		return false, fmt.Errorf("transaction type error: %s", err)
	}

	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return false, fmt.Errorf("\ninvalid amount: %v\n", err)
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
	transcations, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	if _, ok := transcations[year]; !ok {
		transcations[year] = make(map[string]map[string][]Transaction)
	}

	if _, ok := transcations[year][month]; !ok {
		transcations[year][month] = make(map[string][]Transaction)
	}

	if _, ok := transcations[year][month][transactionType]; !ok {
		transcations[year][month][transactionType] = []Transaction{}
	}

	var transactionId string
	if transactionId, err = generateTransactionId(); err != nil {
		return false, fmt.Errorf("Unable to generate transaction id: %s", err)
	}

	newTransaction := Transaction{
		Id:          transactionId,
		Amount:      amount,
		Category:    category,
		Description: description,
	}

	transcations[year][month][transactionType] = append(transcations[year][month][transactionType], newTransaction)
	if saveTransactionErr := saveTransactions(transcations); saveTransactionErr != nil {
		return false, fmt.Errorf("Error saving transaction: %s", saveTransactionErr)
	}

	fmt.Printf("\nadded %s €%.2f | %s | %s\n", transactionType, amount, category, description)

	if _, err = listAllTransactions(); err != nil {
		return false, fmt.Errorf("%s", err)
	}

	return true, nil
}
