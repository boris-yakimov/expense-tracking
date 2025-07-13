package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(args []string) error
}

func main() {
	supportedCommands := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the expense-tracking tool",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Display a help message",
			callback:    commandHelp,
		},
		"list": {
			name:        "list",
			description: "List expenses",
			callback:    listExpenses,
		},
		"show-total": {
			name:        "show-total",
			description: "Show total expenses",
			callback:    showTotal,
		},
		"add": {
			name:        "add",
			description: "Add an expense",
			callback:    addExpense,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("expense-tracking > ")

		scanner.Scan()
		input := scanner.Text()
		sanitizedInput := cleanInput(input)

		command, validCommand := supportedCommands[sanitizedInput[0]]
		args := sanitizedInput[1:]

		if validCommand {
			if err := command.callback(args); err != nil {
				fmt.Printf("%s error: %s\n", command.name, err)
			}
		} else {
			fmt.Println("Unkown command, please run the :help command to see valid options")
		}
	}
}

func cleanInput(text string) []string {
	var sanitizedText = strings.Trim(strings.ToLower(text), " ")
	return strings.Split(sanitizedText, " ")
}

func commandExit(args []string) error {
	fmt.Println("Closing the expense-tracking tool... goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(args []string) error {
	fmt.Println("len of args is: ", len(args))
	fmt.Println("current args are: ", args)
	if len(args) > 0 {

		// help add
		if args[0] == "add" {
			fmt.Printf(`
Expense Tracking Tool
Usage: add <amount> <category> <note>

example 
        - add 12 food "meat from store"
        output :
        added $12.00 | food | "meat from store"
    or
        - add 25 food vegetables from store
        output :
        added $25.00 | food | vegetables from store
`)
			return nil
		}
	}

	// default help menu
	fmt.Printf(`
Expense Tracking Tool
Usage:

list:       List expenses
show-total: Show total expenses
add:        Add an expense - add <amount> <category><note>
del/delete:     Delete an expense - TO BE IMPLEMENTED
help:       Display a help message
exit:       Exit the expense-tracking tool

Detailed usage:
help add:   Format of the add command 
help del/delete : TO BE IMPLEMENTED
`)

	return nil
}

func listExpenses(args []string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}
	for i, e := range expenses {
		fmt.Printf("%d. $%-8.2f | %-6s | %-25s\n", i+1, e.Amount, e.Category, e.Note)
	}

	return nil
}

// TODO: maybe make this a common function for parsing arguments for all other funcs
func addExpense(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: add <amount> <category> <note>")
	}
	// TODO: add validation to where someone cannot add stuff like add 11.00 | food | meat - this results int - $11.00 | | | food | (too many |s)

	amount, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	category := args[1]
	note := strings.Join(args[2:], " ")

	return handleExpenseAdd(amount, category, note)
}

// TODO: show summary after something is added
func handleExpenseAdd(amount float64, category, note string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}
	expenses = append(expenses, Expense{Amount: amount, Category: category, Note: note})
	if saveExpenseErr := saveExpenses(expenses); saveExpenseErr != nil {
		return fmt.Errorf("Error saving expense: %s", saveExpenseErr)
	}

	fmt.Printf("added $%.2f | %s | %s\n", amount, category, note)
	showSummaryCurrentMonth()

	return nil
}

func showSummaryCurrentMonth() error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}

	// TODO: fit this to match the length of the table
	// TODO: make it have a table look - https://gosamples.dev/string-padding/
	fmt.Printf("+%s+\n", strings.Repeat("-", 50))

	month := time.Now().Month()
	year := time.Now().Year()
	fmt.Printf("summary for %v %v\n", month, year)

	// TODO: sort the output by category so that we see a list of food, than a list of apartment, than a list of something else, etc
	for _, e := range expenses {
		fmt.Printf("%s | $%-8.2f | %-15s\n", e.Category, e.Amount, e.Note)
	}

	// TODO: fit this to match the length of the table
	// TODO: make it have a table look - https://gosamples.dev/string-padding/
	fmt.Printf("+%s+\n", strings.Repeat("-", 50))

	return nil
}

// TODO: filter by category
func showTotal(args []string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}
	var total float64
	for _, e := range expenses {
		total += e.Amount
	}
	showSummaryCurrentMonth()
	fmt.Printf("Total expenses: $%.2f\n", total)
	return nil
}
