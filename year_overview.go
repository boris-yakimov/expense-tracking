package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// colors (tview hex color tags) for the P&L flows, matching the palette used by the spending breakdown
const (
	incomeColor     = "[#61afef]" // blue
	expenseColor    = "[#e06c75]" // red
	investmentColor = "[#c678dd]" // purple
	positiveColor   = "[#98c379]" // green - savings >= 0
	negativeColor   = "[#e06c75]" // red - savings < 0
	headerStyle     = "[#56b6c2::b]"
	headerReset     = "[-::-]"
)

// width of the rule separating rows from totals in the year page panels
const yearTableWidth = 67

// green for non-negative savings, red for a loss
func savingsColor(amount float64) string {
	if amount < 0 {
		return negativeColor
	}
	return positiveColor
}

// savings rate as text, or n/a when there was no income to compare against
func formatSavingsRate(pnl PnLResult) string {
	if pnl.incomeTotal == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", pnl.pnlPercent)
}

// pie chart of income vs expenses vs investments (share of all money flows in the period)
func generatePnLPieChart(pnl PnLResult, width, height int) string {
	return drawPieChart([]pieSlice{
		{color: incomeColor, value: pnl.incomeTotal},
		{color: expenseColor, value: pnl.expenseTotal},
		{color: investmentColor, value: pnl.investmentTotal},
	}, width, height)
}

// legend for the P&L pie: each flow with its amount and share of the pie, then savings vs income
func formatPnLLegend(pnl PnLResult) []string {
	total := pnl.incomeTotal + pnl.expenseTotal + pnl.investmentTotal
	share := func(v float64) float64 {
		if total <= 0 || v <= 0 {
			return 0
		}
		return v / total * 100
	}

	return []string{
		fmt.Sprintf("%s■%s %-12s €%10.2f  %5.1f%%", incomeColor, Reset, "Income", pnl.incomeTotal, share(pnl.incomeTotal)),
		fmt.Sprintf("%s■%s %-12s €%10.2f  %5.1f%%", expenseColor, Reset, "Expenses", pnl.expenseTotal, share(pnl.expenseTotal)),
		fmt.Sprintf("%s■%s %-12s €%10.2f  %5.1f%%", investmentColor, Reset, "Investments", pnl.investmentTotal, share(pnl.investmentTotal)),
		strings.Repeat("─", 34),
		fmt.Sprintf("  %-12s %s€%10.2f%s", "Savings", savingsColor(pnl.pnlAmount), pnl.pnlAmount, Reset),
		fmt.Sprintf("  %-12s %s%11s%s  of income", "Savings rate", savingsColor(pnl.pnlAmount), formatSavingsRate(pnl), Reset),
	}
}

