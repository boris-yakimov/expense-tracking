package main

import (
	"fmt"
	"strconv"
	"strings"
)

// update <transaction_type> <transaction_id> <amount> <category> <description>
func updateTransaction(args []string) (success bool, err error) {
	if len(args) < 5 {
		return false, fmt.Errorf("expected arguments for update: <transaction_type> <transaction_id> <amount> <category> <description>, provided %s", args)
	}

	transactionType, err := normalizeTransactionType(args[0])
	if err != nil {
		return false, fmt.Errorf("transaction type normalization error: %s", err)
	}
	if _, ok := validTransactionTypes[transactionType]; !ok {
		return false, fmt.Errorf("invalid transaction type %s, please use expense, investment, income", transactionType)
	}

	transactionId := args[1]
	if len(transactionId) != 8 {
		return false, fmt.Errorf("invalid transaction id, expected 8 char id, got %s", transactionId)
	}

	updatedAmount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return false, fmt.Errorf("\ninvalid amount: %v\n", err)
	}

	updatedCategory := args[3]
	if _, ok := allowedTransactionCategories[transactionType][updatedCategory]; !ok {
		fmt.Println(allowedTransactionCategories[transactionType])
		fmt.Printf("\ninvalid transaction category: \"%s\"", updatedCategory)
		showAllowedCategories(transactionType) // expense, income, investment
		return false, fmt.Errorf("\n\nPlease pick a valid transaction category from the list above.")
	}

	updatedDescription := strings.Join(args[4:], " ")
	if len(updatedDescription) > descriptionMaxLength {
		return false, fmt.Errorf("\ndescription should be a maximum of %v characters, provided %v", descriptionMaxLength, len(updatedDescription))
	}
	if !validDescriptionInputFormat(updatedDescription) {
		return false, fmt.Errorf("\ninvalid character in description, should contain only letters, numbers, spaces, commas, or dashes")
	}

	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %w", loadFileErr)
	}

	// years
	for year, months := range transactions {

		// months
		for month := range months {

			for i, tx := range transactions[year][month][transactionType] {
				if tx.Id == transactionId {
					tx.Amount = updatedAmount
					tx.Description = updatedDescription
					tx.Category = updatedCategory

					transactions[year][month][transactionType][i] = tx
				}
			}
		}
	}

	if saveTransactionErr := saveTransactions(transactions); saveTransactionErr != nil {
		return false, fmt.Errorf("Error saving transaction: %s", saveTransactionErr)
	}
	fmt.Printf("transaction successully updated")

	return true, nil
}
