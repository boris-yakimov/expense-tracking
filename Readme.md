## expense tracking cli tool

v1.4.0 - ui/ux improvements - TODO

v1.3.0 - password protection and encryption - TODO

v1.2.0 - refactor to SQLite for data persistence - Completed
- Added configurable storage backend (SQLite or JSON)
- Environment variable configuration support

v1.1.0 - refactoring with a TUI
![Recording](assets/recording.gif)

v1.0.0 - cli mvp with base functionality and storing transactions in JSON file
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

## Storage Configuration

The expense tracking tool now supports configurable storage backends. You can choose between SQLite (default) or JSON file storage using environment variables.

### Environment Variables

- `EXPENSE_STORAGE_TYPE`: Set to `"sqlite"` (default) or `"json"`
- `EXPENSE_SQLITE_PATH`: Path to SQLite database file (default: `"db/transactions.db"`)
- `EXPENSE_JSON_PATH`: Path to JSON file (default: `"data.json"`)

### Usage Examples

**Use SQLite storage (default):**
```bash
./expense-tracker
```

**Use JSON storage:**
```bash
EXPENSE_STORAGE_TYPE=json ./expense-tracker
```

**Use custom SQLite path:**
```bash
EXPENSE_SQLITE_PATH=/path/to/my/database.db ./expense-tracker
```

**Use custom JSON path:**
```bash
EXPENSE_STORAGE_TYPE=json EXPENSE_JSON_PATH=/path/to/my/data.json ./expense-tracker
```

### Migration

To migrate data from JSON to SQLite, set the `MIGRATE_TRANSACTION_DATA=true` environment variable:

```bash
MIGRATE_TRANSACTION_DATA=true ./expense-tracker
```

**Note:** Migration only works when using SQLite storage and will load data from the configured JSON file path.