// renders the P&L pie with its legend vertically centered to the right of it
func renderYearPnL(pnl PnLResult) string {
	legend := formatPnLLegend(pnl)
	pie := generatePnLPieChart(pnl, categoryPieWidth, categoryPieHeight)
	if pie == "" {
		return "No transactions for this period\n\n" + strings.Join(legend, "\n")
	}

	pieLines := strings.Split(strings.TrimSuffix(pie, "\n"), "\n")
	offset := (len(pieLines) - len(legend)) / 2
	if offset < 0 {
		offset = 0
	}

	var sb strings.Builder
	for i, line := range pieLines {
		sb.WriteString(line)
		if j := i - offset; j >= 0 && j < len(legend) {
			sb.WriteString("   " + legend[j])
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// one row of the monthly results table; amounts are colored per flow so the table reads like the pie legend
func formatPnLRow(label string, pnl PnLResult) string {
	return fmt.Sprintf("%-10s %s%11.2f%s %s%11.2f%s %s%11.2f%s %s%11.2f%s %8s",
		tview.Escape(label),
		incomeColor, pnl.incomeTotal, Reset,
		expenseColor, pnl.expenseTotal, Reset,
		investmentColor, pnl.investmentTotal, Reset,
		savingsColor(pnl.pnlAmount), pnl.pnlAmount, Reset,
		formatSavingsRate(pnl))
}

// table of P&L per month (chronological) with the year total at the bottom
func formatMonthlyPnLTable(months []string, monthlyPnL map[string]PnLResult, yearPnL PnLResult) string {
	if len(months) == 0 {
		return "No transactions for this year"
	}

	// calendar order reads more naturally in a year summary than the newest-first order of the selectors
	sorted := append([]string(nil), months...)
	sort.Slice(sorted, func(i, j int) bool { return monthOrder[sorted[i]] < monthOrder[sorted[j]] })

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s%-10s %11s %11s %11s %11s %8s%s\n",
		headerStyle, "Month", "Income €", "Expenses €", "Invested €", "Savings €", "Rate", headerReset))
	for _, month := range sorted {
		sb.WriteString(formatPnLRow(capitalize(month), monthlyPnL[month]) + "\n")
	}
	sb.WriteString(strings.Repeat("─", yearTableWidth) + "\n")
	sb.WriteString(formatPnLRow("Total", yearPnL) + "\n")

	return sb.String()
}

// shows the year summary: P&L pie, monthly results table and the full-year spending breakdown
func showYearResults(year string) error {
	monthlyPnL, err := calculateYearMonthlyPnL(year)
	if err != nil {
		return fmt.Errorf("unable to calculate monthly pnl: %w", err)
	}

	yearPnL, err := calculateYearPnL(year)
	if err != nil {
		return fmt.Errorf("unable to calculate year pnl: %w", err)
	}

	months, err := getMonthsForYear(year)
	if err != nil {
		return fmt.Errorf("unable to get months for year: %w", err)
	}

	yearCategories, err := calculateYearCategoryBreakdown(year)
	if err != nil {
		return fmt.Errorf("unable to calculate year category breakdown: %w", err)
	}

	newPanel := func(title, text string) *tview.TextView {
		view := styleTextView(tview.NewTextView().
			SetDynamicColors(true).
			SetWordWrap(false))
		view.SetBorder(true).SetTitle(title)
		view.SetText(text)
		return view
	}

	pnlView := newPanel(fmt.Sprintf("Profit & Loss - %s", year), renderYearPnL(yearPnL))
	monthlyView := newPanel(fmt.Sprintf("Monthly Results - %s", year), formatMonthlyPnLTable(months, monthlyPnL, yearPnL))
	categoryView := newPanel(fmt.Sprintf("Spending Breakdown - %s (Full Year)", year), renderSpendingBreakdown(yearCategories))

	// left column: P&L summary on top (sized to fit the pie + borders), monthly table below
	leftColumn := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pnlView, categoryPieHeight+2, 0, true).
		AddItem(monthlyView, 0, 1, false))

	flex := styleFlex(tview.NewFlex().
		AddItem(leftColumn, 0, 1, true).
		AddItem(categoryView, 0, 1, false))

	frame := tview.NewFrame(flex).
		AddText(generateCombinedControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// TAB cycles focus between panels so long content can be scrolled with j/k
	views := []*tview.TextView{pnlView, monthlyView, categoryView}
	current := 0

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ev := exitShortcuts(event); ev == nil {
			// go back to year selector
			pages.SwitchToPage("yearSelector")
			return nil
		}

		event = vimMotions(event)
		switch event.Key() {
		case tcell.KeyTAB, tcell.KeyRight:
			current = (current + 1) % len(views)
			tui.SetFocus(views[current])
			return nil
		case tcell.KeyBacktab, tcell.KeyLeft:
			current = (current - 1 + len(views)) % len(views)
			tui.SetFocus(views[current])
			return nil
		}
		return event
	})

	pages.AddPage("yearResults", frame, true, true)
	tui.SetFocus(pnlView)
	return nil
}
