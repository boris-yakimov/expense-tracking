package main

import (
	"bufio"
	"fmt"
	"os"
)

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
		// TODO: list <transcation_type> <amount> <category> <note>
		// TODO: add option list all
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
			// TODO: add <transcation_type> <amount> <category> <note>
			name:        "add",
			description: "Add an expense, income or investment",
			callback:    addExpense,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)

	// print help menu on start
	// TODO: fix this so that I don't have to constantly pass around args even where they are not really needed
	commandHelp([]string{""})

	for {
		fmt.Printf("\n$ expense-tracking > ")

		scanner.Scan()
		input := scanner.Text()
		sanitizedInput := cleanInput(input)

		command, validCommand := supportedCommands[sanitizedInput[0]]
		args := sanitizedInput[1:]

		if validCommand {
			if err := command.callback(args); err != nil {
				fmt.Printf("\n\nError with command: %s\n", command.name)
				fmt.Printf("%s\n", err)
			}
		} else {
			fmt.Println("Unkown command, please run the :help command to see valid options")
		}
	}
}
