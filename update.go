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

// TODO: maybe make this a common function for parsing arguments for all other funcs
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
		// TODO: create a function to print a list of allowed expense categories for the user
		return fmt.Errorf("\nInvalid expense category: %s", category)
	}

	// TODO: note show contain only alphanumeric and comma, dash
	note := strings.Join(args[2:], " ")
	if len(note) > noteMaxLength {
		return fmt.Errorf("\nNote should be a maximum of %v characters, provided %v", noteMaxLength, len(note))
	}

	return handleExpenseAdd(amount, category, note)
}

// TODO: only allow add if category is allowed
func handleExpenseAdd(amount float64, category, note string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}

	year := strconv.Itoa(time.Now().Year())
	month := time.Now().Month().String()

	//ensure nested structure exists
	if _, ok := expenses[year]; !ok {
		expenses[year] = make(map[string][]Expense)
	}

	if _, ok := expenses[year][month]; !ok {
		expenses[year][month] = []Expense{}
	}

	newExpense := Expense{
		Amount:   amount,
		Category: category,
		Note:     note,
	}

	expenses[year][month] = append(expenses[year][month], newExpense)
	if saveExpenseErr := saveExpenses(expenses); saveExpenseErr != nil {
		return fmt.Errorf("Error saving expense: %s", saveExpenseErr)
	}

	fmt.Printf("\nadded $%.2f | %s | %s\n", amount, category, note)

	// TODO: figure out a better way to define cli callback funcs to avoid just passing aroungs args even in places where they are not mandatory
	var args []string
	showTotal(args)

	return nil
}

// TODO: add an income section - add some predefined sections so that they cannot be mistaken
//   - paycheck
//   - transfers
//   - apartment rental
//   - dividends
//   - business trip - ? maybe just compbine with paycheck as with on-call
//   - capital gains
//
// TODO: add an investment section
//
// TODO: should be
//           year {
//             month {
//                investments {
//                  amount: 123
//                  category: ibkr
//                  note: annual investmentu
//                }
//                income {
//                  amount: 123
//                  category: salary
//                  note: work
//                }
//                expenses {
//                  amount: 123
//                  category: food
//                  note: groceries
//                }
//             }
//          }
