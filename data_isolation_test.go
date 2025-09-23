package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDataIsolation(t *testing.T) {
	// This test ensures that running tests doesn't modify actual transaction data
	// in the db/ directory

	// Get the current state of files in db/ directory
	dbDir := "db"
	originalFiles := make(map[string]fileState)

	// Read existing files in db directory
	if entries, err := os.ReadDir(dbDir); err == nil {
		for _, entry := range entries {
			if info, err := entry.Info(); err == nil {
				originalFiles[entry.Name()] = fileState{
					exists: true,
					size:   info.Size(),
					mtime:  info.ModTime(),
				}
			}
		}
	}

	// Run a comprehensive test that would normally touch the db/ directory
	t.Run("comprehensive_test", func(t *testing.T) {
		// Test SQLite operations
		setupTestStorage(t, StorageSQLite)

		// Test JSON operations
		setupTestStorage(t, StorageJSONFile)

		// Test encryption operations (these now use temporary files)
		TestEncryptDatabase(t)
		TestDecryptDatabase(t)

		// Test all major operations
		TestHandleAddTransaction(t)
		TestHandleDeleteTransaction(t)
		TestHandleUpdateTransaction(t)
		TestCalculateMonthPnL(t)
		TestCalculateYearPnL(t)
	})

	// Verify that files in db/ directory haven't been modified
	if entries, err := os.ReadDir(dbDir); err == nil {
		for _, entry := range entries {
			original, exists := originalFiles[entry.Name()]
			if !exists {
				// New file was created - this might be okay for some files
				t.Logf("New file created in db/ directory: %s", entry.Name())
				continue
			}

			currentInfo, err := entry.Info()
			if err != nil {
				t.Errorf("Failed to get info for file %s: %v", entry.Name(), err)
				continue
			}

			// Check if file was modified (size or modification time changed)
			if original.size != currentInfo.Size() {
				t.Errorf("File %s size changed from %d to %d", entry.Name(), original.size, currentInfo.Size())
			}

			// Allow for small time differences (within 1 second)
			timeDiff := currentInfo.ModTime().Sub(original.mtime)
			if timeDiff > time.Second || timeDiff < -time.Second {
				t.Errorf("File %s modification time changed significantly: %v", entry.Name(), timeDiff)
			}
		}
	}

	// Verify that no test files were left in db/ directory
	testPatterns := []string{"test_*", "*_test*", "*.backup", "*.tmp"}
	for _, pattern := range testPatterns {
		matches, err := filepath.Glob(filepath.Join(dbDir, pattern))
		if err != nil {
			t.Errorf("Error checking for test files with pattern %s: %v", pattern, err)
			continue
		}

		for _, match := range matches {
			// Check if this is a test file that should be cleaned up
			baseName := filepath.Base(match)
			if baseName != "transactions.db" && baseName != "transactions.json" &&
				baseName != "transactions.enc" && baseName != "transactions.salt" {
				t.Errorf("Test file left in db/ directory: %s", match)
			}
		}
	}
}

func TestNoActualDataModification(t *testing.T) {
	// This test specifically ensures that the actual transaction files
	// in db/ directory are not modified by tests

	dbFiles := []string{
		"db/transactions.db",
		"db/transactions.json",
		"db/transactions.enc",
		"db/transactions.salt",
	}

	// Record original file states
	originalStates := make(map[string]fileState)
	for _, file := range dbFiles {
		if info, err := os.Stat(file); err == nil {
			originalStates[file] = fileState{
				exists: true,
				size:   info.Size(),
				mtime:  info.ModTime(),
			}
		} else {
			originalStates[file] = fileState{
				exists: false,
				size:   0,
				mtime:  time.Time{},
			}
		}
	}

	// Run tests that might touch these files
	t.Run("run_all_tests", func(t *testing.T) {
		// Run a subset of tests that might interact with the db directory
		TestConfigSystem(t)
		TestSetGlobalConfig(t)
		TestLoadConfigFromEnvVars(t)
		TestValidDescriptionInputFormat(t)
		TestNormalizeTransactionType(t)
		TestGenerateTransactionId(t)
		TestCapitalize(t)
		TestListOfAllowedCategories(t)
		TestListOfAllowedTransactionTypes(t)
		TestEnforceCharLimit(t)
		TestStyleInputField(t)
		TestStyleDropdown(t)
		TestStyleForm(t)
		TestStyleGrid(t)
		TestStyleTable(t)
		TestStyleList(t)
		TestVimNavigation(t)
		TestAddInitialPassword(t)
		TestSetupGracefulShutdown(t)
		TestSignalHandling(t)
		TestMainFunctionDependencies(t)
	})

	// Verify that actual db files were not modified
	for _, file := range dbFiles {
		original := originalStates[file]
		currentInfo, err := os.Stat(file)

		if !original.exists {
			// File didn't exist originally
			if err == nil {
				t.Errorf("Test created file %s that didn't exist before", file)
			}
			continue
		}

		// File existed originally
		if err != nil {
			t.Errorf("Test deleted file %s that existed before", file)
			continue
		}

		// Check size
		if currentInfo.Size() != original.size {
			t.Errorf("Test modified file %s size: was %d, now %d", file, original.size, currentInfo.Size())
		}

		// Check modification time (allow 1 second tolerance)
		timeDiff := currentInfo.ModTime().Sub(original.mtime)
		if timeDiff > time.Second || timeDiff < -time.Second {
			t.Errorf("Test modified file %s modification time: %v", file, timeDiff)
		}
	}
}

