package main

import (
	"fmt"
)

// delete <transaction_type> <transaction_id>
func deleteTransaction(args []string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	if len(args) < 2 {
		return false, fmt.Errorf("expected arguments for delete: <transaction_type> <transaction_id>, provided %s", args)
	}

	transactionType := normalizeTransactionType(args[0])
	if _, ok := validTransactionTypes[transactionType]; !ok {
		return false, fmt.Errorf("invalid transaction type %s, please use expense, investment, income", transactionType)
	}

	transactionId := args[1]
	if len(transactionId) != 8 {
		return false, fmt.Errorf("invalid transaction id, expected 8 char id, got %s", transactionId)
	}

	for year, months := range transactions {

		for month := range months {

			var txList = transactions[year][month][transactionType]
			for i, t := range txList {
				if t.Id == transactionId {
					transactions[year][month][transactionType] = removeTransactionAtIndex(txList, i)

					if saveTransactionErr := saveTransactions(transactions); saveTransactionErr != nil {
						return false, fmt.Errorf("Error saving transaction: %s", saveTransactionErr)
					}
					fmt.Printf("successfully removed transaction with id %s\n\n", transactionId)

					fmt.Printf("%s for %s %s\n", transactionType, month, year)
					_, err = listTransactionsByMonth(transactionType, month, year)
					if err != nil {
						return false, fmt.Errorf("unable to list remaining transactions: %s", err)
					}

					return true, nil
				}
			}
		}
	}

	return false, fmt.Errorf("\ndid not match any transaction by id %s, please run list %s or show-total and confirm the transaction id that you want to delete\n", transactionId, transactionType)
}

func removeTransactionAtIndex(transactions []Transaction, index int) []Transaction {
	if index < 0 || index >= len(transactions) {
		return transactions // index out of range return original
	}
	return append(transactions[:index], transactions[index+1:]...)
}
