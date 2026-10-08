package main

import (
	"fmt"
	"math"
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

// width (in characters) of the bars in the P&L panel
const pnlBarWidth = 40

// a flow as a share of income, or n/a when there was no income to compare against
func formatPercentOfIncome(amount, income float64) string {
	if income == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", amount/income*100)
}

// P&L summary: every flow is shown relative to income (not as a share of all flows combined, which says nothing)
// bars share one scale so they can be compared directly; the income bar is the reference
// "after investing" is what's left once both expenses and investments are paid - negative means
// more was invested than saved this period, i.e. it came out of existing savings
func renderYearPnL(pnl PnLResult) string {
	if pnl.incomeTotal == 0 && pnl.expenseTotal == 0 && pnl.investmentTotal == 0 {
		return "No transactions for this period"
	}

	afterInvesting := pnl.pnlAmount - pnl.investmentTotal
	scale := max(pnl.incomeTotal, pnl.expenseTotal, pnl.investmentTotal, math.Abs(pnl.pnlAmount), math.Abs(afterInvesting))
	bar := func(color string, v float64) string {
		filled := 0
		if scale > 0 && v > 0 {
			filled = min(pnlBarWidth, int(math.Round(v/scale*pnlBarWidth)))
		}
		return color + strings.Repeat("█", filled) + Reset + strings.Repeat("░", pnlBarWidth-filled)
	}
	row := func(color, label string, v float64, pct string) string {
		return fmt.Sprintf("%s■%s %-16s %s€%10.2f%s  %9s   %s", color, Reset, label, color, v, Reset, pct, bar(color, v))
	}

	return strings.Join([]string{
		"",
		fmt.Sprintf("%s  %-16s %11s  %9s%s", headerStyle, "", "Amount", "of income", headerReset),
		row(incomeColor, "Income", pnl.incomeTotal, ""),
		row(expenseColor, "Expenses", pnl.expenseTotal, formatPercentOfIncome(pnl.expenseTotal, pnl.incomeTotal)),
		row(investmentColor, "Investments", pnl.investmentTotal, formatPercentOfIncome(pnl.investmentTotal, pnl.incomeTotal)),
		strings.Repeat("─", 2+16+1+11+2+9+3+pnlBarWidth),
		row(savingsColor(pnl.pnlAmount), "Savings", pnl.pnlAmount, formatSavingsRate(pnl)),
		row(savingsColor(afterInvesting), "After investing", afterInvesting, formatPercentOfIncome(afterInvesting, pnl.incomeTotal)),
		"",
		mutedColor + "  Savings = income - expenses.  After investing = savings - investments." + Reset,
	}, "\n") + "\n"
}

// background of the highlighted row in the monthly results table
var selectedRowColor = tcell.NewRGBColor(62, 68, 82)

// converts a tview hex color tag like "[#61afef]" into a tcell color
func tagColor(tag string) tcell.Color {
	return tcell.GetColor(strings.Trim(tag, "[]"))
}

// table of P&L per month (chronological) with the year total at the bottom
// month rows are selectable and carry the month name as their reference so Enter can open that month
func createMonthlyPnLTable(months []string, monthlyPnL map[string]PnLResult, yearPnL PnLResult) *tview.Table {
	table := styleTable(tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0))
	table.SetBackgroundColor(theme.FieldBackgroundColor) // match the text panels next to it

	if len(months) == 0 {
		table.SetCell(0, 0, tview.NewTableCell("No transactions for this year").SetSelectable(false))
		return table
	}

	headerColor := tagColor("[#56b6c2]")
	headers := []string{"Month", "Income €", "Expenses €", "Invested €", "Savings €", "Rate"}
	for col, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(headerColor).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		if col > 0 {
			cell.SetAlign(tview.AlignRight)
		}
		table.SetCell(0, col, cell)
	}

	// fills one row; selectable rows get a highlighted background that keeps each column's color
	setRow := func(row int, label string, pnl PnLResult, selectable bool, ref string) {
		values := []struct {
			text  string
			color tcell.Color
		}{
			{label, theme.FieldTextColor},
			{fmt.Sprintf("%11.2f", pnl.incomeTotal), tagColor(incomeColor)},
			{fmt.Sprintf("%11.2f", pnl.expenseTotal), tagColor(expenseColor)},
			{fmt.Sprintf("%11.2f", pnl.investmentTotal), tagColor(investmentColor)},
			{fmt.Sprintf("%11.2f", pnl.pnlAmount), tagColor(savingsColor(pnl.pnlAmount))},
			{fmt.Sprintf("%8s", formatSavingsRate(pnl)), theme.FieldTextColor},
		}
		for col, v := range values {
			cell := tview.NewTableCell(v.text).
				SetTextColor(v.color).
				SetSelectable(selectable).
				SetReference(ref).
				SetSelectedStyle(tcell.StyleDefault.Foreground(v.color).Background(selectedRowColor).Bold(true))
			if col > 0 {
				cell.SetAlign(tview.AlignRight)
			}
			if !selectable {
				cell.SetAttributes(tcell.AttrBold)
			}
			table.SetCell(row, col, cell)
		}
	}

	// calendar order reads more naturally in a year summary than the newest-first order of the selectors
	sorted := append([]string(nil), months...)
	sort.Slice(sorted, func(i, j int) bool { return monthOrder[sorted[i]] < monthOrder[sorted[j]] })

	for i, month := range sorted {
		setRow(i+1, capitalize(month), monthlyPnL[month], true, month)
	}

	// blank spacer + bold total row, neither selectable
	spacerRow := len(sorted) + 1
	table.SetCell(spacerRow, 0, tview.NewTableCell("").SetSelectable(false))
	setRow(spacerRow+1, "Total", yearPnL, false, "")

	// start on the most recent month, it's usually the one worth drilling into
	table.Select(len(sorted), 0)
	return table
}

