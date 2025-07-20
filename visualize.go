package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TODO: calculate % savings rate (should show minus percent of expenses exceed income)
// TODO: montly totals with less details - income, expense, investement, % p&l
// TODO: yearly totals with less details - income, expense, investment, %p&l
// TODO: draw a terminal diagram/pie chart

const (
	amountWidth   = 10
	categoryWidth = 12
	noteWidth     = 40
)

func listExpenses(args []string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %w", loadFileErr)
	}

	if len(expenses) == 0 {
		fmt.Println("No expenses found")
		return nil
	}

	for year, months := range expenses {
		fmt.Printf("\nYear: %s\n", year)
		for month, expenseList := range months {
			fmt.Printf("  Month: %s\n", month)
			if len(expenseList) == 0 {
				fmt.Println("    No expenses recorded.")
				continue
			}
			for i, e := range expenseList {
				fmt.Printf("    %2d. $%-8.2f | %-10s | %-25s\n", i+1, e.Amount, e.Category, e.Note)
			}
		}
	}

	return nil
}

func showTotal(args []string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}

	year := strconv.Itoa(time.Now().Year())
	month := time.Now().Month().String()

	monthExpenses, ok := expenses[year][month]
	if !ok || len(monthExpenses) == 0 {
		fmt.Printf("\nNo expenses found for %s %s.\n", month, year)
	}

	fmt.Printf("\nSummary for %v %v\n", month, year)

	//table border width: + 2 (padding per field) * 3 columns + 4 (pipes) + field widths
	border := "+" + strings.Repeat("-", amountWidth+categoryWidth+noteWidth+10) + "+"
	fmt.Println(border)

	// sort expenses by category in alphabetical order
	sort.Slice(monthExpenses, func(i, j int) bool {
		return monthExpenses[i].Category < monthExpenses[j].Category
	})

	var total float64
	for i, e := range monthExpenses {
		category := padRight(e.Category, categoryWidth)
		note := truncateOrPad(e.Note, noteWidth)
		fmt.Printf("| %2d. $%-8.2f | %-*s | %-*s |\n", i+1, e.Amount, categoryWidth, category, noteWidth, note)
		total += e.Amount
	}

	fmt.Println(border)
	fmt.Printf("Total expenses: $%.2f\n", total)

	return nil
}

// trim string to fit a preset width
func padRight(str string, width int) string {
	if len(str) > width {
		return str[:width]
	}
	return str + strings.Repeat(" ", width-len(str))
}

// ensure note fits within a preset width
func truncateOrPad(str string, width int) string {
	runes := []rune(str)
	if len(runes) > width {
		return string(runes[:width])
	}
	return str + strings.Repeat(" ", width-len(runes))
}
