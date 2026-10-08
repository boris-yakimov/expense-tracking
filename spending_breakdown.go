package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// width (in characters) of the bar that visualizes each category's share of total expenses
const categoryBarWidth = 20

// pie chart size - keeping aspect ratio 2:1 for circular appearance (terminal chars are ~2x taller than wide)
const (
	categoryPieWidth  = 30
	categoryPieHeight = 15
)

// distinct colors (tview hex color tags) for the largest categories, in order of spend
// anything beyond the palette gets grouped into "other" so the pie stays readable
var categoryPalette = []string{
	"[#e06c75]", // red
	"[#61afef]", // blue
	"[#98c379]", // green
	"[#e5c07b]", // yellow
	"[#c678dd]", // purple
	"[#56b6c2]", // cyan
	"[#d19a66]", // orange
	"[#ff79c6]", // pink
	"[#dcdfe4]", // white
}

const categoryOtherColor = "[#5c6370]" // gray

// color for the category at a given position in the (spend-sorted) breakdown
func categoryColor(index int) string {
	if index < len(categoryPalette) {
		return categoryPalette[index]
	}
	return categoryOtherColor
}

// one colored slice of a pie chart
type pieSlice struct {
	color string // tview color tag
	value float64
}

// draws a colored ASCII pie chart where each slice is one expense category
// slices start at 12 o'clock and go clockwise in order of spend (largest first)
func generateCategoryPieChart(breakdown CategoryBreakdown, width, height int) string {
	var slices []pieSlice
	for i, ct := range breakdown.categories {
		if ct.total <= 0 {
			continue
		}
		color := categoryColor(i)
		// merge everything past the palette into a single "other" slice
		if color == categoryOtherColor && len(slices) > 0 && slices[len(slices)-1].color == categoryOtherColor {
			slices[len(slices)-1].value += ct.total
		} else {
			slices = append(slices, pieSlice{color: color, value: ct.total})
		}
	}
	return drawPieChart(slices, width, height)
}

// draws a colored ASCII pie chart, slices start at 12 o'clock and go clockwise in the given order
// non-positive slices are skipped; returns "" when there is nothing to draw
func drawPieChart(input []pieSlice, width, height int) string {
	var slices []pieSlice
	var total float64
	for _, s := range input {
		if s.value > 0 {
			slices = append(slices, s)
			total += s.value
		}
	}

	if total == 0 {
		return ""
	}

	// ensure minimum size
	if width < 20 {
		width = 20
	}
	if height < 10 {
		height = 10
	}

	// cumulative end-boundaries of each slice as a fraction of the full circle
	boundaries := make([]float64, len(slices))
	var cumulative float64
	for i, s := range slices {
		cumulative += s.value / total
		boundaries[i] = cumulative
	}

	centerX := width / 2
	centerY := height / 2
	radius := float64(centerX)

	var sb strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := float64(x - centerX)
			dy := float64(y-centerY) * 2 // scale dy to account for character aspect ratio
			if math.Sqrt(dx*dx+dy*dy) > radius {
				sb.WriteRune(' ')
				continue
			}

			// clockwise angle from 12 o'clock, normalized to [0, 1)
			angle := math.Atan2(dx, -dy)
			if angle < 0 {
				angle += 2 * math.Pi
			}
			anglePct := angle / (2 * math.Pi)

			color := slices[len(slices)-1].color // fallback for float rounding at the very end
			for i, b := range boundaries {
				if anglePct < b {
					color = slices[i].color
					break
				}
			}
			sb.WriteString(color + "█" + Reset)
		}
		sb.WriteRune('\n')
	}

	return sb.String()
}

// renders a category breakdown as text lines: category, amount, % of total and a bar
// the bar color matches the category's pie chart slice so the list acts as the legend
func formatCategoryBreakdown(breakdown CategoryBreakdown) string {
	if len(breakdown.categories) == 0 {
		return "No expenses for this period"
	}

	var sb strings.Builder
	for i, ct := range breakdown.categories {
		filled := int(math.Round(ct.percent / 100 * categoryBarWidth))
		if filled > categoryBarWidth {
			filled = categoryBarWidth
		}
		if filled < 0 { // negative amounts (e.g. refunds entered as expenses) should not break the bar
			filled = 0
		}
		color := categoryColor(i)
		bar := color + strings.Repeat("█", filled) + Reset + strings.Repeat("░", categoryBarWidth-filled)

		sb.WriteString(fmt.Sprintf("%s■%s %-15s €%10.2f  %5.1f%%  %s\n",
			color, Reset, tview.Escape(capitalize(ct.category)), ct.total, ct.percent, bar))
	}

	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("  %-15s €%10.2f  %5.1f%%\n", "Total", breakdown.expenseTotal, 100.0))

	return sb.String()
}

