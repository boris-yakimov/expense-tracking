package main

import (
	"fmt"
)

type PnLResult struct {
	Amount  float64
	Percent float64
}

func calculatePnL(month, year string) (PnLResult, error) {
	var monthReceived float64
	var monthSpent float64
	var calculatedPnL PnLResult

	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return calculatedPnL, fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	for transcationType, transactionList := range transactions[year][month] {
		if len(transactionList) == 0 {
			fmt.Printf("\nNo transactions of type %s\n", transcationType)
			continue
		}

		for _, transaction := range transactionList {

			if transcationType == "Income" {
				monthReceived += transaction.Amount
				calculatedPnL.Amount += transaction.Amount
			}

			if transcationType == "Expenses" || transcationType == "Investments" {
				monthSpent += transaction.Amount
				calculatedPnL.Amount -= transaction.Amount
			}
		}
	}
	calculatedPnL.Percent = ((monthReceived - monthSpent) / monthReceived) * 100

	return calculatedPnL, nil
}
