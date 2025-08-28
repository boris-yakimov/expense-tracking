## expense tracking cli tool - EARLY DEVELOPMENT AND EXPERIMENTATION

v1.3.0 - password protection and encryption

v1.2.0 - refactor to SQLite for data persistence - In Progress

v1.1.0 - refactoring with a TUI - DONE
![Recording](assets/recording.gif)

v1.0.0 - cli mvp with base functionality and storing transactions in JSON file - DONE
```
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
```



list transactions
```
$ expense-tracking > list 
====================
August 2025
====================


  Income
  ------
    ID          Amount      Category    Description
    --          ------      --------    -----------
    5c45c8a3    €1500.00    salary      income from employment
    ec6cbc9a    €500.00     rentals     apartment rental


  Expense
  -------
    ID          Amount     Category    Description
    --          ------     --------    -----------
    d529cb01    €200.00    cash        atm withdrawal
    2a11aa59    €12.30     food        tacos
    b3f921d7    €33.62     food        switch back to pizza
    a2b734a0    €9.68      food        sandwitch
    59c494be    €15.00     food        patato


  Investment
  ----------
    ID          Amount    Category     Description
    --          ------    --------     -----------
    023b3574    €19.25    insurance    prorperty insurance august 2025
    076e2589    €78.00    insurance    life insurance august 2025
    9c022af9    €50.00    crypto       eth

  P&L Result: €1582.15 | 79.1%

```
