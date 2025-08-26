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
	var frame *tview.Frame
	var menu *tview.List

	menu = styleList(tview.NewList().
		AddItem("list transactions", "", 'l', func() {
			if err := gridVisualizeTransactions(); err != nil {
				showErrorModal(fmt.Sprintf("list error:\n\n%s", err), frame, menu)
				return
			}
		}).
		AddItem("add a new transaction", "", 'a', func() {
			if err := formAddTransaction(); err != nil {
				showErrorModal(fmt.Sprintf("add error:\n\n%s", err), frame, menu)
				return
			}
		}).
		AddItem("delete a transaction", "", 'd', func() {
			if err := formDeleteTransaction(); err != nil {
				showErrorModal(fmt.Sprintf("delete error:\n\n%s", err), frame, menu)
				return
			}
		}).
		AddItem("update a transaction", "", 'u', func() {
			if err := formUpdateTransaction(); err != nil {
				showErrorModal(fmt.Sprintf("update error:\n\n%s", err), frame, menu)
				return
			}
		}).
		AddItem("quit", "", 'q', func() {
			tui.Stop()
		}))

	menu.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	// navigation help
	frame = tview.NewFrame(menu).
		AddText(generateControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// Add vim-like navigation with j and k keys
	menu.SetInputCapture(vimNavigation)

	tui.SetRoot(frame, true).SetFocus(menu)
	return nil
}