type fileState struct {
	exists bool
	size   int64
	mtime  time.Time
}

// TestDatabaseFileProtection ensures that actual database files are never modified during tests
func TestDatabaseFileProtection(t *testing.T) {
	// This test runs at the very beginning to ensure no actual database files are touched
	actualDbFiles := []string{
		"db/transactions.db",
		"db/transactions.json",
		"db/transactions.enc",
		"db/transactions.salt",
	}

	// Record original file states
	originalStates := make(map[string]fileState)
	for _, file := range actualDbFiles {
		if info, err := os.Stat(file); err == nil {
			originalStates[file] = fileState{
				exists: true,
				size:   info.Size(),
				mtime:  info.ModTime(),
			}
		} else {
			originalStates[file] = fileState{
				exists: false,
				size:   0,
				mtime:  time.Time{},
			}
		}
	}

	// Run a comprehensive test suite that should not touch actual files
	t.Run("run_all_tests", func(t *testing.T) {
		// Test all major functionality using test storage
		TestConfigSystem(t)
		TestSetGlobalConfig(t)
		TestLoadConfigFromEnvVars(t)
		TestValidDescriptionInputFormat(t)
		TestNormalizeTransactionType(t)
		TestGenerateTransactionId(t)
		TestCapitalize(t)
		TestListOfAllowedCategories(t)
		TestListOfAllowedTransactionTypes(t)
		TestEnforceCharLimit(t)
		TestStyleInputField(t)
		TestStyleDropdown(t)
		TestStyleForm(t)
		TestStyleGrid(t)
		TestStyleTable(t)
		TestStyleList(t)
		TestVimNavigation(t)
		TestExitShortcuts(t)
		TestAddInitialPassword(t)
		TestSetupGracefulShutdown(t)
		TestSignalHandling(t)
		TestMainFunctionDependencies(t)

		// Test database operations with test storage
		TestInitDb(t)
		TestInitDbWithExistingFile(t)
		TestCloseDb(t)
		TestLoadTransactionsFromDb(t)
		TestSaveTransactionsToDb(t)
		TestSaveTransactionsToDbInvalidData(t)

		// Test JSON operations with test storage
		TestLoadTransactionsFromJsonFile(t)
		TestSaveTransactionsToJsonFile(t)
		TestLoadTransactionsFromJsonFileEmptyFile(t)
		TestLoadTransactionsFromJsonFileInvalidJSON(t)

		// Test migration with test storage
		TestMigrateJsonToDb(t)
		TestMigrateJsonToDbEmptyJson(t)
		TestMigrateJsonToDbInvalidData(t)
		TestMigrateJsonToDbInvalidMonth(t)

		// Test encryption with test storage
		TestSetUserPassword(t)
		TestGenerateSalt(t)
		TestSaveAndLoadSalt(t)
		TestLoadSaltNotFound(t)
		TestGetOrCreateSalt(t)
		TestDeriveEncryptionKey(t)
		TestEncryptTransactions(t)
		TestDecryptTransactionsInvalidData(t)
		TestEncryptDatabase(t)
		TestEncryptDatabaseNoPassword(t)
		TestDecryptDatabase(t)
		TestDecryptDatabaseNoPassword(t)
		TestDecryptDatabaseNoEncryptedFile(t)

		// Test transaction operations with test storage
		TestLoadTransactions(t)
		TestSaveTransactions(t)
		TestCreateTransactionsTable(t)
		TestHandleAddTransaction(t)
		TestHandleDeleteTransaction(t)
		TestHandleUpdateTransaction(t)
		TestCalculateMonthPnL(t)
		TestCalculateYearPnL(t)

		// Test utility functions with test storage
		TestGetListOfDetailedTransactions(t)
		TestGetTransactionTypeById(t)
		TestGetMonthsWithTransactions(t)
		TestDetermineLatestMonthAndYear(t)

		// Test visualization with test storage
		TestShowAllowedCategories(t)
	})

	// Verify that actual db files were not modified
	for _, file := range actualDbFiles {
		original := originalStates[file]
		currentInfo, err := os.Stat(file)

		if !original.exists {
			// File didn't exist originally
			if err == nil {
				t.Errorf("Test created file %s that didn't exist before", file)
			}
			continue
		}

		// File existed originally
		if err != nil {
			t.Errorf("Test deleted file %s that existed before", file)
			continue
		}

		// Check size
		if currentInfo.Size() != original.size {
			t.Errorf("Test modified file %s size: was %d, now %d", file, original.size, currentInfo.Size())
		}

		// Check modification time (allow 1 second tolerance)
		timeDiff := currentInfo.ModTime().Sub(original.mtime)
		if timeDiff > time.Second || timeDiff < -time.Second {
			t.Errorf("Test modified file %s modification time: %v", file, timeDiff)
		}
	}
}

