package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

func main() {
	supportedCommands := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the expense-tracking cli",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Display a help message",
			callback:    commandHelp,
		},
		// "expense": {
		// 	name:        "expense",
		// 	description: "Add an expense",
		// 	callback:    commandExpense,
		// },
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("expense-tracking > ")

		scanner.Scan()
		input := scanner.Text()
		sanitizedInput := cleanInput(input)

		// check if command is supported
		command, validCommand := supportedCommands[sanitizedInput[0]]

		if validCommand {
			if err := command.callback(); err != nil {
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

func commandExit() error {
	fmt.Println("Closing the expnse-tracking cli... goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Printf(`
	func commandHelp() error {
Expense Tracking Tool
Usage:

help: Display a help message
exit: Exit the expense-tracking cli

`)
	return nil
}

func listExpenses() error {
	// TODO:
	return nil
}
func addExpense() error {
	// TODO:
	return nil
}
func showTotal() error {
	// TODO:
	return nil
}
