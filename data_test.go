package main

import (
	"os"
	"testing"
)

func TestLoadTransactions(t *testing.T) {
	tmpFile := "test_load_transactions.json"
	originalFilePath := transactionsFilePath
	transactionsFilePath = tmpFile

	// Clean up after test
	defer func() {
		transactionsFilePath = originalFilePath
		os.Remove(tmpFile)
	}()

	// Test loading non-existent file
	transactions, err := loadTransactions()
	if err != nil {
		t.Errorf("Expected no error for non-existent file, got %v", err)
	}
	if len(transactions) != 0 {
		t.Errorf("Expected empty transactions, got %v", transactions)
	}

	// Test loading existing file
	testData := `{"2023":{"01":{"expense":[{"id":"1","amount":10.0,"category":"food","description":"test"}]}}}`
	os.WriteFile(tmpFile, []byte(testData), 0644)

	transactions, err = loadTransactions()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(transactions) == 0 {
		t.Errorf("Expected transactions, got empty")
	}
}

func TestSaveTransactions(t *testing.T) {
	tmpFile := "test_save_transactions.json"
	originalFilePath := transactionsFilePath
	transactionsFilePath = tmpFile

	// Clean up after test
	defer func() {
		transactionsFilePath = originalFilePath
		os.Remove(tmpFile)
	}()

	transactions := TransactionHistory{
		"2023": {
			"01": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	err := saveTransactions(transactions)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Errorf("Expected file to be created")
	}
}

func TestCreateTransactionsTable(t *testing.T) {
	transactions := TransactionHistory{
		"2023": {
			"01": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	table := createTransactionsTable("expense", "01", "2023", transactions)
	if table == nil {
		t.Errorf("Expected table to be created")
	}

	// Test with no transactions
	table = createTransactionsTable("expense", "", "", transactions)
	if table == nil {
		t.Errorf("Expected table to be created")
	}
}
