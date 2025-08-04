package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// TODO: montly totals with less details - income, expense, investment, % p&l
// TODO: yearly totals with less details - income, expense, investment, %p&l
// TODO: draw a terminal diagram/pie chart

const (
	amountWidth   = 10
	categoryWidth = 12
	noteWidth     = 40
)

// TODO: move the looping through year and month in showTotal and remove it from listTransactions, listTransactions should accept the month and year as arguments and just loop through those
func showTotal(args []string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	// years
	for year, months := range transactions {
		fmt.Printf("\nYear: %s\n", year)

		// months
		for month, _ := range months {
			fmt.Printf("  Month: %s\n\n", month)

			if _, err := listTransactionsByMonth([]string{"expenses", month, year}); err != nil {
				return false, fmt.Errorf("%s", err)
			}

			if _, err := listTransactionsByMonth([]string{"investments", month, year}); err != nil {
				return false, fmt.Errorf("%s", err)
			}

			if _, err := listTransactionsByMonth([]string{"income", month, year}); err != nil {
				return false, fmt.Errorf("%s", err)
			}

			calculatedPnL, err := calculatePnL(month, year)
			if err != nil {
				return false, fmt.Errorf("Unable to calculate P&L: %s", err)
			}
			fmt.Printf("\np&l result: €%.2f | %.1f%%\n\n", calculatedPnL.Amount, calculatedPnL.Percent)
		}
	}

	return true, nil
}

func listTransactionsByMonth(args []string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %w", loadFileErr)
	}

	if len(transactions) == 0 {
		fmt.Println("\nNo transactions found")
		return true, nil
	}

	var transactionType string
	var month string
	var year string

	if len(args) > 0 {
		transactionType = normalizeTransactionType(args[0])
		// TODO: seems hacky, figure out a better way
		// panic: runtime error: index out of range [1] with length 1
		if len(args) >= 2 {
			month = args[1]
		}
		if len(args) >= 3 {
			year = args[2]
		}

		if _, ok := validTransactionTypes[transactionType]; !ok {
			return false, fmt.Errorf("invalid transaction type %s, please use expenses, income, or investments", transactionType)
		}
	} else {
		transactionType = "" // used to list all transactions no matter their type in case list is called without arguments
	}

	if transactionType == "" {
		// years
		for year, months := range transactions {
			fmt.Printf("\nYear: %s\n", year)

			// months
			for month, transactionTypes := range months {
				fmt.Printf("  Month: %s\n\n", month)
				// expenses, investments, or income
				for transactionType, transactionList := range transactionTypes {
					fmt.Printf("    %s\n", transactionType)
					if len(transactionList) == 0 {
						fmt.Println("\nNo transactions recorded.")
						continue
					}

					// list of each transaction
					for i, e := range transactionList {
						fmt.Printf("    %2d. €%-8.2f | %-10s | %-25s\n", i+1, e.Amount, e.Category, e.Note)
					}

					fmt.Println()
				}
			}
		}
	} else {
		// fmt.Printf("%s %s\n", month, year)
		fmt.Printf("    %s\n", transactionType)
		// if month and year are passed also passed, list only relevant transactions
		// list of each transaction
		for i, t := range transactions[year][month][transactionType] {
			// TODO: check if transaction list is empty for a specific month and if it is skip it
			fmt.Printf("    %2d. €%-8.2f | %-10s | %-25s\n", i+1, t.Amount, t.Category, t.Note)
		}

		fmt.Println()
	}

	return true, nil
}

func showAllowedCategories(transactionType string) error {
	fmt.Println("\nallowed categories are: ")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\nCategory\tDescription")
	fmt.Fprintln(w, "--------\t-----------")

	// TODO: can i make this into a function so that i don't have to constantly do these cheks
	if transactionType == "expense" || transactionType == "expenses" {
		for key, val := range allowedTransactionCategories["expense"] {
			fmt.Fprintf(w, "%s\t%s\n", key, val)
		}
		w.Flush()
		return nil
	}

	if transactionType == "investment" || transactionType == "investments" {
		for key, val := range allowedTransactionCategories["investment"] {
			fmt.Fprintf(w, "%s\t%s\n", key, val)
		}
		w.Flush()
		return nil
	}

	if transactionType == "income" {
		for key, val := range allowedTransactionCategories["income"] {
			fmt.Fprintf(w, "%s\t%s\n", key, val)
		}
		w.Flush()
		return nil
	}

	w.Flush()
	return fmt.Errorf("\nallowed types are expense, income, or investment - provided %s", transactionType)
}
