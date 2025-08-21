package main

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	amountWidth      = 10
	categoryWidth    = 12
	descriptionWidth = 40
)

const (
	resetColour = "\033[0m"
	lightRed    = "\033[91m"
	lightGreen  = "\033[92m"
)

func gridVisualizeTransactions() error {
	transactions, err := loadTransactions()
	if err != nil {
		return fmt.Errorf("unable to load transactions file: %w", err)
	}

	// TODO: ability to pick a specific month or year
	// the default shows the current one
	// key press shows a list of months or years that have transactions
	// selecting one shows a table of transactions in that specific month and year

	// determine latest year
	var latestYear string
	for y := range transactions {
		if latestYear == "" || y > latestYear {
			latestYear = y
		}
	}

	// determine latest month for the year
	var latestMonth string
	if latestYear != "" {
		for m := range transactions[latestYear] {
			if latestMonth == "" || monthOrder[m] > monthOrder[latestMonth] {
				latestMonth = m
			}
		}
	}

	var headerText string
	if latestYear != "" && latestMonth != "" {
		headerText = fmt.Sprintf("%s %s", capitalize(latestMonth), latestYear)
	}

	var calculatedPnl PnLResult
	var footerText string
	if calculatedPnl, err = calculateMonthPnL(latestMonth, latestYear); err != nil {
		return fmt.Errorf("unable to calculate pnl %w", err)
	}
	if latestYear != "" && latestMonth != "" {
		footerText = fmt.Sprintf("P&L Result: €%.2f | %.1f%%\n\n", calculatedPnl.Amount, calculatedPnl.Percent)
	}

	// build tx table for each tx type
	incomeTable := styleTable(createTransactionsTable("income", latestMonth, latestYear, transactions))
	expenseTable := styleTable(createTransactionsTable("expense", latestMonth, latestYear, transactions))
	investmentTable := styleTable(createTransactionsTable("investment", latestMonth, latestYear, transactions))

	header := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(headerText)
	footer := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(footerText)

	grid := styleGrid(tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(0, 0, 0).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(footer, 2, 0, 1, 3, 0, 0, false).
		AddItem(incomeTable, 1, 0, 1, 1, 0, 0, false).
		AddItem(expenseTable, 1, 1, 1, 1, 0, 0, false).
		AddItem(investmentTable, 1, 2, 1, 1, 0, 0, false))

	grid.SetBorder(false).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	// back to mainMenu on ESC or q key press
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			mainMenu()
			return nil
		}
		return event
	})

	tui.SetRoot(grid, true).SetFocus(grid)

	return nil
}

// TODO: remove these after TUI approach is implemented

func listAllTransactions() (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
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
			fmt.Printf("\n====================\n%s %s\n====================\n\n", capitalize(month), year)

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

				// transaction type header
				fmt.Println()
				fmt.Printf("  %s\n", capitalize(transactionType))
				fmt.Printf("  %s\n", strings.Repeat("-", len(transactionType)))
				if len(transactionList) == 0 {
					fmt.Println("\nno transactions recorded.")
					continue
				}

				// table format
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
				fmt.Fprintln(w, "    ID\tAmount\tCategory\tDescription")
				fmt.Fprintln(w, "    --\t------\t--------\t-----------")

				for _, e := range transactionList {
					fmt.Fprintf(w, "    %s\t€%.2f\t%s\t%s\n", e.Id, e.Amount, e.Category, e.Description)
				}

				w.Flush()
				fmt.Println()
			}

			var calculatedPnl PnLResult
			if calculatedPnl, err = calculateMonthPnL(month, year); err != nil {
				return false, fmt.Errorf("unable to calculate P&L: %w\n", err)
			}

			var pnlColour string
			if calculatedPnl.Amount < 0 {
				pnlColour = lightRed
			} else {
				pnlColour = lightGreen
			}

			// non-color
			// fmt.Printf("P&L Result: €%.2f | %.1f%%\n\n", calculatedPnl.Amount, calculatedPnl.Percent)
			// color
			fmt.Printf("  P&L Result: %s€%.2f | %.1f%%%s\n\n", pnlColour, calculatedPnl.Amount, calculatedPnl.Percent, resetColour)
		}
	}

	return true, nil
}

func listTransactionsByMonth(transactionType, month, year string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	if len(transactions) == 0 {
		fmt.Println("\nno transactions found")
		return true, nil
	}

	transactionType, err = normalizeTransactionType(transactionType)
	if err != nil {
		return false, fmt.Errorf("transaction type error: %w", err)
	}

	// transaction type header
	fmt.Println()
	fmt.Printf("  %s\n", capitalize(transactionType))
	fmt.Printf("  %s\n", strings.Repeat("-", len(transactionType)))

	// transaction table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "    ID\tAmount\tCategory\tDescription")
	fmt.Fprintln(w, "    --\t------\t--------\t-----------")

	for _, t := range transactions[year][month][transactionType] {
		fmt.Fprintf(w, "    %s\t€%.2f\t%s\t%s\n", t.Id, t.Amount, t.Category, t.Description)
	}

	w.Flush()
	fmt.Println()

	return true, nil
}

// TODO: add flags --month / --year so that they can be passed in any order (still validate their content)
func visualizeTransactions(args []string) (success bool, err error) {
	var calculatedPnl PnLResult
	var year string
	var month string

	// list -  prints all transactions with P&L for each month
	if len(args) == 0 {
		if _, err := listAllTransactions(); err != nil {
			return false, fmt.Errorf("%w", err)
		}

		return true, nil
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
			return false, fmt.Errorf("unable to calculate P&L: %w\n", err)
		}
		fmt.Printf("\np&l result: €%.2f | %.1f%%\n\n", calculatedPnl.Amount, calculatedPnl.Percent)
		return true, nil
	}

	// list <month> <year> - prints tranasctions and P&L for a specific month
	if len(args) == 2 {
		if _, ok := monthOrder[args[0]]; !ok {
			return false, fmt.Errorf("invalid month %w", err)
		} else {
			month = args[0]
		}

		re := regexp.MustCompile(`\b(20[0-9]{2}|2100)\b`)
		if re.MatchString(args[1]) {
			year = args[1]
		} else {
			return false, fmt.Errorf("invalid year format provided %s", args[1])
		}

		fmt.Printf("\n====================\n%s %s\n====================\n\n", capitalize(month), year)

		for txType := range allowedTransactionTypes {
			if _, err := listTransactionsByMonth(txType, month, year); err != nil {
				return false, err
			}
		}

		if calculatedPnl, err = calculateMonthPnL(month, year); err != nil {
			return false, fmt.Errorf("unable to calculate P&L: %w\n", err)
		}

		var pnlColour string
		if calculatedPnl.Amount < 0 {
			pnlColour = lightRed
		} else {
			pnlColour = lightGreen
		}

		// non-color
		// fmt.Printf("  p&l result: €%.2f | %.1f%%\n\n", calculatedPnl.Amount, calculatedPnl.Percent)
		// color
		fmt.Printf("  P&L Result: %s€%.2f | %.1f%%%s\n\n", pnlColour, calculatedPnl.Amount, calculatedPnl.Percent, resetColour)
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
		return fmt.Errorf("%w", err)
	}

	for key, val := range allowedTransactionCategories[txType] {
		fmt.Fprintf(w, "%s\t%s\n", key, val)
	}
	w.Flush()
	return nil
}
