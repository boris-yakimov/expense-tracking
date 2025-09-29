package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// generatePieChart returns a string with colored ASCII pie chart
func generatePieChart(pnl PnLResult, width, height int) string {
	total := pnl.incomeTotal + pnl.expenseTotal + pnl.investmentTotal
	if total == 0 {
		return "No data to display"
	}

	incomePct := pnl.incomeTotal / total
	expensePct := pnl.expenseTotal / total
	investmentPct := pnl.investmentTotal / total

	// ensure minimum size
	if width < 20 {
		width = 20
	} // pie chart size - keeping aspect ratio 2:1 for circular appearance
	if height < 10 {
		height = 10
	}
	chart := make([][]rune, height)
	for i := range chart {
		chart[i] = make([]rune, width)
		for j := range chart[i] {
			chart[i][j] = ' '
		}
	}

	centerX := width / 2
	centerY := height / 2

	for y := range chart {
		for x := range chart[y] {
			dx := float64(x - centerX)
			dy := float64(y-centerY) * 2 // scale dy to account for character aspect ratio (characters are taller than wide)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= float64(centerX) {
				chart[y][x] = '█' // all filled with same character, colored later
			}
		}
	}

	// convert to string with colors
	result := ""
	for y := range chart {
		for x := range chart[y] {
			ch := chart[y][x]
			if ch == '█' {
				dx := float64(x - centerX)
				dy := float64(y-centerY) * 2  // same scaling as in drawing
				angle := math.Atan2(dy/2, dx) // use original dy for angle
				if angle < 0 {
					angle += 2 * math.Pi
				}
				anglePct := angle / (2 * math.Pi)

				if anglePct <= incomePct {
					result += Blue + string(ch) + Reset
				} else if anglePct <= incomePct+expensePct {
					result += Red + string(ch) + Reset
				} else {
					result += Green + string(ch) + Reset
				}
			} else {
				result += string(ch)
			}
		}
		result += "\n"
	}

	// add legend
	result += fmt.Sprintf("\n%s█%s Income: €%.2f (%.1f%%)\n", Blue, Reset, pnl.incomeTotal, incomePct*100)
	result += fmt.Sprintf("%s█%s Expenses: €%.2f (%.1f%%)\n", Red, Reset, pnl.expenseTotal, expensePct*100)
	result += fmt.Sprintf("%s█%s Investments: €%.2f (%.1f%%)\n", Green, Reset, pnl.investmentTotal, investmentPct*100)

	return result
}

// shows the year results window with monthly PnL and pie chart
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

	// pie chart size - keeping aspect ratio 2:1 for circular appearance
	pieWidth := 30
	pieHeight := 15

	// left panel: list of months with PnL
	leftText := styleTextView(tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false))

	leftText.SetBorder(true).SetTitle(fmt.Sprintf("Monthly Results - %s", year))

	var leftContent strings.Builder
	for _, month := range months {
		pnl := monthlyPnL[month]
		leftContent.WriteString(fmt.Sprintf("%s:\n", capitalize(month)))
		leftContent.WriteString(fmt.Sprintf("  Income: €%.2f\n", pnl.incomeTotal))
		leftContent.WriteString(fmt.Sprintf("  Expenses: €%.2f\n", pnl.expenseTotal))
		leftContent.WriteString(fmt.Sprintf("  Investments: €%.2f\n", pnl.investmentTotal))
		leftContent.WriteString(fmt.Sprintf("  P&L: €%.2f (%.1f%%)\n\n", pnl.pnlAmount, pnl.pnlPercent))
	}

	leftContent.WriteString("-----------------------------\n")
	leftContent.WriteString("Year Total:\n")
	leftContent.WriteString(fmt.Sprintf("  Income: €%.2f\n", yearPnL.incomeTotal))
	leftContent.WriteString(fmt.Sprintf("  Expenses: €%.2f\n", yearPnL.expenseTotal))
	leftContent.WriteString(fmt.Sprintf("  Investments: €%.2f\n", yearPnL.investmentTotal))
	leftContent.WriteString(fmt.Sprintf("  P&L: €%.2f (%.1f%%)\n", yearPnL.pnlAmount, yearPnL.pnlPercent))

	leftText.SetText(leftContent.String())

	// right panel: ASCII pie chart
	rightText := styleTextView(tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false))

	rightText.SetBorder(true).SetTitle("Year Overview")

	pieChart := generatePieChart(yearPnL, pieWidth, pieHeight)
	rightText.SetText(pieChart)

	// split view
	flex := styleFlex(tview.NewFlex().
		AddItem(leftText, 0, 1, false).
		AddItem(rightText, 0, 1, false))

	// frame with navigation
	frame := tview.NewFrame(flex).
		AddText(generateCombinedControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// input capture for navigation
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ev := exitShortcuts(event); ev == nil {
			// go back to year selector
			if err := showYearSelector(); err != nil {
				showErrorModal(fmt.Sprintf("error showing year selector:\n\n%s", err), nil, flex)
			}
			return nil
		}
		return event
	})

	tui.SetRoot(frame, true).SetFocus(flex)
	return nil
}
