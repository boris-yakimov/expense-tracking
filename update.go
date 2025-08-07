package main

import (
	"fmt"
	"strings"
)

// update <transaction_type> <transaction_id> <amount> <category> <description>
func updateTransaction(args []string) (success bool, err error) {
	if len(args) < 5 {
		return false, fmt.Errorf("expected arguments for update: <transaction_type> <transaction_id> <amount> <category> <description>, provided %s", args)
	}

	transactionType := normalizeTransactionType(args[0])
	if _, ok := validTransactionTypes[transactionType]; !ok {
		return false, fmt.Errorf("invalid transaction type %s, please use expense, investment, income", transactionType)
	}

	transactionId := args[1]
	if len(transactionId) != 8 {
		return false, fmt.Errorf("invalid transaction id, expected 8 char id, got %s", transactionId)
	}

	updatedAmount := args[2]

	updatedCategory := args[3]
	if _, ok := allowedTransactionCategories[transactionType][updatedCategory]; !ok {
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

	// easier to just delete and add a new transaction rather than updating an existing one
	// the id of the transaction gets changed in the process
	_, err = deleteTransaction([]string{transactionType, transactionId})
	if err != nil {
		return false, fmt.Errorf("unable to delete transaction: %s", err)
	}

	_, err = addTransaction([]string{transactionType, updatedAmount, updatedCategory, updatedDescription})
	if err != nil {
		return false, fmt.Errorf("unable to update transaction: %s", err)
	}

	fmt.Printf("transaction successully updated")

	return true, nil
}
