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
			description: "Add an expense, income or investment",
			callback:    addExpense,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("\nexpense-tracking > ")
		// TODO: opening the tool immediately shows the help menu

		scanner.Scan()
		input := scanner.Text()
		sanitizedInput := cleanInput(input)

		command, validCommand := supportedCommands[sanitizedInput[0]]
		args := sanitizedInput[1:]

		if validCommand {
			if err := command.callback(args); err != nil {
				fmt.Printf("\n\ncommand: %s <amount> <category> <note>\n", command.name)
				fmt.Printf("%s\n", err)
			}
		} else {
			fmt.Println("Unkown command, please run the :help command to see valid options")
		}
	}
}
