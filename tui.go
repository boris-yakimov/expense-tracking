package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// creates a footer for the TUI that shows navigation options
func generateCombinedControlsFooter() string {
	return Yellow + "ESC" + Reset + " /" +
		Yellow + "q" + Reset + ": back   " +
		Green + "TAB" + Reset + ": next   "
}

func generateWindowNavigationFooter() string {
	return Yellow + "ESC" + Reset + "/" +
		Yellow + "q" + Reset + ": back  " +
		Yellow + "m" + Reset + ": select month  " +
		Yellow + "y" + Reset + ": select year  " +
		Yellow + "TAB" + Reset + ": next table"
}

func generateTransactionCrudFooter() string {
	return Green + "a" + Reset + ": add  " +
		Red + "d" + Reset + ": delete  " +
		Yellow + "e/u" + Reset + ": update " +
		Blue + "/" + Reset + ": search"
}

func generateTransactionNavigationFooter() string {
	return Green + "j/k" + Reset + " or " + Green + "↑/↓" + Reset + ": move up and down  " +
		Green + "h/l" + Reset + " or " + Green + "←/→" + Reset + ": move left and right"
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
