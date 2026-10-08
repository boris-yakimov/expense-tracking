package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/mattn/go-sqlite3"
)

var db *sql.DB

// opens the transactions database purely in memory, the decrypted data never touches the disk
// data is the content of a decrypted SQLite database file, nil or empty starts with an empty database
func initDb(data []byte) error {
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return fmt.Errorf("unable to initialize db connection, err: %w", err)
	}

	// every connection to ":memory:" is its own separate database, so the pool must keep exactly one connection open forever
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	db.SetConnMaxIdleTime(0)

	if err = db.Ping(); err != nil {
		closeDb()
		return fmt.Errorf("unable to open db connection, err: %w", err)
	}

	// keep temporary tables and indices (sorting, etc) in memory too, instead of temp files on disk
	if _, err = db.Exec("PRAGMA temp_store = MEMORY"); err != nil {
		closeDb()
		return fmt.Errorf("unable to set in-memory temp store, err: %w", err)
	}

	if len(data) > 0 {
		if err = restoreDb(data); err != nil {
			closeDb()
			return fmt.Errorf("unable to load transactions database, err: %w", err)
		}
	}

	prepTransactionTable := `
		CREATE TABLE IF NOT EXISTS transactions (
			id				  TEXT PRIMARY KEY,
			amount 			NUMERIC(12, 2) NOT NULL,
			type 				TEXT NOT NULL CHECK (type IN ('income', 'expense', 'investment')),
			category 		TEXT NOT NULL,
			description TEXT,
			year 				INTEGER NOT NULL,
			month 			TEXT NOT NULL CHECK (
				month IN (
            'january','february','march','april','may','june',
            'july','august','september','october','november','december'
				)
			)
		);
	`

	_, err = db.Exec(prepTransactionTable)
	if err != nil {
		return fmt.Errorf("prep transactions db table err: %w", err)
	}

	return nil
}

func closeDb() {
	if db != nil {
		db.Close()
		db = nil
	}
}

// runs fn with the raw sqlite connection behind the single pooled in-memory connection
func withRawConn(sqlDb *sql.DB, fn func(*sqlite3.SQLiteConn) error) error {
	conn, err := sqlDb.Conn(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close() // returns the connection to the pool, does not close the in-memory db

	return conn.Raw(func(driverConn any) error {
		sqliteConn, ok := driverConn.(*sqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("unexpected driver connection type %T", driverConn)
		}
		return fn(sqliteConn)
	})
}

// copies a serialized SQLite database into the in-memory db
// a deserialized database can't grow, so it is only used as the source of a backup into the regular (growable) in-memory db
func restoreDb(data []byte) error {
	src, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}
	defer src.Close()
	src.SetMaxOpenConns(1)

	return withRawConn(src, func(srcConn *sqlite3.SQLiteConn) error {
		if err := srcConn.Deserialize(data, "main"); err != nil {
			return err
		}
		return withRawConn(db, func(destConn *sqlite3.SQLiteConn) error {
			backup, err := destConn.Backup("main", srcConn, "main")
			if err != nil {
				return err
			}
			if _, err := backup.Step(-1); err != nil {
				backup.Finish()
				return err
			}
			return backup.Finish()
		})
	})
}

// returns the in-memory db as the bytes of a regular SQLite database file
func serializeDb() ([]byte, error) {
	if db == nil {
		return nil, fmt.Errorf("database is not open")
	}
	var data []byte
	err := withRawConn(db, func(conn *sqlite3.SQLiteConn) error {
		var err error
		data, err = conn.Serialize("main")
		return err
	})
	return data, err
}

// guards against a save and a shutdown writing the encrypted file at the same time
var persistMu sync.Mutex

// writes the current in-memory db to the encrypted file on disk
func persistDb() error {
	persistMu.Lock()
	defer persistMu.Unlock()

	data, err := serializeDb()
	if err != nil {
		return fmt.Errorf("failed to serialize database: %w", err)
	}
	defer clear(data) // best effort, don't leave the plaintext copy lying around in memory

	return encryptDatabase(data)
}

// decrypts the transactions (with the password already set in memory) and opens them as an in-memory db
// on the very first run (no encrypted file yet) an empty db is created and encrypted right away
// returns ErrWrongPassword when the password doesn't match the encrypted file
func openTransactions() error {
	data, err := decryptDatabase()
	if err != nil {
		return err
	}
	defer clear(data)

	firstRun := data == nil

	// older versions kept a plaintext db file next to the encrypted one while running; if there is no encrypted
	// file yet, such a file is the only copy of the data, so import it instead of starting empty
	if firstRun && globalConfig.UnencryptedDbFile != "" {
		legacy, err := os.ReadFile(globalConfig.UnencryptedDbFile)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to read plaintext database %s: %w", globalConfig.UnencryptedDbFile, err)
		}
		data = legacy
	}

	if err := initDb(data); err != nil {
		return err
	}

	// encrypt right away, so the password is locked in and any imported plaintext data is safe before deleting it
	if firstRun {
		if err := persistDb(); err != nil {
			closeDb()
			return fmt.Errorf("failed to encrypt database: %w", err)
		}
	}

	removeLegacyPlaintextDb()
	return nil
}

// deletes a plaintext db file left behind by older versions (or by a crash of older versions)
// the encrypted file is always the source of truth, older versions also overwrote this file on every login
func removeLegacyPlaintextDb() {
	if globalConfig == nil || globalConfig.UnencryptedDbFile == "" {
		return
	}
	if err := os.Remove(globalConfig.UnencryptedDbFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("warning: failed to remove plaintext database %s: %s\n", globalConfig.UnencryptedDbFile, err)
	}
	// sqlite side files of the old plaintext db
	for _, suffix := range []string{"-journal", "-wal", "-shm"} {
		os.Remove(globalConfig.UnencryptedDbFile + suffix)
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
			id, txType, category, description, month string
			amount                                   float64
			year                                     int
		)

		if err := rows.Scan(&id, &amount, &txType, &category, &description, &year, &month); err != nil {
			return nil, fmt.Errorf("db scan failed during load transactions: %w", err)
		}

		y := fmt.Sprintf("%d", year)

		if _, ok := transactions[y]; !ok {
			transactions[y] = make(map[string]map[string][]Transaction)
		}
		if _, ok := transactions[y][month]; !ok {
			transactions[y][month] = make(map[string][]Transaction)
		}

		transactions[y][month][txType] = append(transactions[y][month][txType], Transaction{
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

			for txType, list := range types {

				for _, tr := range list {
					_, err = sqlStatement.Exec(
						tr.Id,
						tr.Amount,
						txType,
						tr.Category,
						tr.Description,
						y,     // integer, e.g. 2025
						month, // string, e.g. August
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
