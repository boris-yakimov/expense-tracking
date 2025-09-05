package main

import (
	"fmt"
	"os"

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

	// navigation help and db status line
	frame = tview.NewFrame(menu).
		AddText(generateDbStatusLine(), false, tview.AlignLeft, theme.FieldTextColor).
		AddText(generateControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// Add vim-like navigation with j and k keys
	menu.SetInputCapture(vimNavigation)

	tui.SetRoot(frame, true).SetFocus(menu)
	return nil
}

func generateControlsFooter() string {
	return "[yellow]ESC[-]/[yellow]q[-]: back   [green]TAB[-]: next   [cyan]j/k[-] or [cyan]↑/↓[-]: navigate"
}

// shows DB encryption status
func generateDbStatusLine() string {
	if _, err := os.Stat(globalConfig.SQLitePath); err == nil && userPassword != "" {
		return "[green]DB status:[-] decrypted for session"
	}
	if _, err := os.Stat(encFile); err == nil {
		return "[cyan]DB status:[-] encrypted"
	}
	return "[yellow]DB status:[-] unknown"
}

// pop-up for error messages
func showErrorModal(msg string, previous tview.Primitive, focus tview.Primitive) {
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			// on presssing OK -  set focus back to the previous screen
			tui.SetRoot(previous, true).SetFocus(focus)
		})
	// back to mainMenu on ESC or q key press
	modal.SetInputCapture(exitShortcuts)
	// set focus to the error
	tui.SetRoot(modal, true).SetFocus(modal)
}
