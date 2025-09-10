package main

import (
	"testing"
	"time"
)

func TestListTransactionsByMonth(t *testing.T) {
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

			// Initialize with test data
			year := time.Now().Format("2006")
			month := time.Now().Format("january")

			testTransactions := TransactionHistory{
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

			if err := saveTransactionsToTestStorage(testTransactions); err != nil {
				t.Fatalf("Failed to initialize test storage: %v", err)
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
			_, err := listOfAllowedCategories(c.transactionType)

			if (err != nil) != c.expectedError {
				t.Errorf("showAllowedCategories(%q) error = %v; expected error = %v",
					c.transactionType, err, c.expectedError)
			}
		})
	}
}
