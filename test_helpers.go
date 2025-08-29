package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var testDb *sql.DB

// setupTestDb creates a temporary test database with a transactions table
func setupTestDb(t *testing.T) {
	// Create temporary database file
	tmpDbFile, err := os.CreateTemp("", "test_transactions_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}

	// Store the temp file path for cleanup
	testDbFilePath := tmpDbFile.Name()
	tmpDbFile.Close()

	// Initialize test database
	testDb, err = sql.Open("sqlite3", testDbFilePath)
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}

	// Create transactions table (same as main table for compatibility)
	testSchema := `
		CREATE TABLE transactions (
			id				  TEXT PRIMARY KEY,
			amount 			NUMERIC(12, 2) NOT NULL,
			type 				TEXT NOT NULL CHECK (type IN ('income', 'expense', 'investment')),
			category 		TEXT NOT NULL,
			description TEXT,
			year 				INTEGER NOT NULL,
			month 			INTEGER NOT NULL CHECK (month BETWEEN 1 and 12)
		);
	`

	_, err = testDb.Exec(testSchema)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Replace the global db with testDb for the duration of the test
	originalDb := db
	db = testDb

	// Clean up function
	t.Cleanup(func() {
		db = originalDb
		testDb.Close()
		os.Remove(testDbFilePath)
	})
}

// loadTransactionsFromTestDb loads transactions from the transactions table (in test db)
func loadTransactionsFromTestDb() (TransactionHistory, error) {
	rows, err := db.Query(`
			SELECT id, amount, type, category, description, year, month
			FROM transactions
		`)
	if err != nil {
		return nil, fmt.Errorf("failed to execute load transactions sql query: %w", err)
	}
	defer rows.Close()

	transactions := make(TransactionHistory)

	for rows.Next() {
		var (
			id, txType, category, description string
			amount                            float64
			year, month                       int
		)

		if err := rows.Scan(&id, &amount, &txType, &category, &description, &year, &month); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		y := fmt.Sprintf("%d", year)
		m := fmt.Sprintf("%02d", month)

		if _, ok := transactions[y]; !ok {
			transactions[y] = make(map[string]map[string][]Transaction)
		}
		if _, ok := transactions[y][m]; !ok {
			transactions[y][m] = make(map[string][]Transaction)
		}

		transactions[y][m][txType] = append(transactions[y][m][txType], Transaction{
			Id:          id,
			Amount:      amount,
			Category:    category,
			Description: description,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed during transaction loading: %w", err)
	}

	return transactions, nil
}

// saveTransactionsToTestDb saves transactions to the transactions table (in test db)
func saveTransactionsToTestDb(transactions TransactionHistory) error {
	// Clear existing data first
	_, err := db.Exec("DELETE FROM transactions")
	if err != nil {
		return fmt.Errorf("failed to clear test transactions: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin save transaction failed: %w", err)
	}

	sqlStatement, err := tx.Prepare(`
			INSERT INTO transactions
			(id, amount, type, category, description, year, month)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare insert during save transaction failed: %w", err)
	}
	defer sqlStatement.Close()

	for year, months := range transactions {
		y, err := strconv.Atoi(year)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("invalid year key %q: %w", year, err)
		}

		for month, types := range months {
			m, err := strconv.Atoi(month)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("invalid month key %q: %w", month, err)
			}

			for txType, list := range types {
				for _, tr := range list {
					_, err = sqlStatement.Exec(
						tr.Id,
						tr.Amount,
						txType,
						tr.Category,
						tr.Description,
						y,
						m,
					)
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("insert failed for transaction %s: %w", tr.Id, err)
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}