// TestActualDatabaseFilesNeverAccessed ensures that actual database files are never accessed during tests
func TestActualDatabaseFilesNeverAccessed(t *testing.T) {
	// This test verifies that the actual database files in db/ directory are never accessed
	// during normal test operations. It does this by checking that the files exist and
	// have not been modified by any test operations.

	actualDbFiles := []string{
		"db/transactions.db",
		"db/transactions.json",
		"db/transactions.enc",
		"db/transactions.salt",
	}

	// Record initial state
	initialStates := make(map[string]fileState)
	for _, file := range actualDbFiles {
		if info, err := os.Stat(file); err == nil {
			initialStates[file] = fileState{
				exists: true,
				size:   info.Size(),
				mtime:  info.ModTime(),
			}
		} else {
			initialStates[file] = fileState{
				exists: false,
				size:   0,
				mtime:  time.Time{},
			}
		}
	}

	// Run some tests that should use test storage only
	t.Run("test_storage_operations", func(t *testing.T) {
		// Test SQLite with test storage
		setupTestStorage(t, StorageSQLite)
		TestHandleAddTransaction(t)
		TestHandleDeleteTransaction(t)
		TestHandleUpdateTransaction(t)
		TestCalculateMonthPnL(t)
		TestCalculateYearPnL(t)

		// Test JSON with test storage
		setupTestStorage(t, StorageJSONFile)
		TestHandleAddTransaction(t)
		TestHandleDeleteTransaction(t)
		TestHandleUpdateTransaction(t)
		TestCalculateMonthPnL(t)
		TestCalculateYearPnL(t)

		// Test encryption with test storage
		TestEncryptDatabase(t)
		TestDecryptDatabase(t)
	})

	// Verify that actual database files were not touched
	for _, file := range actualDbFiles {
		initial := initialStates[file]
		currentInfo, err := os.Stat(file)

		if !initial.exists {
			// File didn't exist initially
			if err == nil {
				t.Errorf("Test created actual database file %s that didn't exist before", file)
			}
			continue
		}

		// File existed initially
		if err != nil {
			t.Errorf("Test deleted actual database file %s that existed before", file)
			continue
		}

		// Check that file was not modified
		if currentInfo.Size() != initial.size {
			t.Errorf("Test modified actual database file %s size: was %d, now %d", file, initial.size, currentInfo.Size())
		}

		// Check modification time (allow 1 second tolerance for file system operations)
		timeDiff := currentInfo.ModTime().Sub(initial.mtime)
		if timeDiff > time.Second || timeDiff < -time.Second {
			t.Errorf("Test modified actual database file %s modification time: %v", file, timeDiff)
		}
	}
}
