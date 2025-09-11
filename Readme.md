# expense tracking CLI tool
Track your expenses in the terminal

![Recording](assets/login.gif)

## Authentication & Encryption Overview

This project uses **password-based encryption** to protect the SQLite database that stores expense tracking data.  
The system is designed so that the database on disk is stored **encrypted**. It is only decrypted into plaintext form at runtime after the user successfully authenticates and than re-encrypted back after exit(re-encryption also happens if the program is crashed).

---

### 🔑 Authentication Flow

1. **First Run**
   - When the program starts, it checks if an encrypted database file (`encFile`) exists.
   - If not, the user is prompted to create a new password via the **Set New Password** form.
   - That password is stored in memory and used to derive an encryption key.

2. **Subsequent Runs**
   - On startup, the **Login Form** asks the user to enter their password.
   - The entered password is stored temporarily in memory (`userPassword`).
   - That password is used to derive an encryption key and attempt decryption of the database file.

---

### 🔒 Encryption Details

- **Algorithm:** AES-GCM (Galois/Counter Mode)
  - Provides both confidentiality and integrity (authenticates ciphertext).
- **Key Derivation:** A function (`deriveEncryptionKey`) derives a cryptographic key from the user’s password.  
  - This ensures that the actual AES key is never stored or hardcoded.
- **File Format:**
  - Each encrypted file begins with a random **nonce** (generated during encryption).
  - The nonce is followed by the AES-GCM ciphertext (which also contains the authentication tag).

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

TODO: json option is to be re-evaluated in the future, I am not sure i should maintain this as an option at all, SQLite seems the better approach and currently encryption is only handled for sqlite storage
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
