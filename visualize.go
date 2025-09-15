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
				// go back to grid of visualized transactions but for the selected month and year
				if err := gridVisualizeTransactions(selectedMonth, selectedYear); err != nil {
					showErrorModal(fmt.Sprintf("error showing transactions:\n\n%s", err), nil, list)
					return
				}
			}
		})
	}

	// go back to previous month
	list.AddItem("back to current month", "", 'b', func() {
		// TODO: is passing around ("". "") the best way to do that, seems a bit wierd
		if err := gridVisualizeTransactions("", ""); err != nil {
			showErrorModal(fmt.Sprintf("error showing current transactions:\n\n%s", err), nil, list)
			return
		}
	})

	list.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	// navigation help
	frame := tview.NewFrame(list).
		AddText(generateControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// handle input capture for month selection and exit
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

// creates a grid in the TUI to visualize and structure a list of transactions for a specific month and year
// if a month and year is provided will use it, otherwise will take the latest month
func gridVisualizeTransactions(selectedMonth, selectedYear string) error {
	var displayMonth string
	var displayYear string
	var err error

	// if a month has been selected when rendering the grid use it, otherwise take the latest month
	if selectedMonth != "" || selectedYear != "" {
		displayMonth = selectedMonth
		displayYear = selectedYear
	} else {
		displayMonth, displayYear, err = determineLatestMonthAndYear()
		if err != nil {
			return fmt.Errorf("unable to determine last month or year: %w", err)
		}
	}

	transactions, err := LoadTransactions()
	if err != nil {
		return fmt.Errorf("unable to load transactions file: %w", err)
	}

	var headerText string
	if displayMonth != "" && displayYear != "" {
		headerText = fmt.Sprintf("%s %s", capitalize(displayMonth), displayYear)
	}

	var calculatedPnl PnLResult
	var footerText string
	if calculatedPnl, err = calculateMonthPnL(displayMonth, displayYear); err != nil {
		return fmt.Errorf("unable to calculate pnl: %w", err)
	}
	if displayMonth != "" && displayYear != "" {
		footerText = fmt.Sprintf("P&L Result: €%.2f | %.1f%%", calculatedPnl.Amount, calculatedPnl.Percent)
	}

	// build tx table for each tx type
	incomeTable := styleTable(createTransactionsTable("income", displayMonth, displayYear, transactions))
	expenseTable := styleTable(createTransactionsTable("expense", displayMonth, displayYear, transactions))
	investmentTable := styleTable(createTransactionsTable("investment", displayMonth, displayYear, transactions))

	header := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(headerText)
	pnlFooter := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(footerText)
	helpFooter := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		// TODO: separate helper function that does this
		// TODO: helper at the bottom of list transactions to show all options - a, d, e/u, j/k, tab, q, etc
		SetText("[yellow]ESC[-]/[yellow]q[-]: back   [green]m[-]: select month   " +
			"[cyan]j/k[-] or [cyan]↑/↓[-]: navigate rows   " +
			"[magenta]h/l[-] or [magenta]←/→[-] or [magenta]Tab/Shift+Tab[-]: switch tables")

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

	// TODO: modal in the bottom right that shows a temp message for a few sec with info like - successfully added, deleted, updated transactions, etc

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

		if event.Key() == tcell.KeyRune && event.Rune() == 'a' {
			currentTableType := ""
			switch currentTable {
			case 0:
				currentTableType = "income"
			case 1:
				currentTableType = "expense"
			case 2:
				currentTableType = "investment"
			}
			if err := formAddTransaction(currentTableType); err != nil {
				showErrorModal(fmt.Sprintf("add error:\n\n%s", err), nil, grid)
				return nil
			}
		}

		// TODO: pressing e or u opens updateTransaction form which than triggers handleUpdateTransaction() in which ever txType we were tabbed into (it gets automatically selected)
		// TODO: the update transaction window should show only the previously selected option and fields to change it, there should be no dropdowns to select other transactions in this window, only to change the current one
		if event.Key() == tcell.KeyRune && (event.Rune() == 'e' || event.Rune() == 'u') {
			row, col := tables[currentTable].GetSelection()
			cell := tables[currentTable].GetCell(row, col)
			txId, _ := cell.GetReference().(string)

			currentTableType := ""
			switch currentTable {
			case 0:
				currentTableType = "income"
			case 1:
				currentTableType = "expense"
			case 2:
				currentTableType = "investment"
			}

			if err := formUpdateTransaction(txId, currentTableType); err != nil {
				showErrorModal(fmt.Sprintf("update error:\n\n%s", err), nil, grid)
				return nil
			}
		}

		// TODO: pressing d prompts for confirmation to delete it and calls  handleDeleteTransaction()
		// TODO: maybe there is no need to show a separate form window at all, just provide details of the selected transaction and yes or no to confirm deletion its deletion
		// TODO: how do i fetch transaction ID from selected row ?
		if event.Key() == tcell.KeyRune && event.Rune() == 'd' {
			row, col := tables[currentTable].GetSelection()
			cell := tables[currentTable].GetCell(row, col)
			txId, _ := cell.GetReference().(string)

			currentTableType := ""
			switch currentTable {
			case 0:
				currentTableType = "income"
			case 1:
				currentTableType = "expense"
			case 2:
				currentTableType = "investment"
			}

			if err := formDeleteTransaction(txId, currentTableType); err != nil {
				showErrorModal(fmt.Sprintf("delete error:\n\n%s", err), nil, grid)
				return nil
			}
		}

		return event
	})

	return nil
}
