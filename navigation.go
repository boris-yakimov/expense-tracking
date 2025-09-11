package main

import (
	"github.com/gdamore/tcell/v2"
)

// helper to handle vim-like motions when navigating the TUI
func vimNavigation(event *tcell.EventKey) *tcell.EventKey {
	// rewrite the j/k call to a up or down arrow call instead to simulate vim motions
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j': // move down
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k': // move up
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
	}
	return event
}

// helper to handle exit events - ESC, q, Q
func exitShortcuts(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
		// TODO: login to go directly to list transactions
		// mainMenu() // go back to menu
		gridVisualizeTransactions() // go back to list of transactions
		return nil                  // key event consumed
	}
	return event
}
