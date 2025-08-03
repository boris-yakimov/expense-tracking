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

// TODO: show total by passing a specific month or a year
func showTotal(args []string) (success bool, err error) {
	// essentially forcing args[0] to be a specific transaction type in order to list transactions inside
	if _, err := listTransactions([]string{"expenses"}); err != nil {
		return false, fmt.Errorf("%s", err)
	}

	if _, err := listTransactions([]string{"investments"}); err != nil {
		return false, fmt.Errorf("%s", err)
	}

	if _, err := listTransactions([]string{"income"}); err != nil {
		return false, fmt.Errorf("%s", err)
	}

	// TODO: print a nice summary with separate section for expenses, investments and income and a total p&l based on those

	calculatePnL()

	return true, nil
}

func listTransactions(args []string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("Unable to load transactions file: %w", loadFileErr)
	}

	if len(transactions) == 0 {
		// TODO: simulate this to see how the output looks like
		fmt.Println("No transactions found")
		return true, nil
	}

	var transactionType string
	if len(args) > 0 {
		transactionType = args[0]
		if _, ok := validTransactionTypes[transactionType]; !ok {
			return false, fmt.Errorf("invalid transaction type %s, please use expenses, income, or investments", transactionType)
		}
		// TODO: can i make this into a function so that i don't have to constantly do these cheks
		if transactionType == "expense" || transactionType == "expenses" {
			transactionType = "Expenses"
		}
		if transactionType == "investment" || transactionType == "investments" {
			transactionType = "Investments"
		}
		if transactionType == "income" {
			transactionType = "Income"
		}
	} else {
		transactionType = ""
	}

	// years
	for year, months := range transactions {
		fmt.Printf("\nYear: %s\n", year)

		// if only "list" command without args is passed, print all
		if transactionType == "" {
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
			// move to next month/year after we've printed all
			continue
		} else {
			for month, transactionTypes := range months {
				transactionList, ok := transactionTypes[transactionType]
				if !ok || len(transactionList) == 0 {
					continue // skip months with no data for the requested transaction type
				}

				fmt.Printf("  Month: %s\n\n", month)
				fmt.Printf("    %s\n", transactionType)

				// list of each transaction
				for i, t := range transactionList {
					fmt.Printf("    %2d. €%-8.2f | %-10s | %-25s\n", i+1, t.Amount, t.Category, t.Note)
				}

				fmt.Println()
			}
		}
		// move to next month/year after we've printed all
		continue
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
