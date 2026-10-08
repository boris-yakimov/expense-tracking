package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitDb(t *testing.T) {
	// Test successful in-memory database initialization
	if err := initDb(nil); err != nil {
		t.Fatalf("Expected no error from initDb, got %v", err)
	}
	defer closeDb()

	// the transactions table must exist
	if _, err := db.Exec("SELECT count(*) FROM transactions"); err != nil {
		t.Errorf("Expected transactions table to exist, got %v", err)
	}
}

func TestInitDbWithExistingData(t *testing.T) {
	testTransactions := TransactionHistory{
		"2023": {"january": {"expense": []Transaction{{Id: "1", Amount: 10.0, Category: "food", Description: "test"}}}},
	}

	// build a serialized db with data in it
	if err := initDb(nil); err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}
	if err := saveTransactionsToDb(testTransactions); err != nil {
		t.Fatalf("Failed to save transactions: %v", err)
	}
	data, err := serializeDb()
	if err != nil {
		t.Fatalf("Failed to serialize db: %v", err)
	}
	closeDb()

	// re-open from the serialized bytes
	if err := initDb(data); err != nil {
		t.Fatalf("Expected no error from initDb with existing data, got %v", err)
	}
	defer closeDb()

	loaded, err := loadTransactionsFromDb()
	if err != nil {
		t.Fatalf("Failed to load transactions: %v", err)
	}
	if got := loaded["2023"]["january"]["expense"]; len(got) != 1 || got[0].Description != "test" {
		t.Errorf("Expected restored transaction, got %v", loaded)
	}

	// the restored db must still be writable and able to grow (a plain deserialized db can't)
	many := TransactionHistory{"2023": {"january": {"expense": nil}}}
	for i := 0; i < 2000; i++ {
		many["2023"]["january"]["expense"] = append(many["2023"]["january"]["expense"],
			Transaction{Id: fmt.Sprintf("id%06d", i), Amount: 1, Category: "food", Description: strings.Repeat("x", 100)})
	}
	if err := saveTransactionsToDb(many); err != nil {
		t.Errorf("Expected restored db to accept more data, got %v", err)
	}
}

func TestInitDbWithInvalidData(t *testing.T) {
	if err := initDb([]byte("definitely not a sqlite database, just some random bytes padding it out")); err == nil {
		closeDb()
		t.Errorf("Expected error from initDb with invalid data")
	}
	if db != nil {
		t.Errorf("Expected db to be nil after failed init")
	}
}

func TestCloseDb(t *testing.T) {
	// Test closing nil database (should not panic)
	closeDb()

	if err := initDb(nil); err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}

	// Test that closeDb doesn't panic
	closeDb()
	if db != nil {
		t.Errorf("Expected db to be nil after closeDb")
	}
}

func TestLoadTransactionsFromDb(t *testing.T) {
	if err := initDb(nil); err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}
	defer closeDb()

	// Test loading from empty database
	transactions, err := loadTransactionsFromDb()
	if err != nil {
		t.Errorf("Expected no error loading from empty db, got %v", err)
	}
	if len(transactions) != 0 {
		t.Errorf("Expected empty transactions from empty db, got %v", transactions)
	}
}

func TestSaveTransactionsToDb(t *testing.T) {
	if err := initDb(nil); err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}
	defer closeDb()

	// Test saving transactions
	testTransactions := TransactionHistory{
		"2023": {
			"january": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	if err := saveTransactionsToDb(testTransactions); err != nil {
		t.Errorf("Expected no error saving transactions, got %v", err)
	}

	// Verify transactions were saved
	loadedTransactions, err := loadTransactionsFromDb()
	if err != nil {
		t.Errorf("Expected no error loading transactions, got %v", err)
	}
	if len(loadedTransactions) == 0 {
		t.Errorf("Expected transactions to be saved")
	}
}

func TestSaveTransactionsToDbInvalidData(t *testing.T) {
	if err := initDb(nil); err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}
	defer closeDb()

	// Test saving transactions with invalid year
	invalidTransactions := TransactionHistory{
		"invalid_year": {
			"january": {
				"expense": []Transaction{
					{Id: "1", Amount: 10.0, Category: "food", Description: "test"},
				},
			},
		},
	}

	if err := saveTransactionsToDb(invalidTransactions); err == nil {
		t.Errorf("Expected error saving transactions with invalid year")
	}
}

