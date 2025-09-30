package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var incomeSearch string
var expenseSearch string
var investmentSearch string

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
				if _, err := gridVisualizeTransactions(selectedMonth, selectedYear, "", true); err != nil {
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

	// horizontal centering
	modal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).   // left spacer
		AddItem(frame, 60, 1, true). // form width fixed to fit text
		AddItem(nil, 0, 1, false))   // right spacer

	// vertical centering (height = 0 lets it fit content with varying size)
	centeredModal := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).  // top spacer
		AddItem(modal, 0, 1, true). // enough to fit the text
		AddItem(nil, 0, 1, false))  // bottom spacer

	// handle input capture for month selection and exit
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// handle exit events
		if ev := exitShortcuts(event); ev == nil {
			// go back to the grid
			gridVisualizeTransactions("", "", "", true)
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

// creates a TUI window to show list of available years with transactions
func showYearSelector() error {
	years, err := getYearsWithTransactions()
	if err != nil {
		return fmt.Errorf("unable to get years with transactions: %w", err)
	}

	if len(years) == 0 {
		return fmt.Errorf("no transactions found")
	}

	list := styleList(tview.NewList())
	for _, year := range years {
		yearCopy := year // capture loop variable in each iteration because we need to use it in a closure for AddItem()
		list.AddItem(year, "", 0, func() {
			//yearCopy is used instead of year because closures in Go remember variables passed, not only their values and if we don't create a new variable in each loop iteration we get the same year in each iteration
			//That means if years = []{"2023", "2024", "2025"}:
			// You click 2023 → the callback sees year == "2025"
			// You click 2024 → the callback sees year == "2025"
			// You click 2025 → the callback sees year == "2025"
			if err := showYearResults(yearCopy); err != nil {
				showErrorModal(fmt.Sprintf("error showing year results:\n\n%s", err), nil, list)
				return
			}
		})
	}

	list.SetTitle("Select Year").
		SetTitleAlign(tview.AlignCenter).
		SetBorder(true)

	// navigation help
	frame := tview.NewFrame(list).
		AddText(generateCombinedControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// horizontal centering
	modal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).   // left spacer
		AddItem(frame, 60, 1, true). // form width fixed to fit text
		AddItem(nil, 0, 1, false))   // right spacer

	// vertical centering (height = 0 lets it fit content with varying size)
	centeredModal := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).  // top spacer
		AddItem(modal, 0, 1, true). // enough to fit the text
		AddItem(nil, 0, 1, false))  // bottom spacer

	// handle input capture for month selection and exit
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// handle exit events
		if ev := exitShortcuts(event); ev == nil {
			// go back to the grid
			gridVisualizeTransactions("", "", "", true)
			return nil // key event consumed
		}

		// handle list years event
		if event.Key() == tcell.KeyRune && event.Rune() == 'y' {
			if err := showYearSelector(); err != nil {
				showErrorModal(fmt.Sprintf("error showing year selector:\n\n%s", err), nil, list)
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
func gridVisualizeTransactions(selectedMonth, selectedYear, focusTableType string, setRoot bool) (tview.Primitive, error) {
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
	incomeTable := styleTable(createTransactionsTable("income", displayMonth, displayYear, transactions, incomeSearch))
	expenseTable := styleTable(createTransactionsTable("expense", displayMonth, displayYear, transactions, expenseSearch))
	investmentTable := styleTable(createTransactionsTable("investment", displayMonth, displayYear, transactions, investmentSearch))

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
	switch focusTableType {
	case "income":
		currentTable = 0
	case "expense":
		currentTable = 1
	case "investment":
		currentTable = 2
	default:
		currentTable = 0
	}

	var currentTableType string
	switch currentTable {
	case 0:
		currentTableType = "income"
	case 1:
		currentTableType = "expense"
	case 2:
		currentTableType = "investment"
	}

	// start with focus on the specified table
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
			if err := formAddTransaction(currentTableType, displayMonth, displayYear); err != nil {
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

			// month and year are passed here only for the purposes for the update transation form sending us back to the same month and year that we came from when we triggered it
			if err := formUpdateTransaction(txId, currentTableType, displayMonth, displayYear); err != nil {
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

			// month and year are passed here only for the purposes for the delete transation form sending us back to the same month and year that we came from when we triggered it
			if err := formDeleteTransaction(txId, currentTableType, displayMonth, displayYear); err != nil {
				showErrorModal(fmt.Sprintf("delete error:\n\n%s", err), nil, grid)
				return nil
			}
		}

		if event.Key() == tcell.KeyRune && event.Rune() == 'y' {
			if err := showYearSelector(); err != nil {
				showErrorModal(fmt.Sprintf("error showing year selector:\n\n%s", err), nil, grid)
				return nil
			}
			return nil // key event consumed
		}

		// enter search mode
		if event.Key() == tcell.KeyRune && event.Rune() == '/' {
			var currentSearch string
			switch currentTable {
			case 0:
				currentSearch = incomeSearch
			case 1:
				currentSearch = expenseSearch
			case 2:
				currentSearch = investmentSearch
			}

			searchInput := styleInputField(tview.NewInputField().
				SetLabel("Search: ").
				SetFieldWidth(30).
				SetText(currentSearch))

			searchInput.SetChangedFunc(func(text string) {
				// update the search
				switch currentTable {
				case 0:
					incomeSearch = text
				case 1:
					expenseSearch = text
				case 2:
					investmentSearch = text
				}
				// recreate the grid
				newGrid, err := gridVisualizeTransactions(displayMonth, displayYear, currentTableType, false)
				if err != nil {
					return // ignore error for dynamic update
				}
				flex := styleFlex(tview.NewFlex().SetDirection(tview.FlexRow))
				flex.AddItem(newGrid, 0, 1, false)
				flex.AddItem(searchInput, 1, 1, true) // show search prompt bellow the transaction grid
				tui.SetRoot(flex, true).SetFocus(searchInput)
			})

			searchInput.SetDoneFunc(func(key tcell.Key) {
				switch key {
				case tcell.KeyEnter: // keep focus on the search prompt
					tui.SetRoot(grid, true).SetFocus(tables[currentTable])

				case tcell.KeyEsc: // reset the search
					switch currentTable {
					case 0:
						incomeSearch = ""
					case 1:
						expenseSearch = ""
					case 2:
						investmentSearch = ""
					}

					// recreate the grid
					if _, err := gridVisualizeTransactions(displayMonth, displayYear, currentTableType, true); err != nil {
						showErrorModal(fmt.Sprintf("error refreshing grid:\n\n%s", err), nil, grid)
						return
					}
					tui.SetRoot(grid, true).SetFocus(tables[currentTable])
				}
			})

			// show the search input
			flex := tview.NewFlex().SetDirection(tview.FlexRow)
			flex.AddItem(grid, 0, 1, false)
			flex.AddItem(searchInput, 1, 1, true)
			tui.SetRoot(flex, true).SetFocus(searchInput)
			return nil
		}

		return event
	})

	// in cases like search where we still want to draw the grid of transactions but we want to focus on the search window instead of the grid
	if setRoot {
		tui.SetRoot(grid, true).SetFocus(tables[currentTable])
	}

	return grid, nil
}
