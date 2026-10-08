package main

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func categoryBreakdownTestData() TransactionHistory {
	return TransactionHistory{
		"2026": {
			"january": {
				"income": {
					{Id: "1", Amount: 3000.00, Category: "salary", Description: "salary"},
				},
				"expense": {
					{Id: "2", Amount: 300.00, Category: "food", Description: "groceries"},
					{Id: "3", Amount: 100.00, Category: "food", Description: "restaurant"},
					{Id: "4", Amount: 400.00, Category: "car", Description: "fuel"},
					{Id: "5", Amount: 200.00, Category: "shopping", Description: "clothes"},
				},
				"investment": {
					{Id: "6", Amount: 500.00, Category: "stocks", Description: "etf"},
				},
			},
			"february": {
				"expense": {
					{Id: "7", Amount: 600.00, Category: "food", Description: "groceries"},
					{Id: "8", Amount: 400.00, Category: "bills", Description: "electricity"},
				},
			},
		},
		"2025": {
			"december": {
				"expense": {
					{Id: "9", Amount: 999.00, Category: "food", Description: "other year"},
				},
			},
		},
	}
}

func assertCategory(t *testing.T, ct CategoryTotal, category string, total, percent float64) {
	t.Helper()
	if ct.category != category || math.Abs(ct.total-total) > 1e-9 || math.Abs(ct.percent-percent) > 1e-9 {
		t.Errorf("got {%s %.2f %.2f%%}, expected {%s %.2f %.2f%%}",
			ct.category, ct.total, ct.percent, category, total, percent)
	}
}

func TestBuildExpenseCategoryBreakdownMonth(t *testing.T) {
	b := buildExpenseCategoryBreakdown(categoryBreakdownTestData(), "2026", []string{"january"})

	if b.expenseTotal != 1000.00 {
		t.Fatalf("expenseTotal = %.2f; expected 1000.00 (income and investments must be excluded)", b.expenseTotal)
	}
	if len(b.categories) != 3 {
		t.Fatalf("got %d categories; expected 3", len(b.categories))
	}
	// food and car tie at 400 -> alphabetical: car first
	assertCategory(t, b.categories[0], "car", 400, 40)
	assertCategory(t, b.categories[1], "food", 400, 40)
	assertCategory(t, b.categories[2], "shopping", 200, 20)
}

func TestBuildExpenseCategoryBreakdownYear(t *testing.T) {
	b := buildExpenseCategoryBreakdown(categoryBreakdownTestData(), "2026", []string{"january", "february"})

	if b.expenseTotal != 2000.00 {
		t.Fatalf("expenseTotal = %.2f; expected 2000.00", b.expenseTotal)
	}
	if len(b.categories) != 4 {
		t.Fatalf("got %d categories; expected 4", len(b.categories))
	}
	assertCategory(t, b.categories[0], "food", 1000, 50)
	assertCategory(t, b.categories[1], "bills", 400, 20)
	assertCategory(t, b.categories[2], "car", 400, 20)
	assertCategory(t, b.categories[3], "shopping", 200, 10)
}

func TestBuildExpenseCategoryBreakdownEmpty(t *testing.T) {
	b := buildExpenseCategoryBreakdown(categoryBreakdownTestData(), "2030", []string{"january"})
	if b.expenseTotal != 0 || len(b.categories) != 0 {
		t.Errorf("expected empty breakdown, got %+v", b)
	}
	if got := formatCategoryBreakdown(b); got != "No expenses for this period" {
		t.Errorf("unexpected empty format output: %q", got)
	}
}

