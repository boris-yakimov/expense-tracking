package main

import (
	"fmt"
	"os"
)

// TODO: refactoring cli to a TUI approach

// type cliCommand struct {
// 	name        string
// 	description string
// 	callback    func(args []string) (success bool, err error)
// }

// var supportedCommands = map[string]cliCommand{
// 	"exit": {
// 		name:        "exit",
// 		description: "Exit the expense-tracking tool",
// 		callback:    commandExit,
// 	},
// 	"help": {
// 		name:        "help",
// 		description: "Display a help message",
// 		callback:    commandHelp,
// 	},
// 	"list": {
// 		name:        "list",
// 		description: "Show a list of transactions and p&l",
// 		callback:    visualizeTransactions,
// 	},
// 	"show": {
// 		name:        "show",
// 		description: "Alias to list - show a list of transactions and p&l",
// 		callback:    visualizeTransactions,
// 	},
// 	"add": {
// 		name:        "add",
// 		description: "Add a transaction (expense, investment or income)",
// 		callback:    addTransaction,
// 	},
// 	"delete": {
// 		name:        "delete",
// 		description: "Delete a transaction (expense, investment or income)",
// 		callback:    deleteTransaction,
// 	},
// 	"update": {
// 		name:        "update",
// 		description: "Update a transaction (expense, investment or income)",
// 		callback:    updateTransaction,
// 	},
// }

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
Usage:

add <transaction> <amount> <category> <description>

max char limit for description is %v characters

example 
        - add expense 12.30 food tacos
        expected output :
        added expense €12.30 | food | tacos
    or
        - add investment 78.00 insurance life insurance monthly payment
        expected output :
        added investment €78.00 | insurance | life insurance monthly payment
`, descriptionMaxCharLength)
			return true, nil
		}

		// help update
		if args[0] == "update" {
			fmt.Printf(`
Expense Tracking Tool
Usage:

update <transaction_type> <transaction_id> <amount> <category> <description>

max char limit for description is %v characters

example 
        - update expense b3f921d7 33.62 food switch back to pizza
        output :
        transaction successully updated 
`, descriptionMaxCharLength)
			return true, nil
		}

		// help delete
		if args[0] == "delete" {
			fmt.Printf(`
Expense Tracking Tool
Usage:

delete <transaction_type> <transaction_id>

example 
        - del expense 33c6ce38
        output :
        successfully removed transaction with id 33c6ce38
`)
			return true, nil
		}

		// help show/list
		if args[0] == "show" || args[0] == "list" {
			fmt.Printf(`
Expense Tracking Tool
Usage: 

show and list are aliases and can be used interchangably

show                      - shows all transactions and P&Ls
show <month> <year>       - shows transactions and P&L for specific month 
show <year>               - shows only P&L for the specific year

or 

list                      - shows all transactions and P&Ls
list <month> <year>       - shows transactions and P&L for specific month 
list <year>               - shows only P&L for the specific year
`)
			return true, nil
		}
	}

	// default help menu
	fmt.Printf(`
Expense Tracking Tool

Usage:

list                    List transactions - list (without arguments) ; list <month> <year> ; list <year>
show                    An alias to the list command
add:                    Add a transaction - add <transaction_type> <amount> <category> <description>
delete (del)            Delete a transaction - delete <transaction_type> <transaction_id> (transaction IDs can be seen via the list command)
update                  Update transaction - update <transaction_type> <transaction_id> <amount> <category> <description>
help                    Display a help message
exit                    Exit the expense-tracking tool
     
Detailed usage:     

help add                Get more info on add command 
help delete (del)       Get more info on delete command
help update             Get more info on update command
help show               Get more info on show command
help list               Get more info on list command
`)
	return true, nil
}
