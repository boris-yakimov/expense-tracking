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
				if _, err := gridVisualizeTransactions(selectedMonth, selectedYear); err != nil {
					showErrorModal(fmt.Sprintf("error showing transactions:\n\n%s", err), nil, list)
					return
				}
			}
		})
	}

	list.SetTitle("Select Month").
		SetTitleAlign(tview.AlignCenter).
		SetBorder(true)

	// navigation help
	frame := tview.NewFrame(list).
		AddText(generateCombinedControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)
	//
	// horizontal centering
	modal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).   // left spacer
		AddItem(frame, 60, 1, true). // form width fixed to fit text
		AddItem(nil, 0, 1, false))   // right spacer

	// vertical centering (height = 0 lets it fit content)
	centeredModal := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).  // top spacer
		AddItem(modal, 0, 1, true). // enough to fit the text
		AddItem(nil, 0, 1, false))  // bottom spacer

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

	tui.SetRoot(centeredModal, true).SetFocus(list)
	return nil
}

// creates a grid in the TUI to visualize and structure a list of transactions for a specific month and year
// if a month and year is provided will use it, otherwise will take the latest month
func gridVisualizeTransactions(selectedMonth, selectedYear string) (tview.Primitive, error) {
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
			return nil, fmt.Errorf("unable to determine last month or year: %w", err)
		}
	}

	transactions, err := LoadTransactions()
	if err != nil {
		return nil, fmt.Errorf("unable to load transactions file: %w", err)
	}

	var headerText string
	if displayMonth != "" && displayYear != "" {
		headerText = fmt.Sprintf("%s %s", capitalize(displayMonth), displayYear)
	}

	var calculatedPnl PnLResult
	var footerText string
	if calculatedPnl, err = calculateMonthPnL(displayMonth, displayYear); err != nil {
		return nil, fmt.Errorf("unable to calculate pnl: %w", err)
	}
	if displayMonth != "" && displayYear != "" {
		footerText = fmt.Sprintf("Income: €%.2f | Expenses: €%.2f | Investments: €%.2f \n\nP&L Result: €%.2f | %.1f%%", calculatedPnl.incomeTotal, calculatedPnl.expenseTotal, calculatedPnl.investmentTotal, calculatedPnl.pnlAmount, calculatedPnl.pnlPercent)
	}

	// build tx table for each tx type
	incomeTable := styleTable(createTransactionsTable("income", displayMonth, displayYear, transactions))
	expenseTable := styleTable(createTransactionsTable("expense", displayMonth, displayYear, transactions))
	investmentTable := styleTable(createTransactionsTable("investment", displayMonth, displayYear, transactions))

	// handle wrap around for table navigation (i.e. when last transaction reached wrap around to top)
	enableTableWrap(incomeTable)
	enableTableWrap(expenseTable)
	enableTableWrap(investmentTable)

	header := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(headerText)
	pnlFooter := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(footerText)
	helpLeftFooter := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(generateWindowNavigationFooter())

	helpCenterFooter := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(generateTransactionCrudFooter())

	helpRightFooter := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight).
		SetText(generateTransactionNavigationFooter())

	// nested footer grid with 2 columns
	footerGrid := styleGrid(tview.NewGrid().
		SetColumns(0, 0). // left + right
		AddItem(helpLeftFooter, 0, 0, 1, 1, 0, 0, false).
		AddItem(helpCenterFooter, 0, 1, 1, 1, 0, 0, false).
		AddItem(helpRightFooter, 0, 2, 1, 1, 0, 0, false))

	grid := styleGrid(tview.NewGrid().
		SetRows(3, 0, 3, 2).
		SetColumns(0, 0, 0).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(incomeTable, 1, 0, 1, 1, 0, 0, false).
		AddItem(expenseTable, 1, 1, 1, 1, 0, 0, false).
		AddItem(investmentTable, 1, 2, 1, 1, 0, 0, false).
		AddItem(pnlFooter, 2, 0, 1, 3, 0, 0, false).
		AddItem(footerGrid, 3, 0, 1, 3, 0, 0, false))
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

	return grid, nil
}