func TestCalculateCategoryBreakdownFromStorage(t *testing.T) {
	setupTestStorage(t, StorageSQLite)
	if err := saveTransactionsToTestStorage(categoryBreakdownTestData()); err != nil {
		t.Fatalf("failed to set up test data: %v", err)
	}

	month, err := calculateMonthCategoryBreakdown("february", "2026")
	if err != nil {
		t.Fatalf("calculateMonthCategoryBreakdown error: %v", err)
	}
	if month.expenseTotal != 1000 || len(month.categories) != 2 {
		t.Errorf("unexpected february breakdown: %+v", month)
	}
	assertCategory(t, month.categories[0], "food", 600, 60)

	year, err := calculateYearCategoryBreakdown("2026")
	if err != nil {
		t.Fatalf("calculateYearCategoryBreakdown error: %v", err)
	}
	if year.expenseTotal != 2000 {
		t.Errorf("year expenseTotal = %.2f; expected 2000 (other years must be excluded)", year.expenseTotal)
	}
}

func TestFormatCategoryBreakdown(t *testing.T) {
	b := buildExpenseCategoryBreakdown(categoryBreakdownTestData(), "2026", []string{"january", "february"})
	out := formatCategoryBreakdown(b)

	for _, part := range []string{"Food", "€   1000.00", " 50.0%", "Bills", "Shopping", " 10.0%", "Total", "€   2000.00"} {
		if !strings.Contains(out, part) {
			t.Errorf("expected output to contain %q, got:\n%s", part, out)
		}
	}
	// 50% of a 20-char bar = 10 filled blocks, in the first palette color (largest category)
	if !strings.Contains(out, categoryPalette[0]+strings.Repeat("█", 10)+Reset+strings.Repeat("░", 10)) {
		t.Errorf("expected half-filled bar for food, got:\n%s", out)
	}
}

// counts how many pie cells are drawn in each color tag
func countPieColors(pie string) map[string]int {
	counts := make(map[string]int)
	for _, cell := range strings.Split(pie, "█"+Reset) {
		if i := strings.LastIndex(cell, "[#"); i >= 0 {
			counts[cell[i:]]++
		}
	}
	return counts
}

func TestGenerateCategoryPieChart(t *testing.T) {
	b := buildExpenseCategoryBreakdown(categoryBreakdownTestData(), "2026", []string{"january", "february"})
	pie := generateCategoryPieChart(b, categoryPieWidth, categoryPieHeight)

	if lines := strings.Count(pie, "\n"); lines != categoryPieHeight {
		t.Errorf("expected %d pie lines, got %d", categoryPieHeight, lines)
	}

	counts := countPieColors(pie)
	var total int
	for _, c := range counts {
		total += c
	}
	if len(counts) != 4 {
		t.Fatalf("expected 4 colored slices, got %d: %v", len(counts), counts)
	}
	// slice areas should roughly follow the spend shares: food 50%, bills 20%, car 20%, shopping 10%
	expected := []float64{0.5, 0.2, 0.2, 0.1}
	for i, share := range expected {
		got := float64(counts[categoryPalette[i]]) / float64(total)
		if math.Abs(got-share) > 0.05 {
			t.Errorf("slice %d (%s) share = %.2f; expected ~%.2f", i, b.categories[i].category, got, share)
		}
	}
}

func TestGenerateCategoryPieChartGroupsOther(t *testing.T) {
	// more categories than palette colors -> overflow is grouped into a single gray slice
	var b CategoryBreakdown
	for i := 0; i < len(categoryPalette)+3; i++ {
		b.categories = append(b.categories, CategoryTotal{category: fmt.Sprintf("c%02d", i), total: 10})
		b.expenseTotal += 10
	}
	counts := countPieColors(generateCategoryPieChart(b, categoryPieWidth, categoryPieHeight))
	if len(counts) != len(categoryPalette)+1 {
		t.Errorf("expected %d colors (palette + other), got %d: %v", len(categoryPalette)+1, len(counts), counts)
	}
	if counts[categoryOtherColor] == 0 {
		t.Errorf("expected an 'other' slice in %s", categoryOtherColor)
	}
}

func TestGenerateCategoryPieChartEmpty(t *testing.T) {
	if pie := generateCategoryPieChart(CategoryBreakdown{}, categoryPieWidth, categoryPieHeight); pie != "" {
		t.Errorf("expected empty pie for no expenses, got:\n%s", pie)
	}
}

