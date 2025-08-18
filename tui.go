package main

import (
	"fmt"

	"github.com/rivo/tview"
)

func mainMenu() error {
	tui := tview.NewApplication()
	list := tview.NewList().
		AddItem("help", "display help menu", 'h', nil).
		AddItem("list", "list transactions", 'l', nil).
		AddItem("Quit", "press to exit", 'q', func() {
			tui.Stop()
		})
	if err := tui.SetRoot(list, true).SetFocus(list).Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	return nil
}

func formAddTransaction() error {
	tui := tview.NewApplication()
	form := tview.NewForm().
		AddDropDown("Transaction Type", []string{"income", "expense", "investment"}, 0, nil).
		AddInputField("Amount", "", 20, nil, nil).
		AddInputField("Category", "", 20, nil, nil).
		AddInputField("Description", "", 20, nil, nil).
		AddButton("Add", nil).
		AddButton("Cancel", nil)
	form.SetBorder(true).SetTitle("add transaction").SetTitleAlign(tview.AlignLeft)
	if err := tui.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	return nil
}
