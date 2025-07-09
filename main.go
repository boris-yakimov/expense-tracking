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
			description: "Exit the LZ cli",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Display a help message",
			callback:    commandHelp,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("itgix-landing-zone > ")

		scanner.Scan()
		input := scanner.Text()
		sanitizedInput := cleanInput(input)

		command, validCommand := supportedCommands[sanitizedInput[0]]

		if validCommand {

			if command.name == "exit" {
				exitErr := supportedCommands["exit"].callback()
				if exitErr != nil {
					fmt.Printf("exit error: %s\n", exitErr)
				}
			}

			if command.name == "help" {
				helpErr := supportedCommands["help"].callback()
				if helpErr != nil {
					fmt.Printf("help error: %s\n", helpErr)
				}
			}

		} else {
			fmt.Println("Unkown command")
		}
	}
}

func cleanInput(text string) []string {
	var sanitizedText = strings.Trim(strings.ToLower(text), " ")
	return strings.Split(sanitizedText, " ")
}

func commandExit() error {
	fmt.Println("Closing the LZ cli... goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Printf(`
ITGix AWS Landing Zone
Usage:

help: Display a help message
exit: Exit the Landing Zone cli

`)
	return nil
}
