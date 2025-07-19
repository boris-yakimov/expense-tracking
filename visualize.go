package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TODO: fix paddding
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

func showSummaryCurrentMonth() error {
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

	fmt.Printf("\nSummary for %v %v", month, year)
	fmt.Printf("\n+%s+\n", strings.Repeat("-", 58))

	// sort expenses by category in alphabetical order
	sort.Slice(monthExpenses, func(i, j int) bool {
		return monthExpenses[i].Category < monthExpenses[j].Category
	})

	for i, e := range monthExpenses {
		fmt.Printf("| %2d. $%-8.2f | %-10s | %-27s |\n", i+1, e.Amount, e.Category, e.Note)
	}

	fmt.Printf("+%s+\n", strings.Repeat("-", 58))

	return nil
}

// TODO: filter by category
// TODO: filter by year or month
func showTotal(args []string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}

	year := strconv.Itoa(time.Now().Year())
	month := time.Now().Month().String()

	var total float64
	for _, e := range expenses[year][month] {
		total += e.Amount
	}
	showSummaryCurrentMonth()
	fmt.Printf("Total expenses: $%.2f\n", total)
	return nil
}
