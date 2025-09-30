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
	// Set HOME for test environment
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	testHome := "/tmp/test_home"
	os.Setenv("HOME", testHome)
	os.MkdirAll(testHome, 0755)

	// Test default config
	config, err := DefaultConfig()
	if err != nil {
		t.Errorf("Failed to use default config, err: %v", err)
	}
	if config.StorageType != StorageSQLite {
		t.Errorf("Expected default storage type to be SQLite, got %s", config.StorageType)
	}
	expectedPath := "/tmp/test_home/.expense-tracking/transactions.db"
	if config.UnencryptedDbFile != expectedPath {
		t.Errorf("Expected default SQLite path to be '%s', got %s", expectedPath, config.UnencryptedDbFile)
	}

	defer func() {
		os.Unsetenv("EXPENSE_STORAGE_TYPE")
	}()
}

func TestSetGlobalConfig(t *testing.T) {
	// Test setting global config
	testConfig := &Config{
		UnencryptedDbFile: "test.db",
	}

	SetGlobalConfig(testConfig)
	if globalConfig != testConfig {
		t.Errorf("Expected globalConfig to be set to testConfig")
	}
}

func TestLoadConfigFromEnvVars(t *testing.T) {
	// Test with no environment variables
	os.Clearenv()
	// Set HOME for test
	testHome := "/tmp/test_home"
	os.Setenv("HOME", testHome)
	os.MkdirAll(testHome, 0755)
	config, err := loadConfigFromEnvVars()
	if err != nil {
		t.Errorf("Failed to load config from env var, err %v", err)
	}
	if config.StorageType != StorageSQLite {
		t.Errorf("Expected default storage type to be SQLite, got %s", config.StorageType)
	}

	// Test with SQLite environment variables
	os.Setenv("EXPENSE_STORAGE_TYPE", "sqlite")
	os.Setenv("EXPENSE_UNENCRYPTED_DB_PATH", "custom.db")
	defer os.Clearenv()

	config, err = loadConfigFromEnvVars()
	if err != nil {
		t.Errorf("Failed to load config from env var, err %v", err)
	}
	if config.StorageType != StorageSQLite {
		t.Errorf("Expected storage type to be SQLite, got %s", config.StorageType)
	}
	if config.UnencryptedDbFile != "custom.db" {
		t.Errorf("Expected SQLite path to be 'custom.db', got %s", config.UnencryptedDbFile)
	}

	config, err = loadConfigFromEnvVars()
	if err != nil {
		t.Errorf("Failed to load config from env var, err %v", err)
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

			table := createTransactionsTable("expense", "january", "2023", loadedTransactions, "")
			if table == nil {
				t.Errorf("Expected table to be created")
			}

			// Test with no transactions
			table = createTransactionsTable("expense", "", "", loadedTransactions, "")
			if table == nil {
				t.Errorf("Expected table to be created")
			}
		})
	}
}
