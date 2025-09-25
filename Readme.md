# expense tracking CLI tool
Track your expenses in the terminal

![Recording](assets/tui-recording.gif)

## Installation 

Download a release from https://github.com/boris-yakimov/expense-tracking/releases  

Or directly download the latest release in your terminal
Linux x86
```sh
wget -qO- https://api.github.com/repos/boris-yakimov/expense-tracking/releases/latest \
  | grep "browser_download_url" \
  | grep "expense-tracking-linux-amd64" \
  | cut -d '"' -f 4 \
  | xargs wget

chmod +x expense-tracking-linux-amd64 

# TODO: to make the db location configurable
mkdir db/
# TODO: steps to add PATH 
./expense-tracking-linux-amd64
```

Linux ARM
```sh
wget -qO- https://api.github.com/repos/boris-yakimov/expense-tracking/releases/latest \
  | grep "browser_download_url" \
  | grep "expense-tracking-linux-arm64" \
  | cut -d '"' -f 4 \
  | xargs wget

chmod +x expense-tracking-linux-arm64 

# TODO: to make the db location configurable
mkdir db/
# TODO: steps to add PATH 
./expense-tracking-linux-arm64
```

Windows  
TODO: powershell  
TODO: cmd  

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
- `EXPENSE_UNENCRYPTED_DB_PATH`: Path to unencrypted SQLite database file (default: `"~/.expense-tracking/transactions.db"`)
- `EXPENSE_ENCRYPTED_DB_PATH`: Path to encrypted database file (default: `"~/.expense-tracking/transactions.enc"`)
- `EXPENSE_JSON_PATH`: Path to JSON file (default: `"~/.expense-tracking/transactions.json"`)
- `EXPENSE_LOG_PATH`: Path to log file (default: `"~/.expense-tracking/expense-tracking.log"`)
- `EXPENSE_SALT_PATH`: Path to salt file (default: `"~/.expense-tracking/transactions.salt"`)

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
EXPENSE_UNENCRYPTED_DB_PATH=/path/to/my/database.db ./expense-tracker
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
