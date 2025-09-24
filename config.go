package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	StorageJSONFile StorageType = "json"
	StorageSQLite   StorageType = "sqlite"

	defaultExpenseToolDir = ".expense-tracking"
	defaultUnencryptedDb  = "transactions.db"
	defaultEncryptedDb    = "transactions.enc"
	defaultSaltFile       = "transactions.salt"
	defaultJsonFile       = "transactions.json"
	defaultLogFile        = "expense-tracking.log"

	// encryption configuration
	keyLen     = 32      // AES-256 key length
	iterations = 200_000 // PBKDF2 iterations for key derivation
	saltLen    = 16      // Salt length in bytes

	DescriptionMaxCharLength = 40
	TransactionIDLength      = 8
)

type Config struct {
	StorageType     StorageType
	SQLitePath      string
	JSONFilePath    string
	LogFilePath     string
	EncryptedDBFile string
	SaltFile        string
}

func SetGlobalConfig(config *Config) {
	globalConfig = config
}

var globalConfig *Config

func DefaultConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user's home directory: %w", err)
	}

	// TODO: test if those paths will also work on windows
	expenseToolDir := filepath.Join(homeDir, defaultExpenseToolDir)
	if _, err := os.Stat(expenseToolDir); err != nil {
		if os.IsNotExist(err) { // directory doesn't exist, create it
			if err := os.Mkdir(expenseToolDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create %s dir, err: %w", expenseToolDir, err)
			}
		} else { // other errors, like permission denied, etc
			return nil, fmt.Errorf("failed to check if %s dir exists, err: %w", expenseToolDir, err)
		}
	}

	// TODO: test if those paths will also work on windows
	encryptedDbFilePath := filepath.Join(expenseToolDir, defaultEncryptedDb)
	unencryptedDbFilePath := filepath.Join(expenseToolDir, defaultUnencryptedDb)
	jsonFilePath := filepath.Join(expenseToolDir, defaultJsonFile)
	logFilePath := filepath.Join(expenseToolDir, defaultLogFile)
	saltFilePath := filepath.Join(expenseToolDir, defaultSaltFile)

	return &Config{
		StorageType:     StorageSQLite,
		SQLitePath:      unencryptedDbFilePath,
		EncryptedDBFile: encryptedDbFilePath,
		LogFilePath:     logFilePath,
		JSONFilePath:    jsonFilePath,
		SaltFile:        saltFilePath,
	}, nil
}

type StorageType string

// determine storage type and storage paths from env vars
func loadConfigFromEnvVars() (*Config, error) {
	config, err := DefaultConfig() // sqlite
	if err != nil {
		return nil, fmt.Errorf("failed to use default config, err: %w", err)
	}

	if storageType := os.Getenv("EXPENSE_STORAGE_TYPE"); storageType != "" {
		if storageType == string(StorageJSONFile) {
			config.StorageType = StorageJSONFile
		} else if storageType == string(StorageSQLite) {
			config.StorageType = StorageSQLite
		}
	}

	if sqlitePath := os.Getenv("EXPENSE_SQLITE_PATH"); sqlitePath != "" {
		config.SQLitePath = sqlitePath
	}

	if jsonPath := os.Getenv("EXPENSE_JSON_PATH"); jsonPath != "" {
		config.JSONFilePath = jsonPath
	}

	// TODO: option for the user to select path for salt, .enc, .db (temp file), log

	return config, nil
}

func createLogFileIfNotPresent(logFilePath string) (logFile *os.File, err error) {
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("unable to open log file for writing %s, err: %w", logFilePath, err)
	}
	return logFile, nil
}
