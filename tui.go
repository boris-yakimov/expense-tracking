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

func addTransactionForm() error {
	tui := tview.NewApplication()
}
