package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// generates an ASCII pie chart for the year's PnL
func generateAsciiPieChart(pnl PnLResult) string {
	total := pnl.incomeTotal + pnl.expenseTotal + pnl.investmentTotal
	if total == 0 {
		return "No data to display"
	}

	incomePct := pnl.incomeTotal / total
	expensePct := pnl.expenseTotal / total
	investmentPct := pnl.investmentTotal / total

	// Simple pie chart using characters
	width := 21
	height := 11

	chart := make([][]rune, height)
	for i := range chart {
		chart[i] = make([]rune, width)
		for j := range chart[i] {
			chart[i][j] = ' '
		}
	}

	centerX := width / 2
	centerY := height / 2

	// Draw filled sectors
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= float64(centerX) {
				// Inside circle
				angle := math.Atan2(dy, dx)
				if angle < 0 {
					angle += 2 * math.Pi
				}

				// Normalize angle to 0-1
				anglePct := angle / (2 * math.Pi)

				// Determine sector
				if anglePct <= incomePct {
					chart[y][x] = '█' // Income
				} else if anglePct <= incomePct+expensePct {
					chart[y][x] = '░' // Expense
				} else {
					chart[y][x] = '▒' // Investment
				}
			}
		}
	}

	// Convert to string
	var sb strings.Builder
	for _, row := range chart {
		sb.WriteString(string(row) + "\n")
	}

	// Add legend
	sb.WriteString(fmt.Sprintf("\n█ Income: €%.2f (%.1f%%)\n", pnl.incomeTotal, incomePct*100))
	sb.WriteString(fmt.Sprintf("░ Expenses: €%.2f (%.1f%%)\n", pnl.expenseTotal, expensePct*100))
	sb.WriteString(fmt.Sprintf("▒ Investments: €%.2f (%.1f%%)\n", pnl.investmentTotal, investmentPct*100))

	return sb.String()
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

	pieChart := generateAsciiPieChart(yearPnL)
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
