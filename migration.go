package main

import (
	"fmt"
)

// migrate date from json file to sqlite
func migrateJsonToDb() error {
	history, err := loadTransactions()
	if err != nil {
		return fmt.Errorf("failed to load JSON transactions: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	stmt, err := tx.Prepare(`
        INSERT OR REPLACE INTO transactions
        (id, amount, type, category, description, year, month)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare insert: %w", err)
	}
	defer stmt.Close()

	for year, months := range history {
		for month, types := range months {
			for txType, list := range types {
				for _, tr := range list {
					_, err = stmt.Exec(
						tr.Id,
						tr.Amount,
						txType,
						tr.Category,
						tr.Description,
						year,
						month,
					)
					if err != nil {
						return fmt.Errorf("failed to insert transaction %s: %w", tr.Id, err)
					}
				}
			}
		}
	}

	return err
}
