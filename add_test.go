package main

import (
	"testing"
	"time"
)

func TestHandleAddTransaction(t *testing.T) {
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

			// Initialize empty test storage
			emptyTransactions := make(TransactionHistory)
			if err := saveTransactionsToTestStorage(emptyTransactions); err != nil {
				t.Fatalf("Failed to initialize test storage: %v", err)
			}

			// Get current month and year for testing
			year := time.Now().Format("2006")
			month := time.Now().Format("january")

			cases := []struct {
				name            string
				transactionType string
				amount          string
				category        string
				description     string
				month           string
				year            string
				expectedError   bool
			}{
				{
					name:            "valid add expense",
					transactionType: "expense",
					amount:          "54.30",
					category:        "food",
					description:     "test food description",
					month:           month,
					year:            year,
					expectedError:   false,
				},
				{
					name:            "valid add large expense",
					transactionType: "expense",
					amount:          "5423.87",
					category:        "renovation",
					description:     "fence",
					month:           month,
					year:            year,
					expectedError:   false,
				},
				{
					name:            "invalid transaction type",
					transactionType: "expAnse",
					amount:          "54.30",
					category:        "food",
					description:     "test food description",
					month:           month,
					year:            year,
					expectedError:   true,
				},
				{
					name:            "invalid category",
					transactionType: "expense",
					amount:          "54.30",
					category:        "madeUpCategory",
					description:     "test food description",
					month:           month,
					year:            year,
					expectedError:   true,
				},
				{
					name:            "invalid amount format",
					transactionType: "expense",
					amount:          "invalid-amount",
					category:        "food",
					description:     "test food description",
					month:           month,
					year:            year,
					expectedError:   true,
				},
				{
					name:            "invalid description format",
					transactionType: "expense",
					amount:          "100.00",
					category:        "food",
					description:     "description_with_underscores",
					month:           month,
					year:            year,
					expectedError:   true,
				},
			}

			for _, c := range cases {
				t.Run(c.name, func(t *testing.T) {
					addReq := AddTransactionRequest{
						Type:        c.transactionType,
						Amount:      c.amount,
						Category:    c.category,
						Description: c.description,
						Month:       c.month,
						Year:        c.year,
					}
					err := handleAddTransaction(addReq)

					if (err != nil) != c.expectedError {
						t.Errorf("handleAddTransaction(%q, %q, %q, %q, %q, %q) error = %v; expected error = %v",
							c.transactionType, c.amount, c.category, c.description, c.month, c.year, err, c.expectedError)
					}

					// If no error expected, verify transaction was added
					if !c.expectedError {
						transactions, loadErr := loadTransactionsFromTestStorage()
						if loadErr != nil {
							t.Errorf("Failed to load transactions after successful add: %v", loadErr)
							return
						}

						// Check if transaction exists in the expected location
						yearTx, yearOk := transactions[c.year]
						if !yearOk {
							t.Errorf("Year %s not found in transactions", c.year)
							return
						}
						monthTx, monthOk := yearTx[c.month]
						if !monthOk {
							t.Errorf("Month %s not found in transactions for year %s", c.month, c.year)
							return
						}
						typeTx, typeOk := monthTx[c.transactionType]
						if !typeOk {
							t.Errorf("Transaction type %s not found in transactions for %s %s", c.transactionType, c.month, c.year)
							return
						}

						// Find the transaction by description (since we don't know the ID)
						found := false
						for _, tx := range typeTx {
							if tx.Description == c.description && tx.Category == c.category {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Transaction with description %q and category %q not found", c.description, c.category)
						}
					}
				})
			}
		})
	}
}
