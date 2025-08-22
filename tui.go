package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var tui *tview.Application

func main() {
	tui = tview.NewApplication()
	tui.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		screen.Clear()
		screen.Fill(' ', tcell.StyleDefault.Background(theme.BackgroundColor))
		return false
	})

	if err := mainMenu(); err != nil {
		panic(err)
	}

	if err := tui.Run(); err != nil {
		panic(err)
	}
}

func mainMenu() error {
	menu := styleList(tview.NewList().
		AddItem("list transactions", "", 'l', func() {
			if err := gridVisualizeTransactions(); err != nil {
				fmt.Printf("list transactions error: %s", err)
			}
		}).
		AddItem("add a new transaction", "", 'a', func() {
			if err := formAddTransaction(); err != nil {
				fmt.Printf("add error: %s", err)
			}
		}).
		AddItem("delete a transaction", "", 'd', func() {
			if err := formDeleteTransaction(); err != nil {
				fmt.Printf("delete error: %s", err)
			}
		}).
		AddItem("update a transaction", "", 'u', func() {
			if err := formUpdateTransaction(); err != nil {
				fmt.Printf("update error: %s", err)
			}
		}).
		AddItem("quit", "", 'q', func() {
			tui.Stop()
		}))

	menu.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	// Add vim-like navigation with j and k keys
	menu.SetInputCapture(vimNavigation)

	tui.SetRoot(menu, true).SetFocus(menu)
	return nil
}
