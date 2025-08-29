package main

import (
	"fmt"

	"github.com/rivo/tview"
)

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

func generateControlsFooter() string {
	return "[yellow]ESC[-]/[yellow]q[-]: back   [green]TAB[-]: next   [cyan]j/k[-] or [cyan]↑/↓[-]: navigate"
}

func showErrorModal(msg string, frame *tview.Frame, focus tview.Primitive) {
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			// on presssing OK -  set focus back to the previous screen (menu, form, etc)
			tui.SetRoot(frame, true).SetFocus(focus)
		})
	// back to mainMenu on ESC or q key press
	modal.SetInputCapture(exitShortcuts)
	// set focus to the error
	tui.SetRoot(modal, true).SetFocus(modal)
}