// pie chart followed by the per-category list (legend), used by both the spending breakdown and year overview pages
func renderSpendingBreakdown(breakdown CategoryBreakdown) string {
	pie := generateCategoryPieChart(breakdown, categoryPieWidth, categoryPieHeight)
	if pie == "" {
		return formatCategoryBreakdown(breakdown)
	}
	return pie + "\n" + formatCategoryBreakdown(breakdown)
}

// colors for spending relative to the monthly average: above average is bad (red), below is good (green)
const (
	aboveAvgColor = "[#e06c75]"
	belowAvgColor = "[#98c379]"
	mutedColor    = "[#5c6370]"
)

// changes smaller than this (in %) are shown without an arrow so the table doesn't flag noise
const deltaNoiseThreshold = 5.0

// month spend relative to the monthly average as a fixed-width (7 cell) colored string, e.g. "▲  14%"
func formatDeltaVsAvg(amount, avg float64) string {
	if avg <= 0 {
		return fmt.Sprintf("%7s", "–")
	}
	delta := (amount - avg) / avg * 100
	switch {
	case math.Abs(delta) < deltaNoiseThreshold:
		return fmt.Sprintf("%s%7s%s", mutedColor, fmt.Sprintf("%+.0f%%", delta), Reset)
	case delta > 999:
		return fmt.Sprintf("%s▲%6s%s", aboveAvgColor, ">999%", Reset)
	case delta > 0:
		return fmt.Sprintf("%s▲%5.0f%%%s", aboveAvgColor, delta, Reset)
	default:
		return fmt.Sprintf("%s▼%5.0f%%%s", belowAvgColor, -delta, Reset)
	}
}

// width of the rule above the total row of the month comparison table
const monthComparisonWidth = 82

// table of the month's categories with share, bar, the year's monthly average and the change vs that average
// categories spent on in other months but not this one are listed (muted) at the bottom so nothing "disappears"
func formatMonthComparison(cmp MonthComparison) []string {
	if len(cmp.rows) == 0 {
		return []string{"No expenses for this period"}
	}

	lines := []string{fmt.Sprintf("%s  %-15s %11s  %6s  %-*s  %10s  %7s%s",
		headerStyle, "Category", "This month", "Share", categoryBarWidth, "", "Avg/month", "vs avg", headerReset)}

	for i, row := range cmp.rows {
		name := tview.Escape(capitalize(row.category))
		delta := formatDeltaVsAvg(row.total, row.monthlyAvg)

		if row.total == 0 && i >= len(cmp.month.categories) {
			// nothing spent this month: no pie slice, so no legend color
			lines = append(lines, fmt.Sprintf("%s■ %-15s €%10.2f  %5.1f%%  %s  €%9.2f%s  %s",
				mutedColor, name, 0.0, 0.0, strings.Repeat("░", categoryBarWidth), row.monthlyAvg, Reset, delta))
			continue
		}

		filled := int(math.Round(row.percent / 100 * categoryBarWidth))
		filled = max(0, min(filled, categoryBarWidth)) // negative amounts (e.g. refunds) should not break the bar
		color := categoryColor(i)
		bar := color + strings.Repeat("█", filled) + Reset + strings.Repeat("░", categoryBarWidth-filled)

		lines = append(lines, fmt.Sprintf("%s■%s %-15s €%10.2f  %5.1f%%  %s  €%9.2f  %s",
			color, Reset, name, row.total, row.percent, bar, row.monthlyAvg, delta))
	}

	monthsLabel := "months"
	if cmp.monthsInYear == 1 {
		monthsLabel = "month"
	}
	lines = append(lines,
		strings.Repeat("─", monthComparisonWidth),
		fmt.Sprintf("  %-15s €%10.2f  %5.1f%%  %*s  €%9.2f  %s",
			"Total", cmp.month.expenseTotal, 100.0, categoryBarWidth, "", cmp.avgExpenseTotal,
			formatDeltaVsAvg(cmp.month.expenseTotal, cmp.avgExpenseTotal)),
		"",
		fmt.Sprintf("%sAverages are over the %d %s of the year with transactions%s", mutedColor, cmp.monthsInYear, monthsLabel, Reset),
	)
	return lines
}