// sets up a temp data dir with config pointing into it, returns the dir
func setupSecureStorageTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	original := globalConfig
	SetGlobalConfig(&Config{
		StorageType:       StorageSQLite,
		UnencryptedDbFile: filepath.Join(dir, "transactions.db"),
		EncryptedDBFile:   filepath.Join(dir, "transactions.enc"),
		SaltFile:          filepath.Join(dir, "transactions.salt"),
		LogFilePath:       filepath.Join(dir, "expense-tracking.log"),
	})
	clearHistory()
	t.Cleanup(func() {
		closeDb()
		clearUserPassword()
		clearHistory()
		SetGlobalConfig(original)
	})
	return dir
}

// lists every file in dir, used to make sure nothing besides the encrypted db and salt gets written
func listFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// the decrypted transactions must never be written to disk, only the encrypted file and the salt
func TestPlaintextNeverWrittenToDisk(t *testing.T) {
	dir := setupSecureStorageTest(t)
	const secret = "super-secret-description-42"

	// first run
	setUserPassword("correct horse")
	if err := openTransactions(); err != nil {
		t.Fatalf("first run openTransactions failed: %v", err)
	}

	tx := TransactionHistory{"2025": {"august": {"expense": []Transaction{{Id: "abcd1234", Amount: 42, Category: "food", Description: secret}}}}}
	if err := SaveTransactions(tx); err != nil {
		t.Fatalf("SaveTransactions failed: %v", err)
	}

	// while the app is "running" only the encrypted db and salt exist, and none of them contain the plaintext
	files := listFiles(t, dir)
	if len(files) != 2 {
		t.Errorf("expected only transactions.enc and transactions.salt on disk, got %v", files)
	}
	for _, name := range files {
		content, _ := os.ReadFile(filepath.Join(dir, name))
		if bytes.Contains(content, []byte(secret)) || bytes.Contains(content, []byte("SQLite format")) {
			t.Errorf("file %s contains plaintext transaction data", name)
		}
		info, _ := os.Stat(filepath.Join(dir, name))
		if info.Mode().Perm() != 0600 {
			t.Errorf("file %s has permissions %v, expected 0600", name, info.Mode().Perm())
		}
	}

	// simulate a hard kill: nothing is cleaned up, then log in again; the saved change must be there
	closeDb()
	clearUserPassword()

	setUserPassword("correct horse")
	if err := openTransactions(); err != nil {
		t.Fatalf("second login failed: %v", err)
	}
	loaded, err := LoadTransactions()
	if err != nil {
		t.Fatalf("LoadTransactions failed: %v", err)
	}
	if got := loaded["2025"]["august"]["expense"]; len(got) != 1 || got[0].Description != secret {
		t.Errorf("expected saved transaction after re-login, got %v", loaded)
	}
}

func TestOpenTransactionsWrongPassword(t *testing.T) {
	setupSecureStorageTest(t)

	setUserPassword("right")
	if err := openTransactions(); err != nil {
		t.Fatalf("first run openTransactions failed: %v", err)
	}
	closeDb()

	setUserPassword("wrong")
	if err := openTransactions(); !errors.Is(err, ErrWrongPassword) {
		t.Errorf("expected ErrWrongPassword, got %v", err)
	}
	if db != nil {
		t.Errorf("db must not be opened with a wrong password")
	}
}

// an encrypted db without its salt must fail clearly instead of silently generating a new salt
func TestOpenTransactionsMissingSalt(t *testing.T) {
	setupSecureStorageTest(t)

	setUserPassword("right")
	if err := openTransactions(); err != nil {
		t.Fatalf("first run openTransactions failed: %v", err)
	}
	closeDb()
	os.Remove(globalConfig.SaltFile)

	err := openTransactions()
	if err == nil || errors.Is(err, ErrWrongPassword) || !strings.Contains(err.Error(), "salt") {
		t.Errorf("expected missing salt error, got %v", err)
	}
	if _, err := os.Stat(globalConfig.SaltFile); !os.IsNotExist(err) {
		t.Errorf("a new salt file must not be generated for an existing encrypted db")
	}
}

