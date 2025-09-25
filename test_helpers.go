package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/pbkdf2"
)

var testDb *sql.DB
var testJSONFilePath string
var originalConfig *Config

// setupTestStorage creates temporary storage for testing (either SQLite or JSON)
func setupTestStorage(t *testing.T, storageType StorageType) {
	// Save original config
	originalConfig = globalConfig

	switch storageType {
	case StorageSQLite:
		setupTestDb(t)
	case StorageJSONFile:
		setupTestJSON(t)
	default:
		t.Fatalf("Unsupported storage type for testing: %s", storageType)
	}
}

// setupTestDb creates a temporary test database with a transactions table
func setupTestDb(t *testing.T) {
	// Create temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_transactions_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}

	// Store the temp file path for cleanup
	testDbFilePath := tmpDbFile.Name()
	tmpDbFile.Close()

	// Initialize test database
	testDb, err = sql.Open("sqlite3", testDbFilePath)
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}

	// Create transactions table (same as main table for compatibility)
	testSchema := `
		CREATE TABLE transactions (
			id				  TEXT PRIMARY KEY,
			amount 			NUMERIC(12, 2) NOT NULL,
			type 				TEXT NOT NULL CHECK (type IN ('income', 'expense', 'investment')),
			category 		TEXT NOT NULL,
			description TEXT,
			year 				INTEGER NOT NULL,
			month 			TEXT NOT NULL CHECK (
				month IN (
            'january','february','march','april','may','june',
            'july','august','september','october','november','december'
				)
			)
		);
	`

	_, err = testDb.Exec(testSchema)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Set up test config for SQLite
	testConfig := &Config{
		StorageType:       StorageSQLite,
		UnencryptedDbFile: testDbFilePath,
		JSONFilePath:      "",
	}
	SetGlobalConfig(testConfig)

	// Replace the global db with testDb for the duration of the test
	originalDb := db
	db = testDb

	// Clean up function
	t.Cleanup(func() {
		db = originalDb
		testDb.Close()
		os.Remove(testDbFilePath)
		SetGlobalConfig(originalConfig)
	})
}

