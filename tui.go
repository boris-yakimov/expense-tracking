package main

import (
	"os"

	"github.com/rivo/tview"
)

// creates a footer for the TUI that shows navingation options
// TODO: to be unitified with the updated approach from gridVisualizeTransactions()
func generateControlsFooter() string {
	return "[yellow]ESC[-]/[yellow]q[-]: back   [green]TAB[-]: next   [cyan]j/k[-] or [cyan]↑/↓[-]: navigate"
}

// shows DB encryption status
// TODO: to be moved to to list transactions or maybe removed altogether if it gets too cluttered
func generateDbStatusLine() string {
	if _, err := os.Stat(globalConfig.SQLitePath); err == nil && userPassword != "" {
		return "[green]DB status:[-] decrypted for session"
	}

	if _, err := os.Stat(encFile); err == nil {
		return "[cyan]DB status:[-] encrypted"
	}

	return "[yellow]DB status:[-] unknown"
}

// handles creating a pop-up for error messages in the TUI
func showErrorModal(msg string, previous tview.Primitive, focus tview.Primitive) {
	modal := styleModal(tview.NewModal().
		SetText(msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			// on presssing OK -  set focus back to the previous screen
			tui.SetRoot(previous, true).SetFocus(focus)
		}))
	// back to list of transactions on ESC or q key press
	modal.SetInputCapture(exitShortcuts)
	// set focus to the error
	tui.SetRoot(modal, true).SetFocus(modal)
}