// runs the TUI on a simulated screen so key presses go through the real input capture chain
func startSimulatedTui(t *testing.T) tcell.SimulationScreen {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("simulation screen init: %v", err)
	}
	screen.SetSize(160, 50)

	tui = tview.NewApplication().SetScreen(screen)
	pages = tview.NewPages()
	tui.SetRoot(pages, true)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = tui.Run()
	}()
	t.Cleanup(func() {
		tui.Stop()
		<-done
	})
	return screen
}

// runs fn on the TUI goroutine and waits for it, so state reads don't race with event handling
func onTui(fn func()) {
	finished := make(chan struct{})
	tui.QueueUpdate(func() { fn(); close(finished) })
	<-finished
}

func pressKey(screen tcell.SimulationScreen, key tcell.Key, r rune) {
	// key events keep their order relative to each other, but not relative to QueueUpdate,
	// so assertions after key presses should use waitFor
	screen.InjectKey(key, r, tcell.ModNone)
}

// polls until cond is true (evaluated on the TUI goroutine) or fails the test after a timeout
func waitFor(t *testing.T, desc string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var ok bool
		onTui(func() { ok = cond() })
		if ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for: %s (front page %q, month title %q)", desc, frontPageName(), monthTitle())
}

func frontPageName() string {
	var name string
	onTui(func() { name, _ = pages.GetFrontPage() })
	return name
}

// title of the month panel (first panel) on the breakdown page; must be called on the TUI goroutine or via onTui
func breakdownMonthTitleUnsafe() string {
	_, page := pages.GetFrontPage()
	frame, ok := page.(*tview.Frame)
	if !ok {
		return ""
	}
	flex, ok := frame.GetPrimitive().(*tview.Flex)
	if !ok {
		return ""
	}
	return flex.GetItem(0).(*tview.TextView).GetTitle()
}

func monthTitle() string {
	var title string
	onTui(func() { title = breakdownMonthTitleUnsafe() })
	return title
}

// waits until the breakdown page is in front showing the given month title
func waitForBreakdown(t *testing.T, title string) {
	t.Helper()
	waitFor(t, "breakdown "+title, func() bool {
		name, _ := pages.GetFrontPage()
		return name == "spendingBreakdown" && breakdownMonthTitleUnsafe() == title
	})
}

func waitForMonthSelector(t *testing.T) {
	t.Helper()
	waitFor(t, "month selector", func() bool {
		name, _ := pages.GetFrontPage()
		_, isList := tui.GetFocus().(*tview.List)
		return name == "breakdownMonthSelector" && isList
	})
}

func TestSpendingBreakdownMonthSelection(t *testing.T) {
	setupTestStorage(t, StorageSQLite)
	if err := saveTransactionsToTestStorage(categoryBreakdownTestData()); err != nil {
		t.Fatalf("failed to set up test data: %v", err)
	}
	screen := startSimulatedTui(t)

	var err error
	onTui(func() { err = showSpendingBreakdown("january", "2026", nil) })
	if err != nil {
		t.Fatalf("showSpendingBreakdown: %v", err)
	}
	waitForBreakdown(t, "Spending Breakdown - January 2026")

	// open the selector - months of 2026 only, newest first, starting on the current month (january)
	pressKey(screen, tcell.KeyRune, 'm')
	waitForMonthSelector(t)
	var items []string
	var selected int
	onTui(func() {
		list := tui.GetFocus().(*tview.List)
		for i := 0; i < list.GetItemCount(); i++ {
			main, _ := list.GetItemText(i)
			items = append(items, main)
		}
		selected = list.GetCurrentItem()
	})
	if strings.Join(items, ",") != "February,January" {
		t.Errorf("unexpected months in selector: %v", items)
	}
	if selected != 1 {
		t.Errorf("expected selection to start on january (index 1), got %d", selected)
	}

	// ESC closes the selector without changing the month
	pressKey(screen, tcell.KeyEsc, 0)
	waitForBreakdown(t, "Spending Breakdown - January 2026")

	// select february with vim motion + enter
	pressKey(screen, tcell.KeyRune, 'm')
	pressKey(screen, tcell.KeyRune, 'k')
	pressKey(screen, tcell.KeyEnter, 0)
	waitForBreakdown(t, "Spending Breakdown - February 2026")
}

