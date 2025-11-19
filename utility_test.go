package main

import (
	"slices"
	"testing"
)

func TestValidDescriptionInputFormat(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{"correct description", true},
		{"Valid description 123", true},
		{"Another-description, with commas", true},
		{"dash-separated-description", true},
		{"description with 'single quotes'", true},
		{"contains_underscore", false},
		{"contains@symbol", false},
		{"contains/slash", false},
		{"", false},   // empty string is not valid based on the regex
		{"   ", true}, // spaces only, allowed by regex
		{"strings that is too long for what might be expected as a description, but is still valid also includes - and Capital letter", true},
	}

	for _, c := range cases {
		validFormat := validDescriptionInputFormat(c.input)
		if validFormat != c.expected {
			t.Errorf("validDescriptionInputFormat(%q) = %v; expected %v", c.input, validFormat, c.expected)
		}
	}
}

func TestNormalizeTransactionType(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"expense", "expense", false},
		{"expenses", "expense", false},
		{"Expense", "expense", false},
		{"Expenses", "expense", false},
		{"income", "income", false},
		{"Income", "income", false},
		{"investment", "investment", false},
		{"investments", "investment", false},
		{"Investment", "investment", false},
		{"Investments", "investment", false},
		{"invalid", "", true},
		{"", "", true},
		{"random", "", true},
		{"EXPENSE", "", true},    // uppercase not supported
		{"INCOME", "", true},     // uppercase not supported
		{"INVESTMENT", "", true}, // uppercase not supported
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			result, err := normalizeTransactionType(c.input)

			if (err != nil) != c.hasError {
				t.Errorf("normalizeTransactionType(%q) error = %v; expected error = %v", c.input, err, c.hasError)
			}

			if !c.hasError && result != c.expected {
				t.Errorf("normalizeTransactionType(%q) = %q; expected %q", c.input, result, c.expected)
			}
		})
	}
}

func TestGenerateTransactionId(t *testing.T) {
	// Test that IDs are generated and have correct length
	for range 10 {
		id, err := generateTransactionId()
		if err != nil {
			t.Errorf("generateTransactionId() returned error: %v", err)
		}
		if len(id) != TransactionIDLength {
			t.Errorf("generateTransactionId() returned ID of length %d; expected %v", len(id), TransactionIDLength)
		}
		// Check that ID contains only alphanumeric characters
		for _, char := range id {
			if !((char >= '0' && char <= '9') || (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')) {
				t.Errorf("generateTransactionId() returned ID with invalid character: %c", char)
			}
		}
	}
}

func TestCapitalize(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"test", "Test"},
		{"a", "A"},
		{"", ""},
		{"already Capitalized", "Already capitalized"}, // function converts to lowercase after first char
		{"MIXED case", "Mixed case"},                   // function converts to lowercase after first char
		{"123", "123"},
		{"hello world", "Hello world"},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			result := capitalize(c.input)
			if result != c.expected {
				t.Errorf("capitalize(%q) = %q; expected %q", c.input, result, c.expected)
			}
		})
	}
}

func TestListOfAllowedCategories(t *testing.T) {
	cases := []struct {
		transactionType string
		expectedError   bool
	}{
		{"expense", false},
		{"income", false},
		{"investment", false},
		{"invalidtype", true},
		{"", true},
	}

	for _, c := range cases {
		t.Run(c.transactionType, func(t *testing.T) {
			categories, err := listOfAllowedCategories(c.transactionType)

			if (err != nil) != c.expectedError {
				t.Errorf("listOfAllowedCategories(%q) error = %v; expected error = %v",
					c.transactionType, err, c.expectedError)
			}

			if !c.expectedError && len(categories) == 0 {
				t.Errorf("Expected non-empty categories for valid transaction type %q", c.transactionType)
			}
		})
	}
}

func TestListOfAllowedTransactionTypes(t *testing.T) {
	transactionTypes, err := listOfAllowedTransactionTypes()
	if err != nil {
		t.Errorf("Expected no error getting transaction types, got %v", err)
	}
	if len(transactionTypes) == 0 {
		t.Errorf("Expected non-empty transaction types")
	}

	// Check that expected transaction types are present
	expectedTypes := []string{"expense", "income", "investment"}
	for _, expectedType := range expectedTypes {
		if !slices.Contains(transactionTypes, expectedType) {
			t.Errorf("Expected transaction type %q to be in the list", expectedType)
		}
	}
}

func TestGetListOfDetailedTransactions(t *testing.T) {
	// This test requires setting up test storage
	testCases := []struct {
		name        string
		storageType StorageType
	}{
		{"SQLite", StorageSQLite},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestStorage(t, tc.storageType)

			// Test with empty storage
			transactions, err := getListOfDetailedTransactions()
			if err != nil {
				if tc.storageType == StorageSQLite {
					t.Errorf("Expected no error with empty SQLite storage, got %v", err)
				}
			}
			if len(transactions) != 0 {
				t.Errorf("Expected empty transactions from empty storage, got %v", transactions)
			}

			// Test with some data
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

			if err := saveTransactionsToTestStorage(testTransactions); err != nil {
				t.Fatalf("Failed to save test data: %v", err)
			}

			transactions, err = getListOfDetailedTransactions()
			if err != nil {
				t.Errorf("Expected no error with test data, got %v", err)
			}
			if len(transactions) == 0 {
				t.Errorf("Expected transactions from test data")
			}

			// Verify transaction details format
			if len(transactions) > 0 {
				transaction := transactions[0]
				if transaction == "" {
					t.Errorf("Expected non-empty transaction detail")
				}
			}
		})
	}
}

