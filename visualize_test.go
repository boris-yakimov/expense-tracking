package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestShowAllowedCategories(t *testing.T) {
	cases := []struct {
		name            string
		transactionType string
		expectedError   bool
	}{
		{
			name:            "show expense categories",
			transactionType: "expense",
			expectedError:   false,
		},
		{
			name:            "show income categories",
			transactionType: "income",
			expectedError:   false,
		},
		{
			name:            "show investment categories",
			transactionType: "investment",
			expectedError:   false,
		},
		{
			name:            "invalid transaction type",
			transactionType: "invalidtype",
			expectedError:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := listOfAllowedCategories(c.transactionType)

			if (err != nil) != c.expectedError {
				t.Errorf("showAllowedCategories(%q) error = %v; expected error = %v",
					c.transactionType, err, c.expectedError)
			}
		})
	}
}

// returning to the transactions grid from another page (e.g. ESC on the month selector) used to focus the grid itself,
// so up/down scrolled the whole grid out of the window instead of moving the selection inside the focused table
func TestGridKeepsFocusOnTableAfterReturning(t *testing.T) {
	setupTestStorage(t, StorageSQLite)
	if err := saveTransactionsToTestStorage(categoryBreakdownTestData()); err != nil {
		t.Fatalf("failed to set up test data: %v", err)
	}
	screen := startSimulatedTui(t)

	var grid *tview.Grid
	var err error
	rendered := make(chan struct{})
	tui.QueueUpdateDraw(func() {
		var p tview.Primitive
		p, err = gridVisualizeTransactions("", "", "expense", true)
		grid, _ = p.(*tview.Grid)
		close(rendered)
	})
	<-rendered
	if err != nil || grid == nil {
		t.Fatalf("gridVisualizeTransactions: %v", err)
	}

	// grid items only count as focused once the grid has been drawn
	waitFor(t, "grid drawn", func() bool { return grid.HasFocus() })

	// open and close the month selector, which switches back to the "main" page
	pressKey(screen, tcell.KeyRune, 'm')
	waitFor(t, "month selector", func() bool { name, _ := pages.GetFrontPage(); return name == "monthSelector" })
	pressKey(screen, tcell.KeyEsc, 0)
	waitFor(t, "main page", func() bool { name, _ := pages.GetFrontPage(); return name == "main" })

	for i := 0; i < 3; i++ {
		pressKey(screen, tcell.KeyDown, 0)
		pressKey(screen, tcell.KeyRune, 'j')
	}

	waitFor(t, "focus on a table", func() bool { _, ok := tui.GetFocus().(*tview.Table); return ok })
	var rowOffset int
	onTui(func() { rowOffset, _ = grid.GetOffset() })
	if rowOffset != 0 {
		t.Errorf("expected grid not to scroll, got row offset %d", rowOffset)
	}
}
