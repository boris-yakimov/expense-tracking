package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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

		// check if command is supported
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
	fmt.Println("Closing the expnse-tracking tool... goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(args []string) error {
	fmt.Printf(`
Expense Tracking Tool
Usage:

list:       List expenses
show-total: Show total expenses
add:        Add an expense
help:       Display a help message
exit:       Exit the expense-tracking tool
`)
	return nil
}

func listExpenses(args []string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}
	for i, e := range expenses {
		fmt.Printf("%d. $%.2f | %s | %s\n", i+1, e.Amount, e.Category, e.Note)
	}

	return nil
}

func addExpense(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: add <amount> <category> <note>")
	}

	amount, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	category := args[1]
	note := strings.Join(args[2:], " ")

	return handleExpenseAdd(amount, category, note)
}

func handleExpenseAdd(amount float64, category, note string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}
	expenses = append(expenses, Expense{Amount: amount, Category: category, Note: note})
	return nil
}

func showTotal(args []string) error {
	expenses, loadFileErr := loadExpenses()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load expenses file: %s", loadFileErr)
	}
	var total float64
	for _, e := range expenses {
		total += e.Amount
	}
	fmt.Printf("Total expenses: $%.2f\n")
	return nil
}
