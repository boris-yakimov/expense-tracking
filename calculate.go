package main

import (
	"fmt"
)

type PnLResult struct {
	Amount  float64
	Percent float64
}

func calculatePnL(month, year string) (PnLResult, error) {
	var incomeTotal float64
	var spendTotal float64
	var pnl PnLResult

	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return pnl, fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	for transcationType, transactionList := range transactions[year][month] {
		if len(transactionList) == 0 {
			fmt.Printf("\nNo transactions of type %s\n", transcationType)
			continue
		}

		for _, transaction := range transactionList {

			if transcationType == "Income" {
				incomeTotal += transaction.Amount
			}

			if transcationType == "Expenses" || transcationType == "Investments" {
				spendTotal += transaction.Amount
			}
		}
	}

	// avoid division by zero
	if incomeTotal == 0 {
		pnl.Percent = 0
	} else {
		pnl.Amount = incomeTotal - spendTotal
		pnl.Percent = ((incomeTotal - spendTotal) / incomeTotal) * 100
	}

	return pnl, nil
}
