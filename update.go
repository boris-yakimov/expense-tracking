package main

import (
	"fmt"
	"strconv"
	"strings"
)

func formUpdateTransaction() error {
	// TODO: update function
	return nil
}

// TODO: make a common function that does all validations as we seems to be doing the same checks on add, update, and delete which is redundant - amount float conversion; txType checks; txId validation; category checks; description checks
// TODO: remove these after TUI approach is implemented

// update <transaction_type> <transaction_id> <amount> <category> <description>
func updateTransaction(args []string) (success bool, err error) {
	if len(args) < 5 {
		return false, fmt.Errorf("expected arguments for update: <transaction_type> <transaction_id> <amount> <category> <description>, provided %s", args)
	}

	transactionType, err := normalizeTransactionType(args[0])
	if err != nil {
		return false, fmt.Errorf("transaction type error: %w", err)
	}

	transactionId := args[1]
	if len(transactionId) != 8 {
		return false, fmt.Errorf("invalid transaction id, expected 8 char id, got %s", transactionId)
	}

	updatedAmount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return false, fmt.Errorf("\ninvalid amount: %w\n", err)
	}

	updatedCategory := args[3]
	if _, ok := allowedTransactionCategories[transactionType][updatedCategory]; !ok {
		fmt.Println(allowedTransactionCategories[transactionType])
		fmt.Printf("\ninvalid transaction category: \"%s\"", updatedCategory)
		showAllowedCategories(transactionType) // expense, income, investment
		return false, fmt.Errorf("\n\nplease pick a valid transaction category from the list above.")
	}

	updatedDescription := strings.Join(args[4:], " ")
	if len(updatedDescription) > descriptionMaxCharLength {
		return false, fmt.Errorf("\ndescription should be a maximum of %v characters, provided %v", descriptionMaxCharLength, len(updatedDescription))
	}
	if !validDescriptionInputFormat(updatedDescription) {
		return false, fmt.Errorf("\ninvalid character in description, should contain only letters, numbers, spaces, commas, or dashes")
	}

	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
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
		return false, fmt.Errorf("error saving transaction: %w", saveTransactionErr)
	}
	fmt.Printf("transaction successully updated")

	return true, nil
}
