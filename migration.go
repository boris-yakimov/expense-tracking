package main

import (
	"fmt"
	"strconv"
)

// migrate data from json file to sqlite
func migrateJsonToDb() error {
	history, err := loadTransactionsFromJsonFile()
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
		y, err := strconv.Atoi(year)
		if err != nil {
			return fmt.Errorf("invalid year key %q: %w", year, err)
		}

		for month, types := range months {
			m, ok := monthOrder[month]
			if !ok {
				return fmt.Errorf("invalid month key %s: %w", month, err)
			}

			for txType, list := range types {
				for _, tr := range list {
					_, err = stmt.Exec(
						tr.Id,
						tr.Amount,
						txType,
						tr.Category,
						tr.Description,
						y, // integer year
						m, // integer month
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
