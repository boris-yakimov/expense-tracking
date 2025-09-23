package main

import (
	"os"
	"testing"
)

func TestLoadTransactions(t *testing.T) {
	testCases := []struct {
		name        string
		storageType StorageType
	}{
		{"SQLite", StorageSQLite},
		{"JSON", StorageJSONFile},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestStorage(t, tc.storageType)

			// Test loading empty storage
			transactions, err := loadTransactionsFromTestStorage()
			if err != nil {
				t.Errorf("Expected no error for empty storage, got %v", err)
			}
			if len(transactions) != 0 {
				t.Errorf("Expected empty transactions, got %v", transactions)
			}

			// Test loading storage with data
			testTransactions := TransactionHistory{
				"2023": {
					"january": {
						"expense": []Transaction{
							{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
						},
					},
				},
			}

			err = saveTransactionsToTestStorage(testTransactions)
			if err != nil {
				t.Fatalf("Failed to save test data: %v", err)
			}

			transactions, err = loadTransactionsFromTestStorage()
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if len(transactions) == 0 {
				t.Errorf("Expected transactions, got empty")
			}

			// Verify the loaded data
			if tx, ok := transactions["2023"]["january"]["expense"]; !ok || len(tx) != 1 {
				t.Errorf("Expected one expense transaction, got %v", transactions)
			}
		})
	}
}

func TestConfigSystem(t *testing.T) {
	// Test default config
	config, err := DefaultConfig()
	if err != nil {
		t.Errorf("Failed to use default config, err: %v", err)
	}
	if config.StorageType != StorageSQLite {
		t.Errorf("Expected default storage type to be SQLite, got %s", config.StorageType)
	}
	if config.SQLitePath != "test_data/transactions.db" {
		t.Errorf("Expected default SQLite path to be 'test_data/transactions.db', got %s", config.SQLitePath)
	}
	if config.JSONFilePath != "test_data/transactions.json" {
		t.Errorf("Expected default JSON path to be 'test_data/transactions.json', got %s", config.JSONFilePath)
	}

	// Test loading config from environment
	os.Setenv("EXPENSE_STORAGE_TYPE", "json")
	os.Setenv("EXPENSE_JSON_PATH", "test.json")
	defer func() {
		os.Unsetenv("EXPENSE_STORAGE_TYPE")
		os.Unsetenv("EXPENSE_JSON_PATH")
	}()

	envConfig, err := loadConfigFromEnvVars()
	if err != nil {
		t.Errorf("Failed to load config from env var, err %v", err)
	}

	if envConfig.StorageType != StorageJSONFile {
		t.Errorf("Expected storage type from env to be JSON, got %s", envConfig.StorageType)
	}
	if envConfig.JSONFilePath != "test.json" {
		t.Errorf("Expected JSON path from env to be 'test.json', got %s", envConfig.JSONFilePath)
	}
}

func TestSetGlobalConfig(t *testing.T) {
	// Test setting global config
	testConfig := &Config{
		StorageType:  StorageJSONFile,
		SQLitePath:   "test.db",
		JSONFilePath: "test.json",
	}

	SetGlobalConfig(testConfig)
	if globalConfig != testConfig {
		t.Errorf("Expected globalConfig to be set to testConfig")
	}
}

func TestLoadConfigFromEnvVars(t *testing.T) {
	// Test with no environment variables
	os.Clearenv()
	config, err := loadConfigFromEnvVars()
	if err != nil {
		t.Errorf("Failed to load config from env var, err %v", err)
	}
	if config.StorageType != StorageSQLite {
		t.Errorf("Expected default storage type to be SQLite, got %s", config.StorageType)
	}

	// Test with SQLite environment variables
	os.Setenv("EXPENSE_STORAGE_TYPE", "sqlite")
	os.Setenv("EXPENSE_SQLITE_PATH", "custom.db")
	defer os.Clearenv()

	config, err = loadConfigFromEnvVars()
	if err != nil {
		t.Errorf("Failed to load config from env var, err %v", err)
	}
	if config.StorageType != StorageSQLite {
		t.Errorf("Expected storage type to be SQLite, got %s", config.StorageType)
	}
	if config.SQLitePath != "custom.db" {
		t.Errorf("Expected SQLite path to be 'custom.db', got %s", config.SQLitePath)
	}

	// Test with JSON environment variables
	os.Setenv("EXPENSE_STORAGE_TYPE", "json")
	os.Setenv("EXPENSE_JSON_PATH", "custom.json")

	config, err = loadConfigFromEnvVars()
	if err != nil {
		t.Errorf("Failed to load config from env var, err %v", err)
	}
	if config.StorageType != StorageJSONFile {
		t.Errorf("Expected storage type to be JSON, got %s", config.StorageType)
	}
	if config.JSONFilePath != "custom.json" {
		t.Errorf("Expected JSON path to be 'custom.json', got %s", config.JSONFilePath)
	}

	// Test with invalid storage type
	os.Setenv("EXPENSE_STORAGE_TYPE", "invalid")
	config, err = loadConfigFromEnvVars()
	if err != nil {
		t.Errorf("Failed to load config from env var, err %v", err)
	}
	if config.StorageType != StorageSQLite {
		t.Errorf("Expected invalid storage type to default to SQLite, got %s", config.StorageType)
	}
}

func TestSaveTransactions(t *testing.T) {
	testCases := []struct {
		name        string
		storageType StorageType
	}{
		{"SQLite", StorageSQLite},
		{"JSON", StorageJSONFile},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestStorage(t, tc.storageType)

			transactions := TransactionHistory{
				"2023": {
					"january": {
						"expense": []Transaction{
							{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
						},
					},
				},
			}

			err := saveTransactionsToTestStorage(transactions)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Verify data was saved by loading it back
			loadedTransactions, err := loadTransactionsFromTestStorage()
			if err != nil {
				t.Errorf("Failed to load saved transactions: %v", err)
			}

			if len(loadedTransactions) == 0 {
				t.Errorf("Expected transactions to be saved")
			}
		})
	}
}

func TestCreateTransactionsTable(t *testing.T) {
	testCases := []struct {
		name        string
		storageType StorageType
	}{
		{"SQLite", StorageSQLite},
		{"JSON", StorageJSONFile},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestStorage(t, tc.storageType)

			// Save test data first
			transactions := TransactionHistory{
				"2023": {
					"january": {
						"expense": []Transaction{
							{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
						},
					},
				},
			}

			err := saveTransactionsToTestStorage(transactions)
			if err != nil {
				t.Fatalf("Failed to save test data: %v", err)
			}

			// Load the data back
			loadedTransactions, err := loadTransactionsFromTestStorage()
			if err != nil {
				t.Fatalf("Failed to load test data: %v", err)
			}

			table := createTransactionsTable("expense", "january", "2023", loadedTransactions)
			if table == nil {
				t.Errorf("Expected table to be created")
			}

			// Test with no transactions
			table = createTransactionsTable("expense", "", "", loadedTransactions)
			if table == nil {
				t.Errorf("Expected table to be created")
			}
		})
	}
}