// pie chart on the left, comparison table on the right (vertically centered on the pie when it is shorter)
func renderMonthComparison(cmp MonthComparison) string {
	table := formatMonthComparison(cmp)
	pie := generateCategoryPieChart(cmp.month, categoryPieWidth, categoryPieHeight)
	if pie == "" {
		return strings.Join(table, "\n")
	}

	pieLines := strings.Split(strings.TrimSuffix(pie, "\n"), "\n")
	offset := max(0, (len(pieLines)-len(table))/2)
	blank := strings.Repeat(" ", categoryPieWidth)

	var sb strings.Builder
	for i := 0; i < max(len(pieLines), len(table)+offset); i++ {
		if i < len(pieLines) {
			sb.WriteString(pieLines[i])
		} else {
			sb.WriteString(blank)
		}
		if j := i - offset; j >= 0 && j < len(table) {
			sb.WriteString("   " + table[j])
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// shows expenses grouped by category for the selected month, compared against the monthly averages of its year
// onBack is called on ESC/q; nil means go back to the transactions of the same month
func showSpendingBreakdown(month, year string, onBack func()) error {
	cmp, err := calculateMonthComparison(month, year)
	if err != nil {
		return fmt.Errorf("unable to calculate month comparison: %w", err)
	}

	monthView := styleTextView(tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false))
	monthView.SetBorder(true).SetTitle(fmt.Sprintf("Spending Breakdown - %s %s", capitalize(month), year))
	monthView.SetText(renderMonthComparison(cmp))

	flex := styleFlex(tview.NewFlex().
		AddItem(monthView, 0, 1, true))

	frame := tview.NewFrame(flex).
		AddText(Yellow+"m"+Reset+": select month   "+
			Yellow+"y"+Reset+": year summary   "+
			Yellow+"ESC"+Reset+"/"+Yellow+"q"+Reset+": back   "+
			Green+"j/k"+Reset+" or "+Green+"↑/↓"+Reset+": scroll",
			false, tview.AlignCenter, theme.FieldTextColor)

	// re-opens this exact page (same month and back target), used when returning from pages opened from here
	reopen := func() {
		if err := showSpendingBreakdown(month, year, onBack); err != nil {
			showErrorModal(fmt.Sprintf("error showing spending breakdown:\n\n%s", err), pages)
		}
	}

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ev := exitShortcuts(event); ev == nil {
			pages.RemovePage("spendingBreakdown")
			if onBack != nil {
				onBack()
				return nil
			}
			// go back to the list of transactions for the same month, focused on expenses
			if _, err := gridVisualizeTransactions(month, year, "expense", true); err != nil {
				showErrorModal(fmt.Sprintf("error showing transactions:\n\n%s", err), flex)
			}
			return nil
		}

		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'm': // pick a different month of the same year to break down
				if err := showBreakdownMonthSelector(month, year, onBack); err != nil {
					showErrorModal(fmt.Sprintf("error showing month selector:\n\n%s", err), flex)
				}
				return nil
			case 'y': // jump to the summary of this month's year, ESC there comes back here
				if err := showYearResults(year, reopen); err != nil {
					showErrorModal(fmt.Sprintf("error showing year results:\n\n%s", err), flex)
				}
				return nil
			}
		}

		return vimMotions(event)
	})

	pages.AddPage("spendingBreakdown", frame, true, true)
	tui.SetFocus(monthView)
	return nil
}

// pop-up list of months (with transactions) in the given year; selecting one re-renders the breakdown for it
// onBack is passed through so the breakdown keeps returning to wherever it was opened from
func showBreakdownMonthSelector(currentMonth, year string, onBack func()) error {
	months, err := getMonthsForYear(year) // sorted newest first, same as the main month selector
	if err != nil {
		return fmt.Errorf("unable to get months for year %s: %w", year, err)
	}

	list := styleList(tview.NewList().ShowSecondaryText(false))
	for _, m := range months {
		monthCopy := m // capture loop variable for the closure
		list.AddItem(capitalize(m), "", 0, func() {
			pages.RemovePage("breakdownMonthSelector")
			if err := showSpendingBreakdown(monthCopy, year, onBack); err != nil {
				showErrorModal(fmt.Sprintf("error showing spending breakdown:\n\n%s", err), list)
			}
		})
		// start the selection on the month currently being shown
		if m == currentMonth {
			list.SetCurrentItem(list.GetItemCount() - 1)
		}
	}

	list.SetTitle(fmt.Sprintf("Select Month - %s", year)).
		SetTitleAlign(tview.AlignCenter).
		SetBorder(true)

	frame := tview.NewFrame(list).
		AddText(generateCombinedControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// horizontal centering
	modal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(frame, 60, 1, true).
		AddItem(nil, 0, 1, false))

	// vertical centering - list height fits all 12 months plus border and footer
	centeredModal := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(modal, len(months)+6, 1, true).
		AddItem(nil, 0, 1, false))

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ev := exitShortcuts(event); ev == nil {
			// close the selector and return to the breakdown that was open (re-rendered so focus is restored on its panel)
			pages.RemovePage("breakdownMonthSelector")
			if err := showSpendingBreakdown(currentMonth, year, onBack); err != nil {
				showErrorModal(fmt.Sprintf("error showing spending breakdown:\n\n%s", err), list)
			}
			return nil
		}
		return vimMotions(event)
	})

	pages.AddPage("breakdownMonthSelector", centeredModal, true, true)
	tui.SetFocus(list)
	return nil
}
