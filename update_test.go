package main

import (
	"strconv"
	"testing"
	"time"
)

func TestHandleUpdateTransaction(t *testing.T) {
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

			// Initialize with some test data
			year := time.Now().Format("2006")
			month := time.Now().Format("january")

			testTransactions := TransactionHistory{
				year: {
					month: {
						"expense": {
							{Id: "12345678", Amount: 50.00, Category: "food", Description: "original expense"},
							{Id: "87654321", Amount: 25.00, Category: "transport", Description: "original transport"},
						},
						"income": {
							{Id: "11111111", Amount: 1000.00, Category: "salary", Description: "original income"},
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
				transactionId   string
				amount          string
				category        string
				description     string
				expectedError   bool
			}{
				{
					name:            "valid update expense amount",
					transactionType: "expense",
					transactionId:   "12345678",
					amount:          "75.50",
					category:        "food",
					description:     "updated expense",
					expectedError:   false,
				},
				{
					name:            "valid update expense category",
					transactionType: "expense",
					transactionId:   "87654321",
					amount:          "25.00",
					category:        "entertainment",
					description:     "updated transport to entertainment",
					expectedError:   false,
				},
				{
					name:            "valid update income",
					transactionType: "income",
					transactionId:   "11111111",
					amount:          "1200.00",
					category:        "salary",
					description:     "updated income",
					expectedError:   false,
				},
				{
					name:            "invalid transaction ID length",
					transactionType: "expense",
					transactionId:   "123", // too short
					amount:          "50.00",
					category:        "food",
					description:     "test",
					expectedError:   true,
				},
				{
					name:            "invalid amount format",
					transactionType: "expense",
					transactionId:   "12345678",
					amount:          "invalid-amount",
					category:        "food",
					description:     "test",
					expectedError:   true,
				},
				{
					name:            "invalid category",
					transactionType: "expense",
					transactionId:   "12345678",
					amount:          "50.00",
					category:        "invalidcategory",
					description:     "test",
					expectedError:   true,
				},
				{
					name:            "invalid description format",
					transactionType: "expense",
					transactionId:   "12345678",
					amount:          "50.00",
					category:        "food",
					description:     "description_with_underscores",
					expectedError:   true,
				},
				{
					name:            "invalid transaction type",
					transactionType: "invalidtype",
					transactionId:   "12345678",
					amount:          "50.00",
					category:        "food",
					description:     "test",
					expectedError:   true,
				},
			}

			for _, c := range cases {
				t.Run(c.name, func(t *testing.T) {
					updateReq := UpdateTransactionRequest{
						Type:        c.transactionType,
						Id:          c.transactionId,
						Amount:      c.amount,
						Category:    c.category,
						Description: c.description,
					}

					err := handleUpdateTransaction(updateReq)

					if (err != nil) != c.expectedError {
						t.Errorf("handleUpdateTransaction(%q, %q, %q, %q, %q) error = %v; expected error = %v",
							c.transactionType, c.transactionId, c.amount, c.category, c.description, err, c.expectedError)
					}

					// If no error expected, verify transaction was updated
					if !c.expectedError {
						transactions, loadErr := loadTransactionsFromTestStorage()
						if loadErr != nil {
							t.Errorf("Failed to load transactions after successful update: %v", loadErr)
							return
						}

						// Find and verify the updated transaction
						found := false
						if yearTx, yearOk := transactions[year]; yearOk {
							if monthTx, monthOk := yearTx[month]; monthOk {
								if typeTx, typeOk := monthTx[c.transactionType]; typeOk {
									for _, tx := range typeTx {
										if tx.Id == c.transactionId {
											found = true
											if tx.Description != c.description {
												t.Errorf("Transaction description not updated: expected %q, got %q", c.description, tx.Description)
											}
											if tx.Category != c.category {
												t.Errorf("Transaction category not updated: expected %q, got %q", c.category, tx.Category)
											}
											// Note: Amount comparison with float requires care for precision
											expectedAmount, _ := strconv.ParseFloat(c.amount, 64)
											if tx.Amount != expectedAmount {
												t.Errorf("Transaction amount not updated: expected %f, got %f", expectedAmount, tx.Amount)
											}
											break
										}
									}
								}
							}
						}
						if !found {
							t.Errorf("Transaction with ID %q not found after update", c.transactionId)
						}
					}
				})
			}
		})
	}
}
