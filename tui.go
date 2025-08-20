package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TODO: apply theme to main menu as well
func mainMenu() error {
	menu := tview.NewList().
		AddItem("list", "list transactions", 'l', func() {
			if err := gridVisualizeTransactions(); err != nil {
				fmt.Printf("list transactions error: %s", err)
			}
		}).
		AddItem("add", "add a new transaction", 'a', func() {
			if err := formAddTransaction(); err != nil {
				fmt.Printf("add error: %s", err)
			}
		}).
		// AddItem("help", "display help menu", 'h', nil).
		AddItem("Quit", "press to exit", 'q', func() {
			tui.Stop()
		})

	menu.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	tui.SetRoot(menu, true).SetFocus(menu)
	return nil
}

// TODO: add option for month year - default shows current, but if you start typing a previous month or year it is available based on the data you have
func formAddTransaction() error {
	var transactionType string
	var category string
	var categoryDropdown *tview.DropDown

	allowedTransactionTypes, err := listOfAllowedTransactionTypes()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	typeDropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction Type").
		SetOptions(allowedTransactionTypes, func(selectedOption string, index int) {
			transactionType = selectedOption
			if categoryDropdown != nil {
				opts, err := listOfAllowedCategories(transactionType)
				if err != nil {
					fmt.Println(err)
				}
				categoryDropdown.SetOptions(opts, func(selectedOption string, index int) {
					category = selectedOption
				})
				if len(opts) > 0 {
					categoryDropdown.SetCurrentOption(0)
					category = opts[0]
				} else {
					category = ""
				}
			}
		}))
	typeDropdown.SetCurrentOption(0)

	if _, opt := typeDropdown.GetCurrentOption(); opt != "" {
		transactionType = opt
	}

	amountField := styleInputField(tview.NewInputField().SetLabel("Amount"))

	categoryDropdown = styleDropdown(tview.NewDropDown().
		SetLabel("Category"))
	// scope boundary to isolate opts and err from leaking in the rest of the function
	{
		opts, err := listOfAllowedCategories(transactionType)
		if err != nil {
			fmt.Println(err)
		}
		categoryDropdown.SetOptions(opts, func(selectedOption string, index int) {
			category = selectedOption
		})
	}

	categoryDropdown.SetCurrentOption(0)

	descriptionField := styleInputField(tview.NewInputField().SetLabel("Description"))

	// TODO: display footer that shows ESC or 'q' can be pressed to go back to menu
	form := styleForm(tview.NewForm().
		AddFormItem(typeDropdown).
		AddFormItem(amountField).
		AddFormItem(categoryDropdown).
		AddFormItem(descriptionField).
		AddButton("Add", func() {
			amount := amountField.GetText()
			description := descriptionField.GetText()

			// TODO: refactor add transactions to no longer expect cli args so we can just pass these cleanly
			if _, err := addTransaction([]string{transactionType, amount, category, description}); err != nil {
				// TODO: figure out how to better handle these errors
				fmt.Printf("failed to add transaction: %s", err)
				return
			}

			mainMenu() // go back to menu
		}).
		AddButton("Clear", func() {
			typeDropdown.SetCurrentOption(0)
			amountField.SetText("")
			categoryDropdown.SetCurrentOption(0)
			descriptionField.SetText("")
			transactionType = "expense"
		}).
		AddButton("Cancel", func() {
			mainMenu()
		}))

	form.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	// back to mainMenu on ESC or q key press
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			mainMenu()
			return nil
		}
		return event
	})

	tui.SetRoot(form, true).SetFocus(form)
	return nil
}

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

	// build tx table for each tx type
	incomeTable := createTransactionsTable("income", latestMonth, latestYear, transactions)
	expenseTable := createTransactionsTable("expense", latestMonth, latestYear, transactions)
	investmentTable := createTransactionsTable("investment", latestMonth, latestYear, transactions)

	header := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(headerText)
	footer := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("press ESC or 'q' to go back")

	// TODO: extend to include the P&L
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

// helper to build a table for a specific transaction type for visualization in the TUI
func createTransactionsTable(txType, month, year string, transactions TransactionHistory) *tview.Table {
	table := tview.NewTable()
	table.SetBorder(false)
	table.SetTitle(capitalize(txType)).SetBorder(true)

	headers := []string{"ID", "Amount", "Category", "Description"}
	for c, h := range headers {
		table.SetCell(0, c, tview.NewTableCell(h).SetSelectable(false))
	}

	if year == "" || month == "" {
		table.SetCell(1, 0, tview.NewTableCell("no transactions"))
		return table
	}

	txList := transactions[year][month][txType]
	if len(txList) == 0 {
		table.SetCell(1, 0, tview.NewTableCell("no transactions"))
		return table
	}

	for r, tx := range txList {
		table.SetCell(r+1, 0, tview.NewTableCell(fmt.Sprintf("%s    ", tx.Id)))
		table.SetCell(r+1, 1, tview.NewTableCell(fmt.Sprintf("€%.2f", tx.Amount)))
		table.SetCell(r+1, 2, tview.NewTableCell(tx.Category))
		table.SetCell(r+1, 3, tview.NewTableCell(tx.Description))
	}
	return table
}
