package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TODO: maybe make this a common function for parsing arguments for all other funcs
func addExpense(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: add <amount> <category> <note>")
	}
	// TODO: add validation to where someone cannot add stuff like add 11.00 | food | meat - this results int - $11.00 | | | food | (too many |s)
	// TODO: <command> <amount> <category> <note>
	//         add       5.43      food     shopping for this and that
	// expecially for note have to figure out how to show that this part all goes into a single note

	amount, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	category := args[1]
	note := strings.Join(args[2:], " ")

	return handleExpenseAdd(amount, category, note)
}

// TODO: add predefined categories of common expenses and than anything uncommon will just be accepted as an entry
func handleExpenseAdd(amount float64, category, note string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}

	year := strconv.Itoa(time.Now().Year())
	month := time.Now().Month().String()

	//ensure nested structure exists
	if _, ok := expenses[year]; !ok {
		expenses[year] = make(map[string][]ExpenseDetails)
	}

	if _, ok := expenses[year][month]; !ok {
		expenses[year][month] = []ExpenseDetails{}
	}

	newExpense := ExpenseDetails{
		Amount:   amount,
		Category: category,
		Note:     note,
	}

	expenses[year][month] = append(expenses[year][month], newExpense)
	if saveExpenseErr := saveExpenses(expenses); saveExpenseErr != nil {
		return fmt.Errorf("Error saving expense: %s", saveExpenseErr)
	}

	fmt.Printf("\nadded $%.2f | %s | %s\n", amount, category, note)
	showSummaryCurrentMonth()

	return nil
}

// TODO: add an income section - add some predefined sections so that they cannot be mistaken
//   - paycheck
//   - transfers
//   - apartment rental
//   - dividends
//   - business trip - ? maybe just compbine with paycheck as with on-call
//   - capital gains
// TODO: add an investment section
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
