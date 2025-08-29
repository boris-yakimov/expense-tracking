package main

import (
	"os"
	"testing"
)

func TestLoadTransactions(t *testing.T) {
	setupTestDb(t)

	// Test loading empty database
	transactions, err := loadTransactionsFromTestDb()
	if err != nil {
		t.Errorf("Expected no error for empty database, got %v", err)
	}
	if len(transactions) != 0 {
		t.Errorf("Expected empty transactions, got %v", transactions)
	}

	// Test loading database with data
	testTransactions := TransactionHistory{
		"2023": {
			"01": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	err = saveTransactionsToTestDb(testTransactions)
	if err != nil {
		t.Fatalf("Failed to save test data: %v", err)
	}

	transactions, err = loadTransactionsFromTestDb()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(transactions) == 0 {
		t.Errorf("Expected transactions, got empty")
	}

	// Verify the loaded data
	if tx, ok := transactions["2023"]["01"]["expense"]; !ok || len(tx) != 1 {
		t.Errorf("Expected one expense transaction, got %v", transactions)
	}
}

func TestConfigSystem(t *testing.T) {
	// Test default config
	config := DefaultConfig()
	if config.StorageType != StorageSQLite {
		t.Errorf("Expected default storage type to be SQLite, got %s", config.StorageType)
	}
	if config.SQLitePath != "db/transactions.db" {
		t.Errorf("Expected default SQLite path to be 'db/transactions.db', got %s", config.SQLitePath)
	}
	if config.JSONFilePath != "db/transactions.json" {
		t.Errorf("Expected default JSON path to be 'data.json', got %s", config.JSONFilePath)
	}

	// Test loading config from environment
	os.Setenv("EXPENSE_STORAGE_TYPE", "json")
	os.Setenv("EXPENSE_JSON_PATH", "test.json")
	defer func() {
		os.Unsetenv("EXPENSE_STORAGE_TYPE")
		os.Unsetenv("EXPENSE_JSON_PATH")
	}()

	envConfig := loadConfigFromEnvVars()
	if envConfig.StorageType != StorageJSONFile {
		t.Errorf("Expected storage type from env to be JSON, got %s", envConfig.StorageType)
	}
	if envConfig.JSONFilePath != "test.json" {
		t.Errorf("Expected JSON path from env to be 'test.json', got %s", envConfig.JSONFilePath)
	}
}

func TestSaveTransactions(t *testing.T) {
	setupTestDb(t)

	transactions := TransactionHistory{
		"2023": {
			"01": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	err := saveTransactionsToTestDb(transactions)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify data was saved by loading it back
	loadedTransactions, err := loadTransactionsFromTestDb()
	if err != nil {
		t.Errorf("Failed to load saved transactions: %v", err)
	}

	if len(loadedTransactions) == 0 {
		t.Errorf("Expected transactions to be saved")
	}
}

func TestCreateTransactionsTable(t *testing.T) {
	setupTestDb(t)

	// Save test data first
	transactions := TransactionHistory{
		"2023": {
			"01": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	err := saveTransactionsToTestDb(transactions)
	if err != nil {
		t.Fatalf("Failed to save test data: %v", err)
	}

	// Load the data back
	loadedTransactions, err := loadTransactionsFromTestDb()
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	table := createTransactionsTable("expense", "01", "2023", loadedTransactions)
	if table == nil {
		t.Errorf("Expected table to be created")
	}

	// Test with no transactions
	table = createTransactionsTable("expense", "", "", loadedTransactions)
	if table == nil {
		t.Errorf("Expected table to be created")
	}
}