// shows the year summary: P&L pie, monthly results table and the full-year spending breakdown
// Enter on a month opens its spending breakdown; onBack is called on ESC/q (nil means back to the year selector)
func showYearResults(year string, onBack func()) error {
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
	categoryView := newPanel(fmt.Sprintf("Spending Breakdown - %s (Full Year)", year), renderSpendingBreakdown(yearCategories))

	monthlyTable := createMonthlyPnLTable(months, monthlyPnL, yearPnL)
	monthlyTable.SetBorder(true).SetTitle(fmt.Sprintf("Monthly Results - %s", year))

	// opening a month from here should come back to this same summary
	reopen := func() {
		if err := showYearResults(year, onBack); err != nil {
			showErrorModal(fmt.Sprintf("error showing year results:\n\n%s", err), pages)
		}
	}
	monthlyTable.SetSelectedFunc(func(row, col int) {
		month, _ := monthlyTable.GetCell(row, col).GetReference().(string)
		if month == "" {
			return
		}
		if err := showSpendingBreakdown(month, year, reopen); err != nil {
			showErrorModal(fmt.Sprintf("error showing spending breakdown:\n\n%s", err), monthlyTable)
		}
	})

	// left column: P&L summary on top (sized to fit the pie + borders), monthly table below
	leftColumn := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pnlView, strings.Count(pnlView.GetText(false), "\n")+2, 0, false).
		AddItem(monthlyTable, 0, 1, true))

	flex := styleFlex(tview.NewFlex().
		AddItem(leftColumn, 0, 1, true).
		AddItem(categoryView, 0, 1, false))

	frame := tview.NewFrame(flex).
		AddText(Yellow+"Enter"+Reset+": month breakdown   "+
			Yellow+"ESC"+Reset+"/"+Yellow+"q"+Reset+": back   "+
			Green+"TAB"+Reset+": next panel   "+
			Green+"j/k"+Reset+" or "+Green+"↑/↓"+Reset+": navigate",
			false, tview.AlignCenter, theme.FieldTextColor)

	// TAB cycles focus between panels so long content can be scrolled with j/k; starts on the table
	views := []tview.Primitive{monthlyTable, categoryView, pnlView}
	current := 0

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ev := exitShortcuts(event); ev == nil {
			pages.RemovePage("yearResults")
			if onBack != nil {
				onBack()
				return nil
			}
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
	tui.SetFocus(monthlyTable)
	return nil
}
