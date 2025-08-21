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

	// Add vim-like navigation with j and k keys
	menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				// move down
				currentIndex := menu.GetCurrentItem()
				menu.SetCurrentItem(currentIndex + 1)
				return nil
			case 'k':
				// move up
				currentIndex := menu.GetCurrentItem()
				if currentIndex > 0 {
					menu.SetCurrentItem(currentIndex - 1)
				}
				return nil
			}
		}
		return event
	})

	tui.SetRoot(menu, true).SetFocus(menu)
	return nil
}
