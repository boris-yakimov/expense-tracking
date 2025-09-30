package main

import (
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
func exitShortcutsWithPeriod(selectedMonth, selectedYear, focusTableType string) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			gridVisualizeTransactions(selectedMonth, selectedYear, focusTableType) // go back to the list of transactions (at the same month and year from where we came)
			return nil                                                             // key event consumed
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
