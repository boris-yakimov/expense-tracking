package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// creates a TUI window to show list of available months with transactions
func showMonthSelector() error {
	months, err := getMonthsWithTransactions()
	if err != nil {
		return fmt.Errorf("unable to get months with transactions: %w", err)
	}

	if len(months) == 0 {
		return fmt.Errorf("no transactions found")
	}

	list := styleList(tview.NewList())

	for _, monthYear := range months {
		list.AddItem(monthYear, "", 0, func() {
			// parse the selected month and year
			parts := strings.Split(monthYear, " ")
			if len(parts) == 2 {
				selectedMonth := parts[0]
				selectedYear := parts[1]
				if err := showTransactionsForMonth(selectedMonth, selectedYear); err != nil {
					showErrorModal(fmt.Sprintf("error showing transactions:\n\n%s", err), nil, list)
					return
				}
			}
		})
	}

	// go back to previous month
	list.AddItem("back to current month", "", 'b', func() {
		if err := gridVisualizeTransactions(); err != nil {
			showErrorModal(fmt.Sprintf("error showing current transactions:\n\n%s", err), nil, list)
			return
		}
	})

	list.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	// navigation help
	frame := tview.NewFrame(list).
		AddText(generateControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// Handle input capture for month selection and exit
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// handle exit events
		if ev := exitShortcuts(event); ev == nil {
			return nil // key event consumed
		}

		// handle list months event
		if event.Key() == tcell.KeyRune && event.Rune() == 'm' {
			if err := showMonthSelector(); err != nil {
				showErrorModal(fmt.Sprintf("error showing month selector:\n\n%s", err), nil, list)
				return nil
			}
			return nil // key event consumed
		}
		// handle j/k events to navigate up or down
		return vimMotions(event)
	})

	tui.SetRoot(frame, true).SetFocus(list)
	return nil
}

// populates the 3 TUI wndows in the list transactions section with details for the selected month and year
func showTransactionsForMonth(month, year string) error {
	transactions, err := LoadTransactions()
	if err != nil {
		return fmt.Errorf("unable to load transactions file: %w", err)
	}

	var headerText string
	if year != "" && month != "" {
		headerText = fmt.Sprintf("%s %s", capitalize(month), year)
	}

	var calculatedPnl PnLResult
	var footerText string
	if calculatedPnl, err = calculateMonthPnL(month, year); err != nil {
		return fmt.Errorf("unable to calculate pnl: %w", err)
	}
	if year != "" && month != "" {
		footerText = fmt.Sprintf("P&L Result: €%.2f | %.1f%%", calculatedPnl.Amount, calculatedPnl.Percent)
	}

	// build tx table for each tx type
	incomeTable := styleTable(createTransactionsTable("income", month, year, transactions))
	expenseTable := styleTable(createTransactionsTable("expense", month, year, transactions))
	investmentTable := styleTable(createTransactionsTable("investment", month, year, transactions))

	header := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(headerText)
	pnlFooter := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(footerText)
	helpFooter := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]ESC[-]/[yellow]q[-]: back   [green]m[-]: select month   " +
			"[cyan]j/k[-] or [cyan]↑/↓[-]: navigate rows   " +
			"[magenta]h/l[-] or [magenta]Tab/Shift+Tab[-]: switch tables")

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

	// Handle input capture for month selection and exit
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// handle exit events
		if ev := exitShortcuts(event); ev == nil {
			return nil // key event consumed
		}

		// handle list months event
		if event.Key() == tcell.KeyRune && event.Rune() == 'm' {
			if err := showMonthSelector(); err != nil {
				showErrorModal(fmt.Sprintf("error showing month selector:\n\n%s", err), nil, grid)
				return nil
			}
			return nil // key event consumed
		}
		// handle j/k events to navigate up or down
		return vimMotions(event)
	})

	tui.SetRoot(grid, true).SetFocus(grid)
	return nil
}

// TODO: tab to move from income to expenses to investments
// TODO: pressing e or u opens updateTransaction form which than triggers handleUpdateTransaction() in which ever txType we were tabbed into (it gets automatically selected)
// TODO: pressing d prompts for confirmation to delete it and calls  handleDeleteTransaction()
// TODO: pressing a opens addTransaction form which than triggers handleAddTransaction()
// TODO: press e or a to work as a modal isntead of a separate screen
// TODO: modal in the bottom right that shows a temp message for a few sec with info like - successfully added, deleted, updated transactions, etc

// creates a grid in the TUI to visualize and structure a list of transactions
func gridVisualizeTransactions() error {
	transactions, err := LoadTransactions()
	if err != nil {
		return fmt.Errorf("unable to load transactions file: %w", err)
	}

	latestMonth, latestYear, err := determineLatestMonthAndYear()
	if err != nil {
		return fmt.Errorf("unable to determine last month or year: %w", err)
	}

	var headerText string
	if latestYear != "" && latestMonth != "" {
		headerText = fmt.Sprintf("%s %s", capitalize(latestMonth), latestYear)
	}

	var calculatedPnl PnLResult
	var footerText string
	if calculatedPnl, err = calculateMonthPnL(latestMonth, latestYear); err != nil {
		return fmt.Errorf("unable to calculate pnl: %w", err)
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
		// TODO: separate helper function that does this
		// TODO: helper at the bottom of list transactions to show all options - a, d, e/u, j/k, tab, q, etc
		SetText("[yellow]ESC[-]/[yellow]q[-]: back   [green]m[-]: select month   " +
			"[cyan]j/k[-] or [cyan]↑/↓[-]: navigate rows   " +
			"[magenta]h/l[-] [magenta]←/→[-] or [magenta]Tab/Shift+Tab[-]: switch tables")

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

	// keep a list of tables for focus switching in the TUI
	tables := []*tview.Table{incomeTable, expenseTable, investmentTable}
	currentTable := 0 // index of which table is currently in focus

	// start with focus on incomeTable
	tui.SetRoot(grid, true).SetFocus(tables[currentTable])

	// handle input capture for navigation,
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// handle exit event
		if ev := exitShortcuts(event); ev == nil {
			tui.Stop()
			return nil // consume key event so nothing else see it
		}

		// handle j/k event to navigate up or down and h/l to navigate between tables
		event = vimMotions(event)

		// table switching with Tab / Shifit+Tab or arrow keys
		switch event.Key() {
		case tcell.KeyTAB, tcell.KeyRight:
			currentTable = (currentTable + 1) % len(tables) // % len(tables) wraps back to 0 when we reach the end of list of tables to prevenet out of bounds errors
			tui.SetFocus(tables[currentTable])
			return nil
		case tcell.KeyBacktab, tcell.KeyLeft: // Shift + Tab
			currentTable = (currentTable - 1 + len(tables)) % len(tables) // add +len(tables to prevent out of bounds when on first index and trying pressing to go back, if we don't do this we get index -1, when we do this we get 0-1+len(tables) which takes us to the last elemet of the table list instead of to -1
			tui.SetFocus(tables[currentTable])
			return nil
		}

		// handle list months event
		if event.Key() == tcell.KeyRune && event.Rune() == 'm' {
			if err := showMonthSelector(); err != nil {
				showErrorModal(fmt.Sprintf("error showing month selector:\n\n%s", err), nil, grid)
				return nil
			}
			return nil // key event consumed
		}

		return event
	})

	return nil
}
