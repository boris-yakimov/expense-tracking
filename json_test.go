package main

import (
	"os"
	"testing"
)

func TestLoadTransactionsFromJsonFile(t *testing.T) {
	// Create temporary JSON file
	tmpJSONFile, err := os.CreateTemp("", "test_load_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(tmpJSONFile.Name())

	// Set up test config
	testConfig := &Config{
		StorageType:  StorageJSONFile,
		JSONFilePath: tmpJSONFile.Name(),
	}
	SetGlobalConfig(testConfig)

	// Test loading from non-existent file
	transactions, err := loadTransactionsFromJsonFile()
	if err != nil {
		// The function returns EOF error for non-existent files, which is expected
		if err.Error() != "EOF" {
			t.Errorf("Expected EOF error loading from non-existent file, got %v", err)
		}
	}
	if len(transactions) != 0 {
		t.Errorf("Expected empty transactions from non-existent file, got %v", transactions)
	}

	// Test loading from existing file with data
	testTransactions := TransactionHistory{
		"2023": {
			"01": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	// Save test data
	err = saveTransactionsToJsonFile(testTransactions)
	if err != nil {
		t.Fatalf("Failed to save test data: %v", err)
	}

	// Load and verify
	transactions, err = loadTransactionsFromJsonFile()
	if err != nil {
		t.Errorf("Expected no error loading from existing file, got %v", err)
	}
	if len(transactions) == 0 {
		t.Errorf("Expected transactions to be loaded")
	}
}

func TestSaveTransactionsToJsonFile(t *testing.T) {
	// Create temporary JSON file
	tmpJSONFile, err := os.CreateTemp("", "test_save_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(tmpJSONFile.Name())

	// Set up test config
	testConfig := &Config{
		StorageType:  StorageJSONFile,
		JSONFilePath: tmpJSONFile.Name(),
	}
	SetGlobalConfig(testConfig)

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

	err = saveTransactionsToJsonFile(testTransactions)
	if err != nil {
		t.Errorf("Expected no error saving transactions, got %v", err)
	}

	// Verify file was created and has content
	fileInfo, err := os.Stat(tmpJSONFile.Name())
	if err != nil {
		t.Errorf("Expected file to exist after saving, got error: %v", err)
	}
	if fileInfo.Size() == 0 {
		t.Errorf("Expected file to have content after saving")
	}
}

func TestLoadTransactionsFromJsonFileEmptyFile(t *testing.T) {
	// Create empty JSON file
	tmpJSONFile, err := os.CreateTemp("", "test_empty_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(tmpJSONFile.Name())

	// Set up test config
	testConfig := &Config{
		StorageType:  StorageJSONFile,
		JSONFilePath: tmpJSONFile.Name(),
	}
	SetGlobalConfig(testConfig)

	// Test loading from empty file
	transactions, err := loadTransactionsFromJsonFile()
	if err != nil {
		// The function returns EOF error for empty files, which is expected
		if err.Error() != "EOF" {
			t.Errorf("Expected EOF error loading from empty file, got %v", err)
		}
	}
	if len(transactions) != 0 {
		t.Errorf("Expected empty transactions from empty file, got %v", transactions)
	}
}

func TestLoadTransactionsFromJsonFileInvalidJSON(t *testing.T) {
	// Create JSON file with invalid content
	tmpJSONFile, err := os.CreateTemp("", "test_invalid_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(tmpJSONFile.Name())

	// Write invalid JSON
	_, err = tmpJSONFile.WriteString("invalid json content")
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}
	tmpJSONFile.Close()

	// Set up test config
	testConfig := &Config{
		StorageType:  StorageJSONFile,
		JSONFilePath: tmpJSONFile.Name(),
	}
	SetGlobalConfig(testConfig)

	// Test loading from invalid JSON file
	_, err = loadTransactionsFromJsonFile()
	if err == nil {
		t.Errorf("Expected error loading from invalid JSON file")
	}
}
