package main

import (
	"fmt"
	"os"
	"regexp"
	"slices"
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

	// extract and sort years
	var years []string
	for year := range transactions {
		years = append(years, year)
	}
	slices.Sort(years) // ascending order

	for _, year := range years {

		// extract and sort months
		var months []string
		for month := range transactions[year] {
			months = append(months, month)
		}
		slices.SortFunc(months, func(a, b string) int {
			return monthOrder[a] - monthOrder[b] // ascending order
		})

		for _, month := range months {
			fmt.Printf("%s %s\n\n", month, year)

			// extract and sort transaction types
			var transactionTypes []string
			for txType := range transactions[year][month] {
				transactionTypes = append(transactionTypes, txType)
			}
			slices.SortFunc(transactionTypes, func(a, b string) int {
				return transactionTypeOrder[a] - transactionTypeOrder[b] // ascending order
			})

			// expenses, investments, or income
			for _, transactionType := range transactionTypes {
				transactionList := transactions[year][month][transactionType]

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

			var calculatedPnl PnLResult
			if calculatedPnl, err = calculateMonthPnL(month, year); err != nil {
				return false, fmt.Errorf("Unable to calculate P&L: %s\n", err)
			}
			fmt.Printf("\np&l result: €%.2f | %.1f%%\n\n", calculatedPnl.Amount, calculatedPnl.Percent)
		}
	}

	return true, nil
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
		return false, fmt.Errorf("transaction type error: %s", err)
	}

	fmt.Printf("    %s\n", transactionType)
	for i, t := range transactions[year][month][transactionType] {
		fmt.Printf("    %2d. €%-8.2f | %-10s | %-25s\n", i+1, t.Amount, t.Category, t.Description)
	}

	fmt.Println()

	return true, nil
}

func visualizeTransactions(args []string) (success bool, err error) {
	var calculatedPnl PnLResult
	var year string
	var month string

	// list -  prints all transactions with P&L for each month
	if len(args) == 0 {
		if _, err := listAllTransactions(args); err != nil {
			return false, fmt.Errorf("%s", err)
		}
	}

	// list <year> - prints only P&L for the year
	if len(args) == 1 {
		re := regexp.MustCompile(`\b(20[0-9]{2}|2100)\b`)
		if re.MatchString(args[0]) {
			year = args[0]
		} else {
			return false, fmt.Errorf("invalid year %s", args[0])
		}

		if calculatedPnl, err = calculateYearPnL(year); err != nil {
			return false, fmt.Errorf("Unable to calculate P&L: %s\n", err)
		}
		fmt.Printf("\np&l result: €%.2f | %.1f%%\n\n", calculatedPnl.Amount, calculatedPnl.Percent)
		return true, nil
	}

	// list <month> <year> - prints tranasctions and P&L for a specific month
	if len(args) == 2 {
		if _, ok := monthOrder[args[0]]; !ok {
			return false, fmt.Errorf("invalid month %s", err)
		} else {
			month = args[0]
		}

		re := regexp.MustCompile(`\b(20[0-9]{2}|2100)\b`)
		if re.MatchString(args[1]) {
			year = args[1]
		} else {
			return false, fmt.Errorf("invalid year format provided %s", args[1])
		}

		for txType := range allowedTransactionTypes {
			if _, err := listTransactionsByMonth(txType, month, year); err != nil {
				return false, err
			}
		}

		if calculatedPnl, err = calculateMonthPnL(month, year); err != nil {
			return false, fmt.Errorf("Unable to calculate P&L: %s\n", err)
		}
		fmt.Printf("\np&l result: €%.2f | %.1f%%\n\n", calculatedPnl.Amount, calculatedPnl.Percent)
		return true, nil
	}

	return false, fmt.Errorf("invalid arguments provided to show-total: %s", args)
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
