package main

import (
	"fmt"
)

// TODO: calculate p&l % (should show minus percent if expenses and investments exceed income)

type PnLResult struct {
	Amount  float64
	Percent float64
}

func calculatePnL() (PnLResult, error) {
	var calculatedPnL PnLResult

	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return calculatedPnL, fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	// TODO: figure out a good way to pass year/month to reduce the amount of time I have to loop over these

	// years
	for _, months := range transactions {

		// months
		for _, transactionTypes := range months {

			// expenses, investments, or income
			for transcationType, transactionList := range transactionTypes {
				if len(transactionList) == 0 {
					fmt.Printf("\nNo transactions of type %s\n", transcationType)
					continue
				}

				for _, transaction := range transactionList {

					if transcationType == "Income" {
						calculatedPnL.Amount += transaction.Amount
					}

					if transcationType == "Expenses" || transcationType == "Investments" {
						calculatedPnL.Amount -= transaction.Amount
					}
				}
			}

			fmt.Printf("\np&l result: €%.2f\n", calculatedPnL.Amount)
			calculatedPnL.Amount = 0
		}
		// move to next month/year after we've printed all
		continue
	}

	return calculatedPnL, nil
}
