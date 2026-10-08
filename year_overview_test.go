package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestFormatMonthlyPnLTable(t *testing.T) {
	monthly := map[string]PnLResult{
		"january":  {incomeTotal: 3000, expenseTotal: 1000, investmentTotal: 500, pnlAmount: 2000, pnlPercent: 66.666},
		"february": {expenseTotal: 1000, pnlAmount: -1000},
	}
	year := PnLResult{incomeTotal: 3000, expenseTotal: 2000, investmentTotal: 500, pnlAmount: 1000, pnlPercent: 33.333}

	// months come in newest-first from getMonthsForYear; the table should be chronological
	out := formatMonthlyPnLTable([]string{"february", "january"}, monthly, year)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 5 { // header, 2 months, rule, total
		t.Fatalf("expected 5 lines, got %d:\n%s", len(lines), out)
	}
	if !strings.HasPrefix(lines[1], "January") || !strings.HasPrefix(lines[2], "February") {
		t.Errorf("expected months in calendar order, got:\n%s", out)
	}
	if !strings.Contains(lines[1], "66.7%") {
		t.Errorf("expected january savings rate 66.7%%, got %q", lines[1])
	}
	// no income -> rate is not meaningful, loss shown in red
	if !strings.Contains(lines[2], "n/a") || !strings.Contains(lines[2], negativeColor+"   -1000.00") {
		t.Errorf("expected red loss and n/a rate for february, got %q", lines[2])
	}
	if !strings.HasPrefix(lines[4], "Total") || !strings.Contains(lines[4], "33.3%") {
		t.Errorf("unexpected total row %q", lines[4])
	}

	if got := formatMonthlyPnLTable(nil, nil, PnLResult{}); got != "No transactions for this year" {
		t.Errorf("unexpected empty table output %q", got)
	}
}

func TestRenderYearPnL(t *testing.T) {
	pnl := PnLResult{incomeTotal: 3000, expenseTotal: 2000, investmentTotal: 500, pnlAmount: 1000, pnlPercent: 33.333}
	out := renderYearPnL(pnl)

	if lines := strings.Count(out, "\n"); lines != categoryPieHeight {
		t.Errorf("expected legend beside the pie (%d lines), got %d", categoryPieHeight, lines)
	}
	for _, part := range []string{"Income", "€   3000.00", " 54.5%", "Expenses", " 36.4%", "Investments", "  9.1%", positiveColor + "€   1000.00", "33.3%"} {
		if !strings.Contains(out, part) {
			t.Errorf("expected output to contain %q, got:\n%s", part, out)
		}
	}

	counts := countPieColors(generatePnLPieChart(pnl, categoryPieWidth, categoryPieHeight))
	if len(counts) != 3 || counts[incomeColor] <= counts[expenseColor] || counts[expenseColor] <= counts[investmentColor] {
		t.Errorf("unexpected pie slice sizes: %v", counts)
	}

	if empty := renderYearPnL(PnLResult{}); !strings.HasPrefix(empty, "No transactions for this period") {
		t.Errorf("unexpected empty output:\n%s", empty)
	}
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
		err = showYearResults("2026")
	})
	if err != nil {
		t.Fatalf("showYearResults: %v", err)
	}

	focusedTitle := func() string {
		if tv, ok := tui.GetFocus().(*tview.TextView); ok {
			return tv.GetTitle()
		}
		return ""
	}
	waitFor(t, "P&L panel focused", func() bool { return focusedTitle() == "Profit & Loss - 2026" })

	// TAB / l cycles forward, h goes back
	pressKey(screen, tcell.KeyTAB, 0)
	waitFor(t, "monthly panel focused", func() bool { return focusedTitle() == "Monthly Results - 2026" })
	pressKey(screen, tcell.KeyRune, 'l')
	waitFor(t, "breakdown panel focused", func() bool { return focusedTitle() == "Spending Breakdown - 2026 (Full Year)" })
	pressKey(screen, tcell.KeyTAB, 0)
	waitFor(t, "wrap to P&L panel", func() bool { return focusedTitle() == "Profit & Loss - 2026" })
	pressKey(screen, tcell.KeyRune, 'h')
	waitFor(t, "back to breakdown panel", func() bool { return focusedTitle() == "Spending Breakdown - 2026 (Full Year)" })

	// ESC returns to the year selector
	pressKey(screen, tcell.KeyEsc, 0)
	waitFor(t, "year selector", func() bool {
		name, _ := pages.GetFrontPage()
		return name == "yearSelector"
	})
}
