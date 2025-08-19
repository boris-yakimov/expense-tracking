package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func mainMenu() error {
	menu := tview.NewList().
		AddItem("list", "list transactions", 'l', func() {
			if _, err := listAllTransactions(); err != nil {
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

	tui.SetRoot(menu, true).SetFocus(menu)
	return tui.Run()
}

// TODO: add option for month year - default shows current, but if you start typing a previous month or year it is available based on the data you have
func formAddTransaction() error {
	var transactionType string
	dropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction Type").
		// TODO: probably should not be hardcoded
		SetOptions([]string{"income", "expense", "investment"}, func(selectedOption string, index int) {
			transactionType = selectedOption
		}))
	dropdown.SetCurrentOption(0)

	amountField := styleInputField(tview.NewInputField().SetLabel("Amount"))
	categoryField := styleInputField(tview.NewInputField().SetLabel("Category"))
	descriptionField := styleInputField(tview.NewInputField().SetLabel("Description"))

	form := styleForm(tview.NewForm().
		AddFormItem(dropdown).
		AddFormItem(amountField).
		AddFormItem(categoryField).
		AddFormItem(descriptionField).
		AddButton("Add", func() {
			amount := amountField.GetText()
			category := categoryField.GetText()
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
			amountField.SetText("")
			categoryField.SetText("")
			descriptionField.SetText("")
			dropdown.SetCurrentOption(0)
			transactionType = "expense"
		}).
		AddButton("Cancel", func() {
			mainMenu()
		}))

	form.SetButtonTextColor(tcell.ColorBlack).
		SetButtonBackgroundColor(tcell.ColorGreen).
		SetLabelColor(tcell.ColorYellow)

	form.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignLeft)

	tui.SetRoot(form, true).SetFocus(form)
	return nil
}

// func gridVisualizeTransactions() error {
// 	newPrimitive := func(text string) tview.Primitive {
// 		return tview.NewTextView().
// 			SetTextAlign(tview.AlignCenter).
// 			SetText(text)
// 	}
// 	leftScreen := newPrimitive("Income")
// 	middleScreen := newPrimitive("Expenses")
// 	rightScreen := newPrimitive("Investments")
//
// 	grid := tview.NewGrid().
// 		SetRows(3, 0, 3).
// 		SetColumns(30, 0, 30).
// 		SetBorders(true).
// 		AddItem(newPrimitive("July 2025"), 0, 0, 1, 3, 0, 0, false).
// 		AddItem(newPrimitive("press ESC to go back"), 2, 0, 1, 3, 0, 0, false)
//
// 	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
// 	grid.AddItem(leftScreen, 0, 0, 0, 0, 0, 0, false).
// 		AddItem(middleScreen, 1, 0, 1, 3, 0, 0, false).
// 		AddItem(rightScreen, 0, 0, 0, 0, 0, 0, false)
//
// 	// Layout for screens wider than 100 cells.
// 	grid.AddItem(leftScreen, 1, 0, 1, 1, 0, 100, false).
// 		AddItem(middleScreen, 1, 1, 1, 1, 0, 100, false).
// 		AddItem(rightScreen, 1, 2, 1, 1, 0, 100, false)
//
// 	if err := tview.NewApplication().SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
// 		return fmt.Errorf("tui error: %w", err)
// 	}
//
// 	return nil
// }
