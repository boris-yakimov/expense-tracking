package main

import (
	"os"
	"testing"
	"time"
)

func TestListAllTransactions(t *testing.T) {
	tmpFile := "test_list_all.json"
	originalFilePath := transactionsFilePath
	transactionsFilePath = tmpFile

	// Clean up after test
	defer func() {
		transactionsFilePath = originalFilePath
		os.Remove(tmpFile)
	}()

	// Initialize with test data
	year := time.Now().Format("2006")
	month := time.Now().Format("01")

	testTransactions := map[string]map[string]map[string][]Transaction{
		year: {
			month: {
				"expense": {
					{Id: "1", Amount: 50.00, Category: "food", Description: "groceries"},
					{Id: "2", Amount: 25.00, Category: "transport", Description: "bus"},
				},
				"income": {
					{Id: "3", Amount: 1000.00, Category: "salary", Description: "monthly salary"},
				},
			},
		},
	}

	if err := saveTransactions(testTransactions); err != nil {
		t.Fatalf("Failed to initialize test file: %v", err)
	}

	cases := []struct {
		name          string
		transactions  map[string]map[string]map[string][]Transaction
		expectedError bool
	}{
		{
			name:          "list all transactions with data",
			transactions:  testTransactions,
			expectedError: false,
		},
		{
			name: "list all transactions with empty data",
			transactions: map[string]map[string]map[string][]Transaction{
				year: {
					month: {},
				},
			},
			expectedError: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Set up test data
			if err := saveTransactions(c.transactions); err != nil {
				t.Fatalf("Failed to set up test data: %v", err)
			}

			success, err := listAllTransactions()

			if (err != nil) != c.expectedError {
				t.Errorf("listAllTransactions() error = %v; expected error = %v", err, c.expectedError)
			}

			if !c.expectedError && !success {
				t.Error("listAllTransactions() returned false success when expected true")
			}
		})
	}
}

func TestListTransactionsByMonth(t *testing.T) {
	tmpFile := "test_list_by_month.json"
	originalFilePath := transactionsFilePath
	transactionsFilePath = tmpFile

	// Clean up after test
	defer func() {
		transactionsFilePath = originalFilePath
		os.Remove(tmpFile)
	}()

	// Initialize with test data
	year := time.Now().Format("2006")
	month := time.Now().Format("01")

	testTransactions := map[string]map[string]map[string][]Transaction{
		year: {
			month: {
				"expense": {
					{Id: "1", Amount: 50.00, Category: "food", Description: "groceries"},
				},
				"income": {
					{Id: "2", Amount: 1000.00, Category: "salary", Description: "monthly salary"},
				},
			},
		},
	}

	if err := saveTransactions(testTransactions); err != nil {
		t.Fatalf("Failed to initialize test file: %v", err)
	}

	cases := []struct {
		name            string
		transactionType string
		month           string
		year            string
		expectedError   bool
	}{
		{
			name:            "valid expense listing",
			transactionType: "expense",
			month:           month,
			year:            year,
			expectedError:   false,
		},
		{
			name:            "valid income listing",
			transactionType: "income",
			month:           month,
			year:            year,
			expectedError:   false,
		},
		{
			name:            "invalid transaction type",
			transactionType: "invalidtype",
			month:           month,
			year:            year,
			expectedError:   true,
		},
		{
			name:            "non-existent month",
			transactionType: "expense",
			month:           "13", // invalid month
			year:            year,
			expectedError:   false, // This should not error, just show no transactions
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			success, err := listTransactionsByMonth(c.transactionType, c.month, c.year)

			if (err != nil) != c.expectedError {
				t.Errorf("listTransactionsByMonth(%q, %q, %q) error = %v; expected error = %v",
					c.transactionType, c.month, c.year, err, c.expectedError)
			}

			if !c.expectedError && !success {
				t.Errorf("listTransactionsByMonth(%q, %q, %q) returned false success when expected true",
					c.transactionType, c.month, c.year)
			}
		})
	}
}

func TestVisualizeTransactions(t *testing.T) {
	tmpFile := "test_visualize.json"
	originalFilePath := transactionsFilePath
	transactionsFilePath = tmpFile

	// Clean up after test
	defer func() {
		transactionsFilePath = originalFilePath
		os.Remove(tmpFile)
	}()

	// Initialize with test data
	year := time.Now().Format("2006")
	month := time.Now().Format("01")

	testTransactions := map[string]map[string]map[string][]Transaction{
		year: {
			month: {
				"expense": {
					{Id: "1", Amount: 50.00, Category: "food", Description: "groceries"},
				},
				"income": {
					{Id: "2", Amount: 1000.00, Category: "salary", Description: "monthly salary"},
				},
			},
		},
	}

	if err := saveTransactions(testTransactions); err != nil {
		t.Fatalf("Failed to initialize test file: %v", err)
	}

	cases := []struct {
		name          string
		args          []string
		expectedError bool
	}{
		{
			name:          "visualize without arguments",
			args:          []string{},
			expectedError: false,
		},
		{
			name:          "visualize with valid year",
			args:          []string{year},
			expectedError: false,
		},
		{
			name:          "visualize with valid month and year",
			args:          []string{"august", year},
			expectedError: false,
		},
		{
			name:          "visualize with invalid year",
			args:          []string{"invalid-year"},
			expectedError: true,
		},
		{
			name:          "visualize with invalid month",
			args:          []string{"invalid-month", year},
			expectedError: true,
		},
		{
			name:          "visualize with too many arguments",
			args:          []string{"arg1", "arg2", "arg3"},
			expectedError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			success, err := visualizeTransactions(c.args)

			if (err != nil) != c.expectedError {
				t.Errorf("visualizeTransactions(%v) error = %v; expected error = %v",
					c.args, err, c.expectedError)
			}

			if !c.expectedError && !success {
				t.Errorf("visualizeTransactions(%v) returned false success when expected true", c.args)
			}
		})
	}
}

func TestShowAllowedCategories(t *testing.T) {
	cases := []struct {
		name            string
		transactionType string
		expectedError   bool
	}{
		{
			name:            "show expense categories",
			transactionType: "expense",
			expectedError:   false,
		},
		{
			name:            "show income categories",
			transactionType: "income",
			expectedError:   false,
		},
		{
			name:            "show investment categories",
			transactionType: "investment",
			expectedError:   false,
		},
		{
			name:            "invalid transaction type",
			transactionType: "invalidtype",
			expectedError:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := showAllowedCategories(c.transactionType)

			if (err != nil) != c.expectedError {
				t.Errorf("showAllowedCategories(%q) error = %v; expected error = %v",
					c.transactionType, err, c.expectedError)
			}
		})
	}
}
