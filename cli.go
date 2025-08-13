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

// TODO: functions like calculateMonthPnL, addTransaction, deleteTransaction, updateTransaction print to stdout. maybe they should return data and let CLI-layer functions handle printing.
var supportedCommands = map[string]cliCommand{
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
`, descriptionMaxLength)
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
`, descriptionMaxLength)
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

		// help show-total
		if args[0] == "show-total" {
			fmt.Printf(`
Expense Tracking Tool
Usage: 

show-total                      - shows all transactions and P&Ls
show-total <month> <year>       - shows transactions and P&L for specific month 
show-total <year>               - shows only P&L for the specific year


-----------------------------------------------------------------------------------------------


example:          show-total
expected output:

Year: 2025
  Month: august

    expense
     1. €13.49    | food       | test food in next month
     2. €200.00   | cash       | atm withdrawal
     3. €12.30    | food       | tacos
     4. €33.62    | food       | switch back to pizza

    investment
     1. €19.25    | insurance  | prorperty insurance august 2025
     2. €78.00    | insurance  | life insurance august 2025

    income
     1. €1500.00  | salary     | income from employment


p&l result: €1143.34 | 76.2%%

  Month: july

    expense
     1. €53.40    | food       | groceries
     2. €432.00   | renovation | fence
     3. €45.00    | food       | burgers
     4. €12.35    | food       | sandwitch
     5. €100.00   | travel     | rent a car gas
     6. €52.00    | food       | groceries

    investment
     1. €78.00    | insurance  | life insurance monthly payment
     2. €19.25    | insurance  | property insurance monthly payment
     3. €100.00   | funds      | ibkr etf investment

    income
     1. €1500.00  | salary     | income from employment


p&l result: €608.00 | 40.5%%


-----------------------------------------------------------------------------------------------


example:          show-total 2025
expected output:

p&l result: €1751.34 | 58.4%%


-----------------------------------------------------------------------------------------------


example:          show-total july 2025
expected output:

	expense
     1. €53.40    | food       | groceries
     2. €432.00   | renovation | fence
     3. €45.00    | food       | burgers
     4. €12.35    | food       | sandwitch
     5. €100.00   | travel     | rent a car gas
     6. €52.00    | food       | groceries

investment
     1. €78.00    | insurance  | life insurance monthly payment
     2. €19.25    | insurance  | property insurance monthly payment
     3. €100.00   | funds      | ibkr etf investment

income
     1. €1500.00  | salary     | income from employment

p&l result: €608.00 | 40.5%%

`)
			return true, nil
		}
	}

	// default help menu
	fmt.Printf(`
Expense Tracking Tool

Usage:

list                    List transactions
show-total              Show totals of all transactions
add:                    Add a transaction - add <transaction_type> <amount> <category> <description>
delete (del)            Delete a transaction - delete <transaction_type> <transaction_id> (transaction IDs can be seen in list and show-total)
update                  Update transaction - update <transaction_type> <transaction_id> <amount> <category> <description>
help                    Display a help message
exit                    Exit the expense-tracking tool
     
Detailed usage:     

help add                Get more info on add command 
help delete (del)       Get more info on delete command
help update             Get more info on update command
help show-total         Get more info on show-total command
`)
	return true, nil
}
