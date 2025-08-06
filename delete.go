package main

import (
	"fmt"
)

func deleteTransaction(args []string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	if len(args) < 2 {
		return false, fmt.Errorf("expected arguments for delete: <delete> <transaction_id>, provided %s", args)
	}

	transactionType := args[0]
	if _, ok := validTransactionTypes[transactionType]; !ok {
		return false, fmt.Errorf("invalid transaction type %s, please use expense, investment, income", transactionType)
	}

	transactionId := args[1]
	// TODO: maybe a separate function to do this validation and to also check its format for special chars, spaces, and stuff
	if len(transactionId) != 9 {
		return false, fmt.Errorf("invalid transaction id, expected 8 char id, got %s", transactionId)
	}

	// for loop through year, than month, than each transaction, compare transactionId with t.id and if they match ?
	// maybe re-wraite the whole transactions file without the one that matches
	// isn't there a more efficient way ?

	for year, months := range transactions {
		for month := range months {
			var txList = transactions[year][month][transactionType]
			for i, t := range txList {
				if t.Id == transactionId {
					removeTransactionAtIndex(txList, i)

					if saveTransactionErr := saveTransactions(transactions); saveTransactionErr != nil {
						return false, fmt.Errorf("Error saving transaction: %s", saveTransactionErr)
					}
					fmt.Printf("successfully removed transaction with id %s", transactionId)

					// TODO: show a list of remaining transactions after the deletion has happened
					return true, nil
				}
			}
		}
	}

	return false, fmt.Errorf("\ndid not match any transaction by id %s, please run list or show-total and confirm the transaction id that you want to delete\n", transactionId)
}

func removeTransactionAtIndex(transactions []Transaction, index int) []Transaction {
	if index < 0 || index >= len(transactions) {
		return transactions // index out of range return original
	}
	return append(transactions[:index], transactions[index+1:]...)
}
