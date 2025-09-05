package main

import (
	"os"
	"testing"
)

func TestMigrateJsonToDb(t *testing.T) {
	// Create temporary JSON file with test data
	tmpJSONFile, err := os.CreateTemp("", "test_migrate_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(tmpJSONFile.Name())

	// Create temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_migrate_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	defer os.Remove(tmpDbFile.Name())

	// Set up test config for JSON
	jsonConfig := &Config{
		StorageType:  StorageJSONFile,
		JSONFilePath: tmpJSONFile.Name(),
	}
	SetGlobalConfig(jsonConfig)

	// Save test data to JSON
	testTransactions := TransactionHistory{
		"2023": {
			"january": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test food"},
				},
				"income": []Transaction{
					{Id: "2", Amount: 1000.0, Category: "salary", Description: "test salary"},
				},
			},
		},
	}

	err = saveTransactionsToJsonFile(testTransactions)
	if err != nil {
		t.Fatalf("Failed to save test data to JSON: %v", err)
	}

	// Initialize database
	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer closeDb()

	// Test migration
	err = migrateJsonToDb()
	if err != nil {
		t.Errorf("Expected no error during migration, got %v", err)
	}

	// Verify data was migrated by loading from database
	transactions, err := loadTransactionsFromDb()
	if err != nil {
		t.Errorf("Expected no error loading from database after migration, got %v", err)
	}
	if len(transactions) == 0 {
		t.Errorf("Expected transactions to be migrated to database")
	}
}

func TestMigrateJsonToDbEmptyJson(t *testing.T) {
	// Create empty JSON file
	tmpJSONFile, err := os.CreateTemp("", "test_migrate_empty_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(tmpJSONFile.Name())

	// Create temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_migrate_empty_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	defer os.Remove(tmpDbFile.Name())

	// Set up test config for JSON
	jsonConfig := &Config{
		StorageType:  StorageJSONFile,
		JSONFilePath: tmpJSONFile.Name(),
	}
	SetGlobalConfig(jsonConfig)

	// Initialize database
	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer closeDb()

	// Test migration with empty JSON
	err = migrateJsonToDb()
	if err != nil {
		// The function returns EOF error for empty JSON files, which is expected
		if err.Error() != "failed to load JSON transactions: EOF" {
			t.Errorf("Expected EOF error during migration with empty JSON, got %v", err)
		}
	}
}

func TestMigrateJsonToDbInvalidData(t *testing.T) {
	// Create JSON file with invalid data
	tmpJSONFile, err := os.CreateTemp("", "test_migrate_invalid_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(tmpJSONFile.Name())

	// Create temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_migrate_invalid_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	defer os.Remove(tmpDbFile.Name())

	// Set up test config for JSON
	jsonConfig := &Config{
		StorageType:  StorageJSONFile,
		JSONFilePath: tmpJSONFile.Name(),
	}
	SetGlobalConfig(jsonConfig)

	// Save invalid data to JSON (invalid year)
	invalidTransactions := TransactionHistory{
		"invalid_year": {
			"january": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test food"},
				},
			},
		},
	}

	err = saveTransactionsToJsonFile(invalidTransactions)
	if err != nil {
		t.Fatalf("Failed to save invalid test data to JSON: %v", err)
	}

	// Initialize database
	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer closeDb()

	// Test migration with invalid data
	err = migrateJsonToDb()
	if err == nil {
		t.Errorf("Expected error during migration with invalid data")
	}
}

func TestMigrateJsonToDbInvalidMonth(t *testing.T) {
	// Create JSON file with invalid month
	tmpJSONFile, err := os.CreateTemp("", "test_migrate_invalid_month_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(tmpJSONFile.Name())

	// Create temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_migrate_invalid_month_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	defer os.Remove(tmpDbFile.Name())

	// Set up test config for JSON
	jsonConfig := &Config{
		StorageType:  StorageJSONFile,
		JSONFilePath: tmpJSONFile.Name(),
	}
	SetGlobalConfig(jsonConfig)

	// Save invalid data to JSON (invalid month)
	invalidTransactions := TransactionHistory{
		"2023": {
			"invalid_month": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test food"},
				},
			},
		},
	}

	err = saveTransactionsToJsonFile(invalidTransactions)
	if err != nil {
		t.Fatalf("Failed to save invalid test data to JSON: %v", err)
	}

	// Initialize database
	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer closeDb()

	// Test migration with invalid month
	err = migrateJsonToDb()
	if err == nil {
		t.Errorf("Expected error during migration with invalid month")
	}
}
