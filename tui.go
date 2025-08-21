package main

import (
	"fmt"

	"github.com/rivo/tview"
)

// TODO: navigate up and down with vim keys as well

func mainMenu() error {
	menu := styleList(tview.NewList().
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
		AddItem("del", "delete a transaction", 'd', func() {
			if err := formDeleteTransaction(); err != nil {
				fmt.Printf("delete error: %s", err)
			}
		}).
		AddItem("update", "update a transaction", 'u', func() {
			if err := formUpdateTransaction(); err != nil {
				fmt.Printf("update error: %s", err)
			}
		}).
		AddItem("Quit", "press to exit", 'q', func() {
			tui.Stop()
		}))

	menu.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	tui.SetRoot(menu, true).SetFocus(menu)
	return nil
}