func TestBuildMonthComparison(t *testing.T) {
	// january 2026: car 400, food 400, shopping 200; february: food 600, bills 400 -> 2 months to average over
	cmp := buildMonthComparison(categoryBreakdownTestData(), "january", "2026")

	if cmp.monthsInYear != 2 || cmp.avgExpenseTotal != 1000 || cmp.month.expenseTotal != 1000 {
		t.Fatalf("unexpected totals: months=%d avg=%.2f month=%.2f", cmp.monthsInYear, cmp.avgExpenseTotal, cmp.month.expenseTotal)
	}

	expected := []struct {
		category   string
		total, avg float64
	}{
		{"car", 400, 200},
		{"food", 400, 500},
		{"shopping", 200, 100},
		{"bills", 0, 200}, // only spent in february -> listed last with zero this month
	}
	if len(cmp.rows) != len(expected) {
		t.Fatalf("got %d rows, expected %d: %+v", len(cmp.rows), len(expected), cmp.rows)
	}
	for i, e := range expected {
		r := cmp.rows[i]
		if r.category != e.category || r.total != e.total || r.monthlyAvg != e.avg {
			t.Errorf("row %d = {%s %.2f avg %.2f}, expected {%s %.2f avg %.2f}", i, r.category, r.total, r.monthlyAvg, e.category, e.total, e.avg)
		}
	}

	if empty := buildMonthComparison(categoryBreakdownTestData(), "january", "2030"); len(empty.rows) != 0 || empty.monthsInYear != 0 {
		t.Errorf("expected empty comparison, got %+v", empty)
	}
}

func TestFormatDeltaVsAvg(t *testing.T) {
	cases := []struct {
		amount, avg float64
		expected    string
	}{
		{400, 200, aboveAvgColor + "▲  100%" + Reset},
		{400, 500, belowAvgColor + "▼   20%" + Reset},
		{0, 200, belowAvgColor + "▼  100%" + Reset},
		{1020, 1000, mutedColor + "    +2%" + Reset}, // below the noise threshold -> no arrow
		{5000, 100, aboveAvgColor + "▲ >999%" + Reset},
		{100, 0, "      –"},
	}
	for _, c := range cases {
		if got := formatDeltaVsAvg(c.amount, c.avg); got != c.expected {
			t.Errorf("formatDeltaVsAvg(%.0f, %.0f) = %q; expected %q", c.amount, c.avg, got, c.expected)
		}
	}
}

func TestRenderMonthComparison(t *testing.T) {
	out := renderMonthComparison(buildMonthComparison(categoryBreakdownTestData(), "january", "2026"))

	for _, part := range []string{"Avg/month", "vs avg", "Car", "€    200.00", "▲  100%", "▼   20%", "Bills", "Total", "averages are over the 2 months"} {
		if !strings.Contains(strings.ToLower(out), strings.ToLower(part)) {
			t.Errorf("expected output to contain %q, got:\n%s", part, out)
		}
	}
	// the pie only has slices for categories spent this month
	if counts := countPieColors(generateCategoryPieChart(buildMonthCategoryBreakdownForTest(), categoryPieWidth, categoryPieHeight)); len(counts) != 3 {
		t.Errorf("expected 3 pie slices for january, got %v", counts)
	}
	// bills was not spent in january: muted row, empty bar
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Bills") && !strings.Contains(line, mutedColor+"■ Bills") {
			t.Errorf("expected muted bills row, got %q", line)
		}
	}

	if got := renderMonthComparison(MonthComparison{}); got != "No expenses for this period" {
		t.Errorf("unexpected empty output %q", got)
	}
}

func buildMonthCategoryBreakdownForTest() CategoryBreakdown {
	return buildMonthComparison(categoryBreakdownTestData(), "january", "2026").month
}
