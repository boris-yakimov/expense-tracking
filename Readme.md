## expense tracking cli tool
Track your expenses in the terminal

![Recording](assets/recording.gif)

```

## Storage Configuration

The expense tracking tool now supports configurable storage backends. You can choose between SQLite (default) or JSON file storage using environment variables.

### Environment Variables

- `EXPENSE_STORAGE_TYPE`: Set to `"sqlite"` (default) or `"json"`
- `EXPENSE_SQLITE_PATH`: Path to SQLite database file (default: `"db/transactions.db"`)
- `EXPENSE_JSON_PATH`: Path to JSON file (default: `"db/transactions.json"`)

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
