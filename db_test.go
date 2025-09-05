package main

import (
	"os"
	"testing"
)

func TestInitDb(t *testing.T) {
	// Create a temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_init_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	tmpDbFile.Close()
	defer os.Remove(tmpDbFile.Name())

	// Test successful database initialization
	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Errorf("Expected no error from initDb, got %v", err)
	}

	// Test that database file was created
	if _, err := os.Stat(tmpDbFile.Name()); os.IsNotExist(err) {
		t.Errorf("Expected database file to be created")
	}

	// Clean up
	if db != nil {
		db.Close()
	}
}

func TestInitDbWithExistingFile(t *testing.T) {
	// Create a temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_existing_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	tmpDbFile.Close()
	defer os.Remove(tmpDbFile.Name())

	// Test initialization with existing file
	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Errorf("Expected no error from initDb with existing file, got %v", err)
	}

	// Clean up
	if db != nil {
		db.Close()
	}
}

func TestCloseDb(t *testing.T) {
	// Test closing nil database (should not panic)
	closeDb()

	// Test closing actual database
	tmpDbFile, err := os.CreateTemp("", "test_close_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	tmpDbFile.Close()
	defer os.Remove(tmpDbFile.Name())

	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}

	// Test that closeDb doesn't panic
	closeDb()
}

func TestLoadTransactionsFromDb(t *testing.T) {
	// Create a temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_load_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	tmpDbFile.Close()
	defer os.Remove(tmpDbFile.Name())

	// Initialize database
	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}
	defer closeDb()

	// Test loading from empty database
	transactions, err := loadTransactionsFromDb()
	if err != nil {
		t.Errorf("Expected no error loading from empty db, got %v", err)
	}
	if len(transactions) != 0 {
		t.Errorf("Expected empty transactions from empty db, got %v", transactions)
	}
}

func TestSaveTransactionsToDb(t *testing.T) {
	// Create a temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_save_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	tmpDbFile.Close()
	defer os.Remove(tmpDbFile.Name())

	// Initialize database
	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}
	defer closeDb()

	// Test saving transactions
	testTransactions := TransactionHistory{
		"2023": {
			"01": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	err = saveTransactionsToDb(testTransactions)
	if err != nil {
		t.Errorf("Expected no error saving transactions, got %v", err)
	}

	// Verify transactions were saved
	loadedTransactions, err := loadTransactionsFromDb()
	if err != nil {
		t.Errorf("Expected no error loading transactions, got %v", err)
	}
	if len(loadedTransactions) == 0 {
		t.Errorf("Expected transactions to be saved")
	}
}

func TestSaveTransactionsToDbInvalidData(t *testing.T) {
	// Create a temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_invalid_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	tmpDbFile.Close()
	defer os.Remove(tmpDbFile.Name())

	// Initialize database
	err = initDb(tmpDbFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}
	defer closeDb()

	// Test saving transactions with invalid year
	invalidTransactions := TransactionHistory{
		"invalid_year": {
			"01": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	err = saveTransactionsToDb(invalidTransactions)
	if err == nil {
		t.Errorf("Expected error saving transactions with invalid year")
	}
}
