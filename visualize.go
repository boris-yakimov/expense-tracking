package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
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
		for month, expenseList := range months {
			for i, e := range expenseList {
				fmt.Printf("year: %s\n", year)
				fmt.Printf("month: %s\n", month)
				fmt.Printf("%d. $%-8.2f | %-6s | %-25s\n", i+1, e.Amount, e.Category, e.Note)
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

	// TODO: fit this to match the length of the table
	// TODO: make it have a table look - https://gosamples.dev/string-padding/
	fmt.Printf("\n+%s+\n", strings.Repeat("-", 50))

	year := strconv.Itoa(time.Now().Year())
	month := time.Now().Month().String()
	fmt.Printf("summary for %v %v\n", month, year)

	// sort expenses by category in alphabetical order
	sort.Slice(expenses[year][month], func(i, j int) bool {
		return expenses[year][month][i].Category < expenses[year][month][j].Category
	})

	for _, e := range expenses[year][month] {
		fmt.Printf("%s | $%-8.2f | %-15s\n", e.Category, e.Amount, e.Note)
	}

	// TODO: fit this to match the length of the table
	// TODO: make it have a table look - https://gosamples.dev/string-padding/
	fmt.Printf("+%s+\n", strings.Repeat("-", 50))

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
