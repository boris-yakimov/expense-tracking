package main

import (
	"testing"
	"time"
)

func TestCalculateMonthPnL(t *testing.T) {
	setupTestDb(t)

	// Get current month and year for testing
	year := time.Now().Format("2006")
	month := time.Now().Format("01")

	cases := []struct {
		name            string
		transactions    TransactionHistory
		month           string
		year            string
		expectedAmount  float64
		expectedPercent float64
		expectedError   bool
	}{
		{
			name: "positive P&L with income and expenses",
			transactions: TransactionHistory{
				year: {
					month: {
						"income": {
							{Id: "1", Amount: 1000.00, Category: "salary", Description: "salary"},
						},
						"expense": {
							{Id: "2", Amount: 300.00, Category: "food", Description: "groceries"},
							{Id: "3", Amount: 200.00, Category: "transport", Description: "bus"},
						},
					},
				},
			},
			month:           month,
			year:            year,
			expectedAmount:  500.00, // 1000 - 300 - 200
			expectedPercent: 50.0,   // (1000-500)/1000 * 100
			expectedError:   false,
		},
		{
			name: "negative P&L with high expenses",
			transactions: TransactionHistory{
				year: {
					month: {
						"income": {
							{Id: "1", Amount: 500.00, Category: "salary", Description: "salary"},
						},
						"expense": {
							{Id: "2", Amount: 800.00, Category: "food", Description: "groceries"},
						},
					},
				},
			},
			month:           month,
			year:            year,
			expectedAmount:  -300.00, // 500 - 800
			expectedPercent: -60.0,   // (500-800)/500 * 100
			expectedError:   false,
		},
		{
			name: "zero income with expenses",
			transactions: TransactionHistory{
				year: {
					month: {
						"expense": {
							{Id: "1", Amount: 100.00, Category: "food", Description: "groceries"},
						},
					},
				},
			},
			month:           month,
			year:            year,
			expectedAmount:  -100.00, // 0 - 100
			expectedPercent: 0.0,     // division by zero case
			expectedError:   false,
		},
		{
			name: "only income no expenses",
			transactions: TransactionHistory{
				year: {
					month: {
						"income": {
							{Id: "1", Amount: 1000.00, Category: "salary", Description: "salary"},
						},
					},
				},
			},
			month:           month,
			year:            year,
			expectedAmount:  1000.00, // 1000 - 0
			expectedPercent: 100.0,   // (1000-0)/1000 * 100
			expectedError:   false,
		},
		{
			name: "no transactions",
			transactions: TransactionHistory{
				year: {
					month: {},
				},
			},
			month:           month,
			year:            year,
			expectedAmount:  0.0,
			expectedPercent: 0.0,
			expectedError:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Set up test data
			if err := saveTransactionsToTestDb(c.transactions); err != nil {
				t.Fatalf("Failed to set up test data: %v", err)
			}

			result, err := calculateMonthPnL(c.month, c.year)

			if (err != nil) != c.expectedError {
				t.Errorf("calculateMonthPnL(%q, %q) error = %v; expected error = %v",
					c.month, c.year, err, c.expectedError)
			}

			if !c.expectedError {
				if result.Amount != c.expectedAmount {
					t.Errorf("calculateMonthPnL(%q, %q) amount = %f; expected amount = %f",
						c.month, c.year, result.Amount, c.expectedAmount)
				}
				if result.Percent != c.expectedPercent {
					t.Errorf("calculateMonthPnL(%q, %q) percent = %f; expected percent = %f",
						c.month, c.year, result.Percent, c.expectedPercent)
				}
			}
		})
	}
}

func TestCalculateYearPnL(t *testing.T) {
	setupTestDb(t)

	year := time.Now().Format("2006")

	cases := []struct {
		name            string
		transactions    TransactionHistory
		year            string
		expectedAmount  float64
		expectedPercent float64
		expectedError   bool
	}{
		{
			name: "year with multiple months",
			transactions: TransactionHistory{
				year: {
					"01": {
						"income": {
							{Id: "1", Amount: 1000.00, Category: "salary", Description: "jan salary"},
						},
						"expense": {
							{Id: "2", Amount: 300.00, Category: "food", Description: "jan food"},
						},
					},
					"02": {
						"income": {
							{Id: "3", Amount: 1000.00, Category: "salary", Description: "feb salary"},
						},
						"expense": {
							{Id: "4", Amount: 400.00, Category: "food", Description: "feb food"},
						},
					},
				},
			},
			year:            year,
			expectedAmount:  1300.00, // (1000-300) + (1000-400) = 700 + 600
			expectedPercent: 65.0,    // (1300)/2000 * 100
			expectedError:   false,
		},
		{
			name: "year with only expenses",
			transactions: TransactionHistory{
				year: {
					"01": {
						"expense": {
							{Id: "1", Amount: 500.00, Category: "food", Description: "food"},
						},
					},
				},
			},
			year:            year,
			expectedAmount:  -500.00,
			expectedPercent: 0.0, // division by zero case
			expectedError:   false,
		},
		{
			name: "empty year",
			transactions: TransactionHistory{
				year: {},
			},
			year:            year,
			expectedAmount:  0.0,
			expectedPercent: 0.0,
			expectedError:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Set up test data
			if err := saveTransactionsToTestDb(c.transactions); err != nil {
				t.Fatalf("Failed to set up test data: %v", err)
			}

			result, err := calculateYearPnL(c.year)

			if (err != nil) != c.expectedError {
				t.Errorf("calculateYearPnL(%q) error = %v; expected error = %v",
					c.year, err, c.expectedError)
			}

			if !c.expectedError {
				if result.Amount != c.expectedAmount {
					t.Errorf("calculateYearPnL(%q) amount = %f; expected amount = %f",
						c.year, result.Amount, c.expectedAmount)
				}
				if result.Percent != c.expectedPercent {
					t.Errorf("calculateYearPnL(%q) percent = %f; expected percent = %f",
						c.year, result.Percent, c.expectedPercent)
				}
			}
		})
	}
}
