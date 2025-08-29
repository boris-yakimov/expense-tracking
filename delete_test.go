package main

import (
	"testing"
	"time"
)

func TestHandleDeleteTransaction(t *testing.T) {
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
			month := time.Now().Format("01")

			testTransactions := TransactionHistory{
				year: {
					month: {
						"expense": {
							{Id: "12345678", Amount: 50.00, Category: "food", Description: "test expense"},
							{Id: "87654321", Amount: 25.00, Category: "transport", Description: "test transport"},
						},
						"income": {
							{Id: "11111111", Amount: 1000.00, Category: "salary", Description: "test income"},
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
				expectedError   bool
			}{
				{
					name:            "valid delete expense",
					transactionType: "expense",
					transactionId:   "12345678",
					expectedError:   false,
				},
				{
					name:            "valid delete income",
					transactionType: "income",
					transactionId:   "11111111",
					expectedError:   false,
				},
				{
					name:            "invalid transaction ID length",
					transactionType: "expense",
					transactionId:   "123", // too short
					expectedError:   true,
				},
				{
					name:            "transaction ID not found",
					transactionType: "expense",
					transactionId:   "99999999", // doesn't exist
					expectedError:   true,
				},
				{
					name:            "invalid transaction type",
					transactionType: "invalidtype",
					transactionId:   "12345678",
					expectedError:   true,
				},
			}

			for _, c := range cases {
				t.Run(c.name, func(t *testing.T) {
					err := handleDeleteTransaction(c.transactionType, c.transactionId)

					if (err != nil) != c.expectedError {
						t.Errorf("handleDeleteTransaction(%q, %q) error = %v; expected error = %v",
							c.transactionType, c.transactionId, err, c.expectedError)
					}

					// If no error expected, verify transaction was deleted
					if !c.expectedError {
						// Reload transactions from storage to verify deletion
						transactions, loadErr := loadTransactionsFromTestStorage()
						if loadErr != nil {
							t.Errorf("Failed to load transactions after successful delete: %v", loadErr)
							return
						}

						// Check that the transaction is no longer in the list
						found := false
						if yearTx, yearOk := transactions[year]; yearOk {
							if monthTx, monthOk := yearTx[month]; monthOk {
								if typeTx, typeOk := monthTx[c.transactionType]; typeOk {
									for _, tx := range typeTx {
										if tx.Id == c.transactionId {
											found = true
											break
										}
									}
								}
							}
						}
						if found {
							t.Errorf("Transaction with ID %q was not deleted", c.transactionId)
						}
					}
				})
			}
		})
	}
}
