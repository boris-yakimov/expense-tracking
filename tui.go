package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// creates a footer for the TUI that shows navingation options
func generateCombinedControlsFooter() string {
	return "[yellow]ESC[-]/[yellow]q[-]: back   [green]TAB[-]: next   [cyan]j/k[-] or [cyan]↑/↓[-]: navigate"
	// return generateWindowNavigationFooter() + generateTransactionNavigationFooter()
}

func generateWindowNavigationFooter() string {
	return "[yellow]ESC[-]/[yellow]q[-]: back  " +
		"[yellow]m[-]: select month  " +
		"[yellow]y[-]: select year " +
		"[yellow]TAB[-]: next table"
}

func generateTransactionCrudFooter() string {
	return "[green]a[-]: add  " +
		"[red]d[-]: delete  " +
		"[yellow]e/u[-]: update"
}

func generateTransactionNavigationFooter() string {
	return "[green]j/k[-] or [green]↑/↓[-]: move up and down  " +
		"[green]h/l[-] or [green]←/→[-]: move left and right"
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
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			// go back to previous screen
			tui.SetRoot(previous, true).SetFocus(focus)
			return nil
		}
		return event
	})
	// set focus to the error
	tui.SetRoot(modal, true).SetFocus(modal)
}
