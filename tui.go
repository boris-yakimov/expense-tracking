package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// creates a footer for the TUI that shows navigation options
func generateCombinedControlsFooter() string {
	return Yellow + "ESC" + Reset + " /" +
		Yellow + "q" + Reset + ": back   " +
		Green + "TAB" + Reset + ": next   " +
		Green + "j/k" + Reset + " or " + Green + "↑/↓" + Reset + ": navigate"
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
		AddButtons([]string{"OK"}))

	// handle closing (OK, ESC, or 'q')
	closeModal := func() {
		// remove the modal page and go back to previous
		pages.RemovePage("errorModal")
		tui.SetFocus(focus)
	}

	modal.SetDoneFunc(func(_ int, _ string) {
		closeModal()
	})

	// back to list of transactions on ESC or q key press
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			// go back to previous screen
			pages.RemovePage("errorModal")
			tui.SetFocus(focus)
			return nil
		}
		return event
	})

	// instead of replacing the entire root, overlay the modal on top of the existing screen
	overlay := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(modal, 10, 1, true). // modal height
		AddItem(nil, 0, 1, false)

	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(overlay, 60, 1, true). // modal width
		AddItem(nil, 0, 1, false)

	// add modal as a page to overlay on top of existing content
	pages.AddPage("errorModal", centered, true, true)
	tui.SetFocus(modal)
}
