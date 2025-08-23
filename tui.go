package main

import (
	"fmt"
	"os"

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
		fmt.Fprintf(os.Stderr, "failed to initialize main menu: %v\n", err)
		os.Exit(1)
	}

	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui failed: %v\n", err)
		os.Exit(1)
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

	// navigation help
	frame := tview.NewFrame(menu).
		AddText(generateControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// Add vim-like navigation with j and k keys
	menu.SetInputCapture(vimNavigation)

	tui.SetRoot(frame, true).SetFocus(menu)
	return nil
}

// TODO: implement show error modal window
