package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestCreateMonthlyPnLTable(t *testing.T) {
	monthly := map[string]PnLResult{
		"january":  {incomeTotal: 3000, expenseTotal: 1000, investmentTotal: 500, pnlAmount: 2000, pnlPercent: 66.666},
		"february": {expenseTotal: 1000, pnlAmount: -1000},
	}
	year := PnLResult{incomeTotal: 3000, expenseTotal: 2000, investmentTotal: 500, pnlAmount: 1000, pnlPercent: 33.333}

	// months come in newest-first from getMonthsForYear; the table should be chronological
	table := createMonthlyPnLTable([]string{"february", "january"}, monthly, year)
	if rows := table.GetRowCount(); rows != 5 { // header, 2 months, spacer, total
		t.Fatalf("expected 5 rows, got %d", rows)
	}
	text := func(row, col int) string { return strings.TrimSpace(table.GetCell(row, col).Text) }

	if text(1, 0) != "January" || text(2, 0) != "February" {
		t.Errorf("expected months in calendar order, got %q, %q", text(1, 0), text(2, 0))
	}
	if ref, _ := table.GetCell(1, 3).GetReference().(string); ref != "january" {
		t.Errorf("month rows should reference their month, got %q", ref)
	}
	if text(1, 5) != "66.7%" {
		t.Errorf("expected january savings rate 66.7%%, got %q", text(1, 5))
	}
	// no income -> rate is not meaningful, loss shown in red
	fg, _, _ := table.GetCell(2, 4).Style.Decompose()
	if text(2, 5) != "n/a" || text(2, 4) != "-1000.00" || fg != tagColor(negativeColor) {
		t.Errorf("expected red loss and n/a rate for february, got %q / %q", text(2, 4), text(2, 5))
	}
	if text(4, 0) != "Total" || text(4, 5) != "33.3%" {
		t.Errorf("unexpected total row %q ... %q", text(4, 0), text(4, 5))
	}
	// only month rows can be selected, and the latest month is selected initially
	for _, row := range []int{0, 3, 4} {
		if table.GetCell(row, 0).NotSelectable == false {
			t.Errorf("row %d should not be selectable", row)
		}
	}
	if row, _ := table.GetSelection(); row != 2 {
		t.Errorf("expected latest month (row 2) selected, got row %d", row)
	}

	empty := createMonthlyPnLTable(nil, nil, PnLResult{})
	if got := empty.GetCell(0, 0).Text; got != "No transactions for this year" {
		t.Errorf("unexpected empty table output %q", got)
	}
}

func TestRenderYearPnL(t *testing.T) {
	// invested more than was saved -> "after investing" goes negative
	pnl := PnLResult{incomeTotal: 3000, expenseTotal: 1200, investmentTotal: 2100, pnlAmount: 1800, pnlPercent: 60}
	out := renderYearPnL(pnl)

	lineWith := func(label string) string {
		for _, l := range strings.Split(out, "\n") {
			if strings.Contains(l, label) {
				return l
			}
		}
		t.Fatalf("no line with %q in:\n%s", label, out)
		return ""
	}

	// percentages are relative to income, not to the sum of all flows
	if l := lineWith("Expenses"); !strings.Contains(l, "40.0%") {
		t.Errorf("expected expenses 40.0%% of income, got %q", l)
	}
	if l := lineWith("Investments"); !strings.Contains(l, "70.0%") {
		t.Errorf("expected investments 70.0%% of income, got %q", l)
	}
	if l := lineWith("Savings  "); !strings.Contains(l, positiveColor+"€   1800.00") || !strings.Contains(l, "60.0%") {
		t.Errorf("unexpected savings line %q", l)
	}
	if l := lineWith("After investing"); !strings.Contains(l, negativeColor+"€   -300.00") || !strings.Contains(l, "-10.0%") {
		t.Errorf("unexpected after investing line %q", l)
	}
	// shared scale: income is the largest flow, so its bar is full and expenses get 40% of it
	if l := lineWith("Income"); !strings.Contains(l, strings.Repeat("█", pnlBarWidth)) {
		t.Errorf("expected a full income bar, got %q", l)
	}
	if l := lineWith("Expenses"); !strings.Contains(l, strings.Repeat("█", 16)+Reset+strings.Repeat("░", 24)) {
		t.Errorf("expected expenses bar at 16/40, got %q", l)
	}

	if l := renderYearPnL(PnLResult{expenseTotal: 100, pnlAmount: -100}); !strings.Contains(l, "n/a") {
		t.Errorf("expected n/a percentages without income, got:\n%s", l)
	}
	if empty := renderYearPnL(PnLResult{}); empty != "No transactions for this period" {
		t.Errorf("unexpected empty output:\n%s", empty)
	}
}

