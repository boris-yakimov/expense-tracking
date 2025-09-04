package main

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDb(dbFilePath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return fmt.Errorf("unable to initialize db connection, err: %w", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("unable to open db connection, err: %w", err)
	}

	prepTransactionTable := `
		CREATE TABLE IF NOT EXISTS transactions (
			id				  TEXT PRIMARY KEY,
			amount 			NUMERIC(12, 2) NOT NULL,
			type 				TEXT NOT NULL CHECK (type IN ('income', 'expense', 'investment')),
			category 		TEXT NOT NULL,
			description TEXT,
			year 				INTEGER NOT NULL,
			month 			INTEGER NOT NULL CHECK (month BETWEEN 1 and 12)
		);
	`

	_, err = db.Exec(prepTransactionTable)
	if err != nil {
		return fmt.Errorf("prep transactions db table err: %w", err)
	}

	// the expecation is that we maintain only one auth password at the moment
	prepAuthTable := `
		CREATE TABLE IF NOT EXISTS authentication (
			id INT PRIMARY KEY DEFAULT 1,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	  );
	`

	_, err = db.Exec(prepAuthTable)
	if err != nil {
		return fmt.Errorf("prep auth db table err: %w", err)
	}

	return nil
}

func closeDb() {
	if db != nil {
		db.Close()
	}
}

func loadTransactionsFromDb() (TransactionHistory, error) {
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
			return nil, fmt.Errorf("db scan failed during load transactions: %w", err)
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

func saveTransactionsToDb(transactions TransactionHistory) error {
	sqlTx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin save transaction failed: %w", err)
	}

	// Clear existing data first
	_, err = sqlTx.Exec("DELETE FROM transactions")
	if err != nil {
		sqlTx.Rollback()
		return fmt.Errorf("failed to clear transactions: %w", err)
	}

	sqlStatement, err := sqlTx.Prepare(`
			INSERT INTO transactions
			(id, amount, type, category, description, year, month)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`)
	if err != nil {
		sqlTx.Rollback()
		return fmt.Errorf("prepare insert during save transaction failed: %w", err)
	}
	defer sqlStatement.Close()

	for year, months := range transactions {
		y, err := strconv.Atoi(year)
		if err != nil {
			sqlTx.Rollback()
			return fmt.Errorf("invalid year key %q: %w", year, err)
		}

		for month, types := range months {
			m, err := strconv.Atoi(month)
			if err != nil {
				sqlTx.Rollback()
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
						sqlTx.Rollback()
						return fmt.Errorf("insert failed for transaction %s: %w", tr.Id, err)
					}
				}
			}
		}
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}
