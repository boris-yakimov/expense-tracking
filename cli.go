package main

import (
	"fmt"
	"os"
)

type cliCommand struct {
	name        string
	description string
	callback    func(args []string) (success bool, err error)
}

func commandExit(args []string) (success bool, err error) {
	fmt.Println("Closing the expense-tracking tool... goodbye!")
	os.Exit(0)
	return true, nil
}

func commandHelp(args []string) (success bool, err error) {
	if len(args) > 0 {

		// help add
		if args[0] == "add" {
			fmt.Printf(`
Expense Tracking Tool
Usage: add <transaction> <amount> <category> <description>

max char limit for description is 40 characters

example 
        - add expense 12.30 food tacos
        expected output :
        added expense €12.30 | food | tacos
    or
        - add investment 78.00 insurance life insurance monthly payment
        expected output :
        added investment €78.00 | insurance | life insurance monthly payment
`)
			return true, nil
		}

		// help delete
		if args[0] == "delete" {
			fmt.Printf(`
Expense Tracking Tool
Usage: add <transaction> <amount> <category> <description>

example 
        - del expense 33c6ce38
        output :
        successfully removed transaction with id 33c6ce38
`)
			return true, nil
		}
	}

	// default help menu
	fmt.Printf(`
Expense Tracking Tool
Usage:

list:             List transactions
show-total:       Show totals of all transactions
add:              Add a transaction - add <transaction_type> <amount> <category> <description>
del/delete        Delete a transaction - delete <transaction_type> <transaction_id> (transaction IDs can be seen in list and show-total)
help              Display a help message
exit              Exit the expense-tracking tool

Detailed usage:
help add:         Get more info how to use the add command 
help del/delete:  Get more info how to use the delete command
`)

	return true, nil
}
