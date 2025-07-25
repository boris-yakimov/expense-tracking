package main

import (
	"fmt"
	"os"
)

type cliCommand struct {
	name        string
	description string
	callback    func(args []string) error
}

func commandExit(args []string) error {
	fmt.Println("Closing the expense-tracking tool... goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(args []string) error {
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
