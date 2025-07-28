package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	noteMaxLength = 42
)

func addExpense(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: add <amount> <category> <note>")
	}

	amount, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	category := args[1]
	if _, ok := allowedExpenseCategories[category]; !ok {
		fmt.Printf("\ninvalid expense category: \"%s\"", category)
		showAllowedCategories("expense") // expense, income, investment
		return fmt.Errorf("\n\nPlease pick a valid expense category from the list above.")
	}

	note := strings.Join(args[2:], " ")
	if len(note) > noteMaxLength {
		return fmt.Errorf("\nnote should be a maximum of %v characters, provided %v", noteMaxLength, len(note))
	}

	if !validNoteInputFormat(note) {
		return fmt.Errorf("\ninvalid character in note, notes should contain only letters, numbers, spaces, commas, or dashes")
	}

	return handleExpenseAdd(amount, category, note)
}

func handleExpenseAdd(amount float64, category, note string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}

	year := strconv.Itoa(time.Now().Year())
	month := time.Now().Month().String()

	transactionType := "Expenses"

	// ensure nested structure exists
	if _, ok := expenses[year]; !ok {
		expenses[year] = make(map[string]map[string][]Transaction)
	}

	if _, ok := expenses[year][month]; !ok {
		expenses[year][month] = make(map[string][]Transaction)
	}

	if _, ok := expenses[year][month][transactionType]; !ok {
		expenses[year][month][transactionType] = []Transaction{}
	}

	newExpense := Transaction{
		Amount:   amount,
		Category: category,
		Note:     note,
	}

	expenses[year][month][transactionType] = append(expenses[year][month][transactionType], newExpense)
	if saveExpenseErr := saveExpenses(expenses); saveExpenseErr != nil {
		return fmt.Errorf("Error saving expense: %s", saveExpenseErr)
	}

	fmt.Printf("\nadded $%.2f | %s | %s\n", amount, category, note)

	// TODO: figure out a better way to define cli callback funcs to avoid just passing aroungs args even in places where they are not mandatory
	var args []string
	showTotal(args)

	return nil
}