// a plaintext db left by an older version is imported (when no encrypted db exists) and then deleted
func TestOpenTransactionsMigratesLegacyPlaintextDb(t *testing.T) {
	setupSecureStorageTest(t)

	// build a legacy plaintext db file
	if err := initDb(nil); err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}
	legacyTx := TransactionHistory{"2024": {"may": {"income": []Transaction{{Id: "legacy01", Amount: 100, Category: "salary", Description: "legacy"}}}}}
	if err := saveTransactionsToDb(legacyTx); err != nil {
		t.Fatalf("Failed to save transactions: %v", err)
	}
	data, err := serializeDb()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}
	closeDb()
	if err := os.WriteFile(globalConfig.UnencryptedDbFile, data, 0644); err != nil {
		t.Fatalf("Failed to write legacy db: %v", err)
	}

	setUserPassword("pass")
	if err := openTransactions(); err != nil {
		t.Fatalf("openTransactions failed: %v", err)
	}

	if _, err := os.Stat(globalConfig.UnencryptedDbFile); !os.IsNotExist(err) {
		t.Errorf("legacy plaintext db must be removed after migration")
	}
	loaded, err := LoadTransactions()
	if err != nil {
		t.Fatalf("LoadTransactions failed: %v", err)
	}
	if got := loaded["2024"]["may"]["income"]; len(got) != 1 || got[0].Description != "legacy" {
		t.Errorf("expected migrated legacy transaction, got %v", loaded)
	}

	// and the data must now be in the encrypted file
	closeDb()
	if err := openTransactions(); err != nil {
		t.Fatalf("re-open failed: %v", err)
	}
	loaded, _ = LoadTransactions()
	if got := loaded["2024"]["may"]["income"]; len(got) != 1 {
		t.Errorf("expected migrated transaction in encrypted db, got %v", loaded)
	}
}

// a stale plaintext db next to an existing encrypted db gets deleted on login
func TestOpenTransactionsRemovesStalePlaintextDb(t *testing.T) {
	setupSecureStorageTest(t)

	setUserPassword("pass")
	if err := openTransactions(); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	closeDb()

	if err := os.WriteFile(globalConfig.UnencryptedDbFile, []byte("stale plaintext"), 0644); err != nil {
		t.Fatalf("Failed to write stale db: %v", err)
	}
	if err := openTransactions(); err != nil {
		t.Fatalf("openTransactions failed: %v", err)
	}
	if _, err := os.Stat(globalConfig.UnencryptedDbFile); !os.IsNotExist(err) {
		t.Errorf("stale plaintext db must be removed on login")
	}
}

func TestRestrictDataFilePermissions(t *testing.T) {
	dir := setupSecureStorageTest(t)
	for _, p := range []string{globalConfig.EncryptedDBFile, globalConfig.SaltFile, globalConfig.LogFilePath} {
		if err := os.WriteFile(p, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		os.Chmod(p, 0644) // WriteFile is subject to umask
	}

	restrictDataFilePermissions(globalConfig)

	for _, name := range listFiles(t, dir) {
		info, _ := os.Stat(filepath.Join(dir, name))
		if info.Mode().Perm() != 0600 {
			t.Errorf("%s has permissions %v, expected 0600", name, info.Mode().Perm())
		}
	}
}

func TestDefaultConfigRestrictsDirPermissions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dataDir := filepath.Join(home, defaultExpenseToolDir)

	// new dir
	if _, err := DefaultConfig(); err != nil {
		t.Fatalf("DefaultConfig failed: %v", err)
	}
	if info, _ := os.Stat(dataDir); info.Mode().Perm() != 0700 {
		t.Errorf("new data dir has permissions %v, expected 0700", info.Mode().Perm())
	}

	// existing dir with loose permissions (older versions created it with 0755)
	os.Chmod(dataDir, 0755)
	if _, err := DefaultConfig(); err != nil {
		t.Fatalf("DefaultConfig failed: %v", err)
	}
	if info, _ := os.Stat(dataDir); info.Mode().Perm() != 0700 {
		t.Errorf("existing data dir has permissions %v, expected 0700", info.Mode().Perm())
	}
}
