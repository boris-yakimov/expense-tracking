package main

import (
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
