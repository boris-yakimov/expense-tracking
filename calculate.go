package main

import (
	"fmt"
	"log"
)

type PnLResult struct {
	incomeTotal     float64
	expenseTotal    float64
	investmentTotal float64
	pnlAmount       float64
	pnlPercent      float64
}

// calculates the p&l for a specific month
func calculateMonthPnL(month, year string) (PnLResult, error) {
	var pnl PnLResult

	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return pnl, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	for txType, txList := range transactions[year][month] {
		if len(txList) == 0 {
			log.Printf("\nno transactions of type %s for %s %s\n", txType, month, year)
			continue
		}

		for _, tx := range txList {

			if txType == "income" {
				pnl.incomeTotal += tx.Amount
			}

			if txType == "expense" {
				pnl.expenseTotal += tx.Amount
			}

			if txType == "investment" {
				pnl.investmentTotal += tx.Amount
			}
		}
	}

	// calucalte the reulting P&L: income - spend(expenses and investments)
	// in absolute value and in % savings i.e. if you receved 1000 and spent 400, this will be 60% savings rate
	if pnl.incomeTotal == 0 { // avoid division by zero
		pnl.pnlAmount = pnl.incomeTotal - (pnl.expenseTotal + pnl.investmentTotal)
		pnl.pnlPercent = 0
	} else {
		pnl.pnlAmount = pnl.incomeTotal - (pnl.expenseTotal + pnl.investmentTotal)
		pnl.pnlPercent = ((pnl.incomeTotal - (pnl.expenseTotal + pnl.investmentTotal)) / pnl.incomeTotal) * 100
	}

	return pnl, nil
}

// calculates the p&l for a specific month
func calculateYearPnL(year string) (PnLResult, error) {
	var pnl PnLResult

	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return pnl, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	for month := range transactions[year] {
		for txType, txList := range transactions[year][month] {
			if len(txList) == 0 {
				log.Printf("\nno transactions of type %s for month %s\n", txType, month)
				continue
			}

			for _, tx := range txList {

				if txType == "income" {
					pnl.incomeTotal += tx.Amount
				}

				if txType == "expense" {
					pnl.expenseTotal += tx.Amount
				}

				if txType == "investment" {
					pnl.investmentTotal += tx.Amount
				}
			}
		}
	}

	// calucalte the reulting P&L: income - spend(expenses and investments)
	// in absolute value and in % savings i.e. if you receved 1000 and spent 400, this will be 60% savings rate
	if pnl.incomeTotal == 0 { // avoid division by zero
		pnl.pnlAmount = pnl.incomeTotal - (pnl.expenseTotal + pnl.investmentTotal)
		pnl.pnlPercent = 0
	} else {
		pnl.pnlAmount = pnl.incomeTotal - (pnl.expenseTotal + pnl.investmentTotal)
		pnl.pnlPercent = ((pnl.incomeTotal - (pnl.expenseTotal + pnl.investmentTotal)) / pnl.incomeTotal) * 100
	}

	return pnl, nil
}
