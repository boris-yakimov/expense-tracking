package main

import (
	"fmt"
)

type PnLResult struct {
	Amount  float64
	Percent float64
}

// calculates the p&l for a specific month
func calculateMonthPnL(month, year string) (PnLResult, error) {
	var incomeTotal float64
	var spendTotal float64
	var pnl PnLResult

	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return pnl, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	for transcationType, transactionList := range transactions[year][month] {
		if len(transactionList) == 0 {
			fmt.Printf("\nno transactions of type %s for %s %s\n", transcationType, month, year)
			continue
		}

		for _, transaction := range transactionList {

			if transcationType == "income" {
				incomeTotal += transaction.Amount
			}

			if transcationType == "expense" || transcationType == "investment" {
				spendTotal += transaction.Amount
			}
		}
	}

	// avoid division by zero
	if incomeTotal == 0 {
		pnl.Amount = incomeTotal - spendTotal
		pnl.Percent = 0
	} else {
		pnl.Amount = incomeTotal - spendTotal
		pnl.Percent = ((incomeTotal - spendTotal) / incomeTotal) * 100
	}

	return pnl, nil
}

// calculates the p&l for a specific month
func calculateYearPnL(year string) (PnLResult, error) {
	var incomeTotal float64
	var spendTotal float64
	var pnl PnLResult

	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return pnl, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	for month := range transactions[year] {
		for txType, txList := range transactions[year][month] {
			if len(txList) == 0 {
				fmt.Printf("\nno transactions of type %s for month %s\n", txType, month)
				continue
			}

			for _, transaction := range txList {

				if txType == "income" {
					incomeTotal += transaction.Amount
				}

				if txType == "expense" || txType == "investment" {
					spendTotal += transaction.Amount
				}
			}
		}
	}

	// avoid division by zero
	if incomeTotal == 0 {
		pnl.Amount = incomeTotal - spendTotal
		pnl.Percent = 0
	} else {
		pnl.Amount = incomeTotal - spendTotal
		pnl.Percent = ((incomeTotal - spendTotal) / incomeTotal) * 100
	}

	return pnl, nil
}
