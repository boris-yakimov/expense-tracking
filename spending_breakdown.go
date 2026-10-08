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

// shows expenses grouped by category for the selected month and for the whole year of that month
func showSpendingBreakdown(month, year string) error {
	monthBreakdown, err := calculateMonthCategoryBreakdown(month, year)
	if err != nil {
		return fmt.Errorf("unable to calculate month category breakdown: %w", err)
	}

	yearBreakdown, err := calculateYearCategoryBreakdown(year)
	if err != nil {
		return fmt.Errorf("unable to calculate year category breakdown: %w", err)
	}

	monthView := styleTextView(tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false))
	monthView.SetBorder(true).SetTitle(fmt.Sprintf("Spending Breakdown - %s %s", capitalize(month), year))
	monthView.SetText(renderSpendingBreakdown(monthBreakdown))

	yearView := styleTextView(tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false))
	yearView.SetBorder(true).SetTitle(fmt.Sprintf("Spending Breakdown - %s (Full Year)", year))
	yearView.SetText(renderSpendingBreakdown(yearBreakdown))

	flex := styleFlex(tview.NewFlex().
		AddItem(monthView, 0, 1, true).
		AddItem(yearView, 0, 1, false))

	frame := tview.NewFrame(flex).
		AddText(Yellow+"m"+Reset+": select month   "+generateCombinedControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// keep track of which panel is focused so TAB can switch between them (useful for scrolling long lists)
	views := []*tview.TextView{monthView, yearView}
	current := 0

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ev := exitShortcuts(event); ev == nil {
			pages.RemovePage("spendingBreakdown")
			// go back to the list of transactions for the same month, focused on expenses
			if _, err := gridVisualizeTransactions(month, year, "expense", true); err != nil {
				showErrorModal(fmt.Sprintf("error showing transactions:\n\n%s", err), flex)
			}
			return nil
		}

		// pick a different month of the same year to break down
		if event.Key() == tcell.KeyRune && event.Rune() == 'm' {
			if err := showBreakdownMonthSelector(month, year); err != nil {
				showErrorModal(fmt.Sprintf("error showing month selector:\n\n%s", err), flex)
			}
			return nil
		}

		event = vimMotions(event)
		switch event.Key() {
		case tcell.KeyTAB, tcell.KeyBacktab, tcell.KeyLeft, tcell.KeyRight:
			current = (current + 1) % len(views)
			tui.SetFocus(views[current])
			return nil
		}
		return event
	})

	pages.AddPage("spendingBreakdown", frame, true, true)
	tui.SetFocus(monthView)
	return nil
}

// pop-up list of months (with transactions) in the given year; selecting one re-renders the breakdown for it
// the full-year panel stays the same since the year doesn't change
func showBreakdownMonthSelector(currentMonth, year string) error {
	months, err := getMonthsForYear(year) // sorted newest first, same as the main month selector
	if err != nil {
		return fmt.Errorf("unable to get months for year %s: %w", year, err)
	}

	list := styleList(tview.NewList().ShowSecondaryText(false))
	for _, m := range months {
		monthCopy := m // capture loop variable for the closure
		list.AddItem(capitalize(m), "", 0, func() {
			pages.RemovePage("breakdownMonthSelector")
			if err := showSpendingBreakdown(monthCopy, year); err != nil {
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
			if err := showSpendingBreakdown(currentMonth, year); err != nil {
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
