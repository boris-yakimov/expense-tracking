package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

const (
	amountWidth      = 10
	categoryWidth    = 12
	descriptionWidth = 40
)

func listAllTransactions(args []string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %w", loadFileErr)
	}

	// years
	for year, months := range transactions {
		fmt.Printf("\nYear: %s\n", year)

		// months
		for month, transactionTypes := range months {
			fmt.Printf("  Month: %s\n\n", month)

			// expenses, investments, or income
			for transactionType, transactionList := range transactionTypes {
				fmt.Printf("    %s\n\n", transactionType)
				if len(transactionList) == 0 {
					fmt.Println("\nNo transactions recorded.")
					continue
				}

				// TODO: make this better using padding
				fmt.Printf("ID             Amount      Category     Description\n")

				// list of each transaction
				for _, e := range transactionList {
					fmt.Printf("%s |    €%-8.2f | %-10s | %-25s\n", e.Id, e.Amount, e.Category, e.Description)
				}

				fmt.Println()
			}
		}
	}

	return true, nil
}

func showTotal(args []string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	transactionTypes := []string{"expense", "investment", "income"}

	// TODO: this should be done better, expected behaviour :
	// no args prints everything for all months
	// arg1=month; arg2=year print a list of all transactions for a specific month and P&L for the month at the bottom
	// arg1=year print only a P&L for the year
	if len(args) == 2 {
		// TODO: add validation that both are provided and match a valid year and month
		month := args[0]
		year := args[1]
		var totalPnl PnLResult

		for _, txType := range transactionTypes {
			fmt.Printf("\n%s\n", txType)
			for i, t := range transactions[year][month][txType] {
				fmt.Printf("    %2d. €%-8.2f | %-10s | %-25s\n", i+1, t.Amount, t.Category, t.Description)
			}
		}

		totalPnl, err = calculateMonthPnL(month, year)
		if err != nil {
			return false, fmt.Errorf("Unable to calculate P&L: %s\n", err)
		}
		fmt.Printf("\np&l result: €%.2f | %.1f%%\n\n", totalPnl.Amount, totalPnl.Percent)
		return true, nil
	}

	if len(args) == 1 {
		year := args[0]
		var totalPnl PnLResult

		if totalPnl, err = calculateYearPnL(year); err != nil {
			return false, fmt.Errorf("Unable to calculate P&L: %s\n", err)
		}
		fmt.Printf("\np&l result: €%.2f | %.1f%%\n\n", totalPnl.Amount, totalPnl.Percent)
		return true, nil
	}

	if len(args) == 0 {
		for year, months := range transactions {
			// years
			fmt.Printf("\nYear: %s\n", year)

			// months
			for month := range months {
				fmt.Printf("  Month: %s\n\n", month)

				for _, t := range transactionTypes {
					if _, err := listTransactionsByMonth(t, month, year); err != nil {
						return false, fmt.Errorf("%s", err)
					}
				}

				calculatedPnL, err := calculateMonthPnL(month, year)
				if err != nil {
					return false, fmt.Errorf("Unable to calculate P&L: %s\n", err)
				}
				fmt.Printf("\np&l result: €%.2f | %.1f%%\n\n", calculatedPnL.Amount, calculatedPnL.Percent)
			}
		}
		return true, nil
	}

	return false, fmt.Errorf("invalid arguments provided to show-total: %s", args)
}

func listTransactionsByMonth(transactionType, month, year string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %w", loadFileErr)
	}

	if len(transactions) == 0 {
		fmt.Println("\nNo transactions found")
		return true, nil
	}

	transactionType, err = normalizeTransactionType(transactionType)
	if err != nil {
		return false, fmt.Errorf("transaction type normalization error: %s", err)
	}

	if _, ok := validTransactionTypes[transactionType]; !ok {
		return false, fmt.Errorf("invalid transaction type %s, please use expenses, income, or investments", transactionType)
	}

	fmt.Printf("    %s\n", transactionType)
	for i, t := range transactions[year][month][transactionType] {
		fmt.Printf("    %2d. €%-8.2f | %-10s | %-25s\n", i+1, t.Amount, t.Category, t.Description)
	}

	fmt.Println()

	return true, nil
}

func showAllowedCategories(transactionType string) error {
	fmt.Println("\nallowed categories are: ")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\nCategory\tDescription")
	fmt.Fprintln(w, "--------\t-----------")

	txType, err := normalizeTransactionType(transactionType)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	for key, val := range allowedTransactionCategories[txType] {
		fmt.Fprintf(w, "%s\t%s\n", key, val)
	}
	w.Flush()
	return nil
}
