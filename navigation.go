package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// helper to handle vim-like motions when navigating the TUI - h, j, k, l
func vimMotions(event *tcell.EventKey) *tcell.EventKey {
	// rewrite the j/k call to a up or down arrow call instead to simulate vim motions
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j': // move down
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k': // move up
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case 'l': // tab
			return tcell.NewEventKey(tcell.KeyTAB, 0, tcell.ModNone)
		case 'h': // backtab (equivalent to shift+tab)
			return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
		}
	}
	return event
}

// helper to handle exit events - ESC, q, Q
func exitShortcuts(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
		return nil // key event consumed
	}
	return event // key event not consumed, so return it
}

// similar to exitShortcuts but accepts also a month and year to send back to when the key press is consumed
// returns a closure function around the scope of month, year that were passed
// the TUI will actually call the returned function when a key is pressed
// it is defined in this way because the form.SetInputCapture() expects a function as an arugment
// TODO: focusTableType is unused, what did we use this for before refactoring to the pages model ?
func exitShortcutsWithPeriod(selectedMonth, selectedYear, focusTableType string) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		// If user presses ESC or 'q'/'Q', decide whether to exit.
		if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			// If a text input field has focus, do not exit (allow typing to continue).
			if tui != nil {
				if focus := tui.GetFocus(); focus != nil {
					switch focus.(type) {
					case *tview.InputField:
						return event // consume nothing; keep focus for typing
					}
				}
			}

			// determine the correct page name to switch to with a safe fallback
			pageName := "main"
			if selectedMonth != "" && selectedYear != "" {
				candidate := fmt.Sprintf("main_%s_%s", selectedMonth, selectedYear)
				// Prefer the month-specific page if it exists, otherwise fall back to the generic main page
				if pages != nil && pages.HasPage(candidate) {
					pageName = candidate
				} else if pages != nil && pages.HasPage("main") {
					pageName = "main"
				} else {
					return nil
				}
			}
			// Remove the add-transaction modal if it's present to ensure clean focus restoration
			if pages != nil {
				pages.RemovePage("add-transaction")
			}
			pages.SwitchToPage(pageName) // go back to the list of transactions (at the same month and year from where we came)
			// Attempt to restore focus by re-rendering the main grid with the correct focus table
			if selectedMonth != "" && selectedYear != "" {
				if _, err := gridVisualizeTransactions(selectedMonth, selectedYear, focusTableType, true); err != nil {
					// ignore render errors; not fatal
				}
			}
			return nil
		}
		return event // key event not consumed, so return it
	}
}

// handle wrap around for table navigation (i.e. when last transaction reached wrap around to top)
func enableTableWrap(table *tview.Table) {
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, col := table.GetSelection()
		rowCount := table.GetRowCount()

		switch event.Key() {
		case tcell.KeyDown:
			if row == rowCount-1 { // at last row
				table.Select(0, col) // wrap to top
				return nil           // consume key event
			}
		case tcell.KeyUp:
			if row == 0 {
				table.Select(rowCount-1, col) // wrap to bottom
				return nil                    // consume key event
			}
		}

		return event
	})
}
