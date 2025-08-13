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
			description: "List transactions",
			callback:    listAllTransactions,
		},
		"show-total": {
			name:        "show-total",
			description: "Show totals of all transactions",
			callback:    showTotal,
		},
		"add": {
			name:        "add",
			description: "Add a transaction (expense, investment or income)",
			callback:    addTransaction,
		},
		"delete": {
			name:        "delete",
			description: "Delete a transaction (expense, investment or income)",
			callback:    deleteTransaction,
		},
		"update": {
			name:        "update",
			description: "Update a transaction (expense, investment or income)",
			callback:    updateTransaction,
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
		sanitizedInput := cleanTerminalInput(input)

		// if blank enter just prompt again
		if len(sanitizedInput) == 0 {
			continue
		}

		inputCommand := sanitizedInput[0]
		args := sanitizedInput[1:]

		command, validCommand := supportedCommands[inputCommand]
		cmdMatches := []string{}
		// TODO: functions like calculateMonthPnL, addTransaction, deleteTransaction, updateTransaction print to stdout. maybe they should return data and let CLI-layer functions handle printing.
		if validCommand {
			if _, err := command.callback(args); err != nil {
				fmt.Printf("\n\nError with command: %s\n", command.name)
				fmt.Printf("%s\n", err)
			}
			// if successful command run just prompt again
			continue
		} else {
			// try partial command match
			for cmd := range supportedCommands {
				if len(inputCommand) > 0 && len(cmd) >= len(inputCommand) && cmd[:len(inputCommand)] == inputCommand {
					cmdMatches = append(cmdMatches, cmd)
				}
			}

			if len(cmdMatches) == 1 {
				command = supportedCommands[cmdMatches[0]]
				if _, err := command.callback(args); err != nil {
					fmt.Printf("\n\nError with command: %s\n", command.name)
					fmt.Printf("%s\n", err)
				}
				// if successful command run just prompt again
				continue
			} else if len(cmdMatches) > 1 {
				fmt.Println("did you mean one of these?")
				for _, m := range cmdMatches {
					fmt.Printf("  - %s\n", m)
				}
				// give suggestion and re-prompt
				continue
			}

			// fallback: unknown command
			fmt.Printf("Unkown command: \"%s\", please run the :help command to see valid options", inputCommand)
		}
	}
}
