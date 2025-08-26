package main

import (
	"fmt"

	"github.com/rivo/tview"
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

	// TODO: can I move those into a separate function
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
		footerText = fmt.Sprintf("P&L Result: €%.2f | %.1f%%", calculatedPnl.Amount, calculatedPnl.Percent)
	}

	// build tx table for each tx type
	incomeTable := styleTable(createTransactionsTable("income", latestMonth, latestYear, transactions))
	expenseTable := styleTable(createTransactionsTable("expense", latestMonth, latestYear, transactions))
	investmentTable := styleTable(createTransactionsTable("investment", latestMonth, latestYear, transactions))

	header := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(headerText)
	pnlFooter := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(footerText)
	helpFooter := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(generateControlsFooter())

	// TODO: explore options to redesign grid into a flex
	grid := styleGrid(tview.NewGrid().
		SetRows(3, 0, 3, 2).
		SetColumns(0, 0, 0).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(incomeTable, 1, 0, 1, 1, 0, 0, false).
		AddItem(expenseTable, 1, 1, 1, 1, 0, 0, false).
		AddItem(investmentTable, 1, 2, 1, 1, 0, 0, false)).
		AddItem(pnlFooter, 2, 0, 1, 3, 0, 0, false).
		AddItem(helpFooter, 3, 0, 1, 3, 0, 0, false)
	grid.SetBorder(false).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	// back to mainMenu on ESC or q key press
	grid.SetInputCapture(exitShortcuts)

	tui.SetRoot(grid, true).SetFocus(grid)

	return nil
}