func TestGetTransactionTypeById(t *testing.T) {
	testCases := []struct {
		name        string
		storageType StorageType
	}{
		{"SQLite", StorageSQLite},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestStorage(t, tc.storageType)

			// Test with empty storage
			_, err := getTransactionTypeById("nonexistent")
			if err == nil {
				t.Errorf("Expected error for nonexistent transaction ID")
			}

			// Test with some data
			testTransactions := TransactionHistory{
				"2023": {
					"january": {
						"expense": []Transaction{
							{Id: "12345678", Amount: 10.0, Category: "food", Description: "test food"},
						},
						"income": []Transaction{
							{Id: "87654321", Amount: 1000.0, Category: "salary", Description: "test salary"},
						},
					},
				},
			}

			if err := saveTransactionsToTestStorage(testTransactions); err != nil {
				t.Fatalf("Failed to save test data: %v", err)
			}

			// Test finding expense transaction
			txType, err := getTransactionTypeById("12345678")
			if err != nil {
				t.Errorf("Expected no error finding expense transaction, got %v", err)
			}
			if txType != "expense" {
				t.Errorf("Expected transaction type 'expense', got %q", txType)
			}

			// Test finding income transaction
			txType, err = getTransactionTypeById("87654321")
			if err != nil {
				t.Errorf("Expected no error finding income transaction, got %v", err)
			}
			if txType != "income" {
				t.Errorf("Expected transaction type 'income', got %q", txType)
			}

			// Test finding nonexistent transaction
			_, err = getTransactionTypeById("99999999")
			if err == nil {
				t.Errorf("Expected error for nonexistent transaction ID")
			}
		})
	}
}

func TestEnforceCharLimit(t *testing.T) {
	cases := []struct {
		text     string
		lastChar rune
		expected bool
	}{
		{"short", 'a', true},
		{"", 'a', true},
		{string(make([]byte, DescriptionMaxCharLength)), 'a', true},
		{string(make([]byte, DescriptionMaxCharLength+1)), 'a', false},
		{"exactly forty chars long text here!", 'a', true},
		{"this text is way too long and exceeds the maximum character limit allowed and should definitely be invalid because it goes way beyond the normal limits of what should be allowed for a simple description field in any reasonable application", 'a', false},
	}

	for _, c := range cases {
		t.Run(c.text, func(t *testing.T) {
			result := enforceCharLimit(c.text, c.lastChar)
			if result != c.expected {
				t.Errorf("enforceCharLimit(%q, %c) = %v; expected %v", c.text, c.lastChar, result, c.expected)
			}
		})
	}
}

func TestGetMonthsWithTransactions(t *testing.T) {
	testCases := []struct {
		name        string
		storageType StorageType
	}{
		{"SQLite", StorageSQLite},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestStorage(t, tc.storageType)

			// Test with empty storage
			months, err := getMonthsWithTransactions()
			if err != nil {
				if tc.storageType == StorageSQLite {
					t.Errorf("Expected no error with empty SQLite storage, got %v", err)
				}
			}
			if len(months) != 0 {
				t.Errorf("Expected empty months from empty storage, got %v", months)
			}

			// Test with some data
			testTransactions := TransactionHistory{
				"2023": {
					"january": {
						"expense": []Transaction{
							{Id: "1", Amount: 10.0, Category: "food", Description: "test food"},
						},
					},
					"february": {
						"income": []Transaction{
							{Id: "2", Amount: 1000.0, Category: "salary", Description: "test salary"},
						},
					},
				},
				"2024": {
					"march": {
						"expense": []Transaction{
							{Id: "3", Amount: 20.0, Category: "transport", Description: "test transport"},
						},
					},
				},
			}

			if err := saveTransactionsToTestStorage(testTransactions); err != nil {
				t.Fatalf("Failed to save test data: %v", err)
			}

			months, err = getMonthsWithTransactions()
			if err != nil {
				t.Errorf("Expected no error with test data, got %v", err)
			}
			if len(months) == 0 {
				t.Errorf("Expected months from test data")
			}

			// Verify expected months are present
			expectedMonths := []string{"january 2023", "february 2023", "march 2024"}
			for _, expectedMonth := range expectedMonths {
				if !slices.Contains(months, expectedMonth) {
					t.Errorf("Expected month %q to be in months list", expectedMonth)
				}
			}
		})
	}
}

func TestDetermineLatestMonthAndYear(t *testing.T) {
	testCases := []struct {
		name        string
		storageType StorageType
	}{
		{"SQLite", StorageSQLite},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestStorage(t, tc.storageType)

			// Test with empty storage
			month, year, err := determineLatestMonthAndYear()
			if err != nil {
				if tc.storageType == StorageSQLite {
					t.Errorf("Expected no error with empty SQLite storage, got %v", err)
				}
			}
			if month != "" || year != "" {
				t.Errorf("Expected empty month and year from empty storage, got %q, %q", month, year)
			}

			// Test with some data
			testTransactions := TransactionHistory{
				"2023": {
					"january": {
						"expense": []Transaction{
							{Id: "1", Amount: 10.0, Category: "food", Description: "test food"},
						},
					},
					"june": {
						"income": []Transaction{
							{Id: "2", Amount: 1000.0, Category: "salary", Description: "test salary"},
						},
					},
				},
				"2024": {
					"march": {
						"expense": []Transaction{
							{Id: "3", Amount: 20.0, Category: "transport", Description: "test transport"},
						},
					},
				},
			}

			if err := saveTransactionsToTestStorage(testTransactions); err != nil {
				t.Fatalf("Failed to save test data: %v", err)
			}

			month, year, err = determineLatestMonthAndYear()
			if err != nil {
				t.Errorf("Expected no error with test data, got %v", err)
			}
			if month != "march" || year != "2024" {
				t.Errorf("Expected latest month 'march' and year '2024', got %q, %q", month, year)
			}
		})
	}
}
