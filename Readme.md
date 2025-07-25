## expense tracking cli tool - IN EARLY DEVELOPMENT AND EXPERIMENTATION

### Some examples from the current version as of July 2025
List (current month)
```
expense-tracking > list

Year: 2025
  Month: July
     1. $53.40    | food       | groceries
     2. $5432.00  | renovation | fence
     3. $45.00    | food       | burgers
     4. $12.35    | food       | sandwitch
     5. $100.00   | travel     | rent a car gas
```

Show total
```
expense-tracking > show-total

Summary for July 2025
+------------------------------------------------------------------------+
|  1. $53.40    | food         | groceries                                |
|  2. $45.00    | food         | burgers                                  |
|  3. $12.35    | food         | sandwitch                                |
|  4. $5432.00  | renovation   | fence                                    |
|  5. $100.00   | travel       | rent a car gas                           |
+------------------------------------------------------------------------+
Total expenses: $5642.75
```

Help
```
expense-tracking > help                                                                                                                                                                                                                                                                 
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
```

Exit
```
expense-tracking > exit
Closing the expense-tracking tool... goodbye!
```

Add
TODO:
