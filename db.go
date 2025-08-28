package main

import (
	"database/sql"
	"fmt"

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

	prepTransactionSchema := `
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

	_, err = db.Exec(prepTransactionSchema)
	if err != nil {
		return fmt.Errorf("prep db schema err: %w", err)
	}

	fmt.Printf("db schema initialized successfully\n\n")
	return nil
}

func closeDb() {
	if db != nil {
		db.Close()
	}
}

// TODO: figure out how to import the data from json to sql
// TODO: fetch transactions from database and convert them to my transactions struct/nested object
// TODO: update sql connection to expect _auth_user / _auth_pass
