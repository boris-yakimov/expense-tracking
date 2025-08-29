package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
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
			month 			INTEGER NOT NULL CHECK (month BETWEEN 1 and 12)
		);
	`

	_, err = testDb.Exec(testSchema)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Set up test config for SQLite
	testConfig := &Config{
		StorageType:  StorageSQLite,
		SQLitePath:   testDbFilePath,
		JSONFilePath: "",
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
		StorageType:  StorageJSONFile,
		SQLitePath:   "",
		JSONFilePath: testJSONFilePath,
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
	if globalConfig.StorageType == StorageSQLite {
		return loadTransactionsFromTestDb()
	} else if globalConfig.StorageType == StorageJSONFile {
		return loadTransactionsFromTestJSON()
	}
	return nil, fmt.Errorf("unsupported storage type for testing: %s", globalConfig.StorageType)
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
			id, txType, category, description string
			amount                            float64
			year, month                       int
		)

		if err := rows.Scan(&id, &amount, &txType, &category, &description, &year, &month); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		y := fmt.Sprintf("%d", year)
		m := fmt.Sprintf("%02d", month)

		if _, ok := transactions[y]; !ok {
			transactions[y] = make(map[string]map[string][]Transaction)
		}
		if _, ok := transactions[y][m]; !ok {
			transactions[y][m] = make(map[string][]Transaction)
		}

		transactions[y][m][txType] = append(transactions[y][m][txType], Transaction{
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
	if globalConfig.StorageType == StorageSQLite {
		return saveTransactionsToTestDb(transactions)
	} else if globalConfig.StorageType == StorageJSONFile {
		return saveTransactionsToTestJSON(transactions)
	}
	return fmt.Errorf("unsupported storage type for testing: %s", globalConfig.StorageType)
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
			m, err := strconv.Atoi(month)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("invalid month key %q: %w", month, err)
			}

			for txType, list := range types {
				for _, tr := range list {
					_, err = sqlStatement.Exec(
						tr.Id,
						tr.Amount,
						txType,
						tr.Category,
						tr.Description,
						y,
						m,
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