// setupTestJSON creates a temporary JSON file for testing
func setupTestJSON(t *testing.T) {
	// Create temporary JSON file
	tmpJSONFile, err := os.CreateTemp("", "test_transactions_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	testJSONFilePath = tmpJSONFile.Name()
	tmpJSONFile.Close()

	// Set up test config for JSON
	testConfig := &Config{
		StorageType:       StorageJSONFile,
		UnencryptedDbFile: "",
		JSONFilePath:      testJSONFilePath,
	}
	SetGlobalConfig(testConfig)

	// Clean up function
	t.Cleanup(func() {
		os.Remove(testJSONFilePath)
		SetGlobalConfig(originalConfig)
	})
}

// loadTransactionsFromTestStorage loads transactions from the current test storage (SQLite or JSON)
func loadTransactionsFromTestStorage() (TransactionHistory, error) {
	switch globalConfig.StorageType {
	case StorageSQLite:
		return loadTransactionsFromTestDb()
	case StorageJSONFile:
		return loadTransactionsFromTestJSON()
	default:
		return nil, fmt.Errorf("unsupported storage type for testing: %s", globalConfig.StorageType)
	}
}

// loadTransactionsFromTestDb loads transactions from the transactions table (in test db)
func loadTransactionsFromTestDb() (TransactionHistory, error) {
	rows, err := db.Query(`
			SELECT id, amount, type, category, description, year, month
			FROM transactions
		`)
	if err != nil {
		return nil, fmt.Errorf("failed to execute load transactions sql query: %w", err)
	}
	defer rows.Close()

	transactions := make(TransactionHistory)

	for rows.Next() {
		var (
			id, txType, category, description, month string
			amount                                   float64
			year                                     int
		)

		if err := rows.Scan(&id, &amount, &txType, &category, &description, &year, &month); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		y := fmt.Sprintf("%d", year)

		if _, ok := transactions[y]; !ok {
			transactions[y] = make(map[string]map[string][]Transaction)
		}
		if _, ok := transactions[y][month]; !ok {
			transactions[y][month] = make(map[string][]Transaction)
		}

		transactions[y][month][txType] = append(transactions[y][month][txType], Transaction{
			Id:          id,
			Amount:      amount,
			Category:    category,
			Description: description,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed during transaction loading: %w", err)
	}

	return transactions, nil
}

// loadTransactionsFromTestJSON loads transactions from the test JSON file
func loadTransactionsFromTestJSON() (TransactionHistory, error) {
	file, err := os.Open(testJSONFilePath)
	if os.IsNotExist(err) {
		return make(TransactionHistory), nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var transactions TransactionHistory
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&transactions)
	if err != nil {
		// If the file exists but is empty or has EOF, return empty transactions
		if err.Error() == "EOF" {
			return make(TransactionHistory), nil
		}
		return nil, err
	}
	return transactions, nil
}

// saveTransactionsToTestStorage saves transactions to the current test storage (SQLite or JSON)
func saveTransactionsToTestStorage(transactions TransactionHistory) error {
	switch globalConfig.StorageType {
	case StorageSQLite:
		return saveTransactionsToTestDb(transactions)
	case StorageJSONFile:
		return saveTransactionsToTestJSON(transactions)
	default:
		return fmt.Errorf("unsupported storage type for testing: %s", globalConfig.StorageType)
	}
}

// saveTransactionsToTestDb saves transactions to the transactions table (in test db)
func saveTransactionsToTestDb(transactions TransactionHistory) error {
	// Clear existing data first
	_, err := db.Exec("DELETE FROM transactions")
	if err != nil {
		return fmt.Errorf("failed to clear test transactions: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin save transaction failed: %w", err)
	}

	sqlStatement, err := tx.Prepare(`
			INSERT INTO transactions
			(id, amount, type, category, description, year, month)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare insert during save transaction failed: %w", err)
	}
	defer sqlStatement.Close()

	for year, months := range transactions {
		y, err := strconv.Atoi(year)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("invalid year key %q: %w", year, err)
		}

		for month, types := range months {

			for txType, list := range types {

				for _, tr := range list {
					_, err = sqlStatement.Exec(
						tr.Id,
						tr.Amount,
						txType,
						tr.Category,
						tr.Description,
						y,     // integer, e.g. 2025
						month, // string, e.g. August
					)
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("insert failed for transaction %s: %w", tr.Id, err)
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// saveTransactionsToTestJSON saves transactions to the test JSON file
func saveTransactionsToTestJSON(transactions TransactionHistory) error {
	file, err := os.Create(testJSONFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(transactions)
}

// setupTestEncryption creates temporary directories for encryption testing
func setupTestEncryption(t *testing.T) (string, string) {
	// Create temporary directory for encryption files
	tmpDir, err := os.MkdirTemp("", "test_encryption_*")
	if err != nil {
		t.Fatalf("Failed to create temp encryption dir: %v", err)
	}

	// Create temporary encryption file paths
	testEncFile := filepath.Join(tmpDir, "test_transactions.enc")
	testSaltFile := filepath.Join(tmpDir, "test_transactions.salt")

	t.Cleanup(func() {
		// Clean up temporary files
		os.RemoveAll(tmpDir)
	})

	return testEncFile, testSaltFile
}

// testEncryptDatabase encrypts a database file using test-specific paths
func testEncryptDatabase(_ *testing.T, dbPath, testEncFile, testSaltFile string) error {
	if userPassword == "" {
		return fmt.Errorf("user password not set")
	}

	dbData, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read database file: %w", err)
	}

	// Use test-specific salt file
	salt, err := testGetOrCreateSalt(testSaltFile)
	if err != nil {
		return fmt.Errorf("failed to get salt: %w", err)
	}

	key := pbkdf2.Key([]byte(userPassword), salt, iterations, keyLen, sha256.New)

	encryptedData, err := encryptTransactions(key, dbData)
	if err != nil {
		return fmt.Errorf("failed to encrypt database: %w", err)
	}

	// Write to test-specific encrypted file
	if err := os.WriteFile(testEncFile, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted database: %w", err)
	}

	return nil
}

// testDecryptDatabase decrypts a database file using test-specific paths
func testDecryptDatabase(_ *testing.T, dbPath, testEncFile, testSaltFile string) error {
	if userPassword == "" {
		return fmt.Errorf("user password not set")
	}

	// Check if encrypted file exists
	if _, err := os.Stat(testEncFile); os.IsNotExist(err) {
		return nil // nothing to decrypt
	}

	encryptedData, err := os.ReadFile(testEncFile)
	if err != nil {
		return fmt.Errorf("failed to read encrypted database: %w", err)
	}

	// Use test-specific salt file
	salt, err := testLoadSalt(testSaltFile)
	if err != nil {
		return fmt.Errorf("failed to load salt: %w", err)
	}

	key := pbkdf2.Key([]byte(userPassword), salt, iterations, keyLen, sha256.New)

	decryptedData, err := decryptTransactions(key, encryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt database: %w", err)
	}

	// Write decrypted data to database file
	if err := os.WriteFile(dbPath, decryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted database: %w", err)
	}

	return nil
}

// testGetOrCreateSalt gets or creates salt using test-specific path
func testGetOrCreateSalt(testSaltFile string) ([]byte, error) {
	// if it exists return it
	if _, err := os.Stat(testSaltFile); err == nil {
		return testLoadSalt(testSaltFile)
		// if it doesn't create it and than return it
	} else if os.IsNotExist(err) {
		salt, err := generateSalt()
		if err != nil {
			return nil, err
		}

		if err := testSaveSalt(salt, testSaltFile); err != nil {
			return nil, err
		}

		return salt, nil
	} else {
		// unexpected error
		return nil, fmt.Errorf("failed to check salt file: %w", err)
	}
}

// testLoadSalt loads salt from test-specific file
func testLoadSalt(testSaltFile string) ([]byte, error) {
	salt, err := os.ReadFile(testSaltFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("salt file not found: %w", err)
		}
		return nil, fmt.Errorf("failed to load salt: %w", err)
	}

	if len(salt) != saltLen {
		return nil, fmt.Errorf("invalid salt length: expected %d, got %d", saltLen, len(salt))
	}

	return salt, nil
}

// testSaveSalt saves salt to test-specific file
func testSaveSalt(salt []byte, testSaltFile string) error {
	dir := filepath.Dir(testSaltFile)

	// make sure dir exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create salt directory: %w", err)
	}

	if err := os.WriteFile(testSaltFile, salt, 0600); err != nil {
		return fmt.Errorf("failed to save salt: %w", err)
	}

	return nil
}