func focusedTitle() string {
	switch p := tui.GetFocus().(type) {
	case *tview.TextView:
		return p.GetTitle()
	case *tview.Table:
		return p.GetTitle()
	}
	return ""
}

func TestYearResultsNavigation(t *testing.T) {
	setupTestStorage(t, StorageSQLite)
	if err := saveTransactionsToTestStorage(categoryBreakdownTestData()); err != nil {
		t.Fatalf("failed to set up test data: %v", err)
	}
	screen := startSimulatedTui(t)

	var err error
	onTui(func() {
		pages.AddPage("yearSelector", tview.NewList(), true, false)
		err = showYearResults("2026", nil)
	})
	if err != nil {
		t.Fatalf("showYearResults: %v", err)
	}

	waitFor(t, "monthly table focused", func() bool { return focusedTitle() == "Monthly Results - 2026" })

	// TAB / l cycles forward, h goes back
	pressKey(screen, tcell.KeyTAB, 0)
	waitFor(t, "breakdown panel focused", func() bool { return focusedTitle() == "Spending Breakdown - 2026 (Full Year)" })
	pressKey(screen, tcell.KeyRune, 'l')
	waitFor(t, "P&L panel focused", func() bool { return focusedTitle() == "Profit & Loss - 2026" })
	pressKey(screen, tcell.KeyTAB, 0)
	waitFor(t, "wrap to monthly table", func() bool { return focusedTitle() == "Monthly Results - 2026" })
	pressKey(screen, tcell.KeyRune, 'h')
	waitFor(t, "back to P&L panel", func() bool { return focusedTitle() == "Profit & Loss - 2026" })
	pressKey(screen, tcell.KeyRune, 'l')
	waitFor(t, "monthly table focused again", func() bool { return focusedTitle() == "Monthly Results - 2026" })

	// Enter opens the selected (latest) month, k moves to january first
	pressKey(screen, tcell.KeyRune, 'k')
	pressKey(screen, tcell.KeyEnter, 0)
	waitForBreakdown(t, "Spending Breakdown - January 2026")

	// ESC from the breakdown comes back to the year summary, ESC again to the year selector
	pressKey(screen, tcell.KeyEsc, 0)
	waitFor(t, "back on year summary", func() bool {
		name, _ := pages.GetFrontPage()
		return name == "yearResults" && focusedTitle() == "Monthly Results - 2026"
	})
	pressKey(screen, tcell.KeyEsc, 0)
	waitFor(t, "year selector", func() bool {
		name, _ := pages.GetFrontPage()
		return name == "yearSelector"
	})
}

func TestSpendingBreakdownOpensYearSummary(t *testing.T) {
	setupTestStorage(t, StorageSQLite)
	if err := saveTransactionsToTestStorage(categoryBreakdownTestData()); err != nil {
		t.Fatalf("failed to set up test data: %v", err)
	}
	screen := startSimulatedTui(t)

	var err error
	onTui(func() { err = showSpendingBreakdown("february", "2026", nil) })
	if err != nil {
		t.Fatalf("showSpendingBreakdown: %v", err)
	}
	waitForBreakdown(t, "Spending Breakdown - February 2026")

	// y jumps to the year summary, ESC returns to the same month breakdown
	pressKey(screen, tcell.KeyRune, 'y')
	waitFor(t, "year summary", func() bool {
		name, _ := pages.GetFrontPage()
		return name == "yearResults" && focusedTitle() == "Monthly Results - 2026"
	})
	pressKey(screen, tcell.KeyEsc, 0)
	waitForBreakdown(t, "Spending Breakdown - February 2026")
}
