package main

import (
	"fmt"
)

func deleteTransaction(args []string) (success bool, err error) {
	transcations, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	transactionType := args[0]
	fmt.Println(transactionType)

	// TODO: need to add IDs or numbers before each transaction so that delete can find them easier in delete or update function
	// TODO: del investment id or number

	if saveTransactionErr := saveTransactions(transcations); saveTransactionErr != nil {
		return false, fmt.Errorf("Error saving transaction: %s", saveTransactionErr)
	}
	return true, nil
}
