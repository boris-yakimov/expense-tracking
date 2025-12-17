package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// previously also supported JSON but was deprecated, leaving the current approach in case I want to extend with other storage options in the future
	StorageSQLite StorageType = "sqlite"

	defaultExpenseToolDir = ".expense-tracking"
	defaultUnencryptedDb  = "transactions.db"
	defaultEncryptedDb    = "transactions.enc"
	defaultSaltFile       = "transactions.salt"
	defaultLogFile        = "expense-tracking.log"

	// encryption configuration
	keyLen     = 32      // AES-256 key length
	iterations = 200_000 // PBKDF2 iterations for key derivation
	saltLen    = 16      // Salt length in bytes

	DescriptionMaxCharLength = 160

	TransactionIDLength = 8
)

type Config struct {
	StorageType       StorageType
	UnencryptedDbFile string
	LogFilePath       string
	EncryptedDBFile   string
	SaltFile          string
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

	encryptedDbFilePath := filepath.Join(expenseToolDir, defaultEncryptedDb)
	unencryptedDbFilePath := filepath.Join(expenseToolDir, defaultUnencryptedDb)
	logFilePath := filepath.Join(expenseToolDir, defaultLogFile)
	saltFilePath := filepath.Join(expenseToolDir, defaultSaltFile)

	return &Config{
		StorageType:       StorageSQLite,
		UnencryptedDbFile: unencryptedDbFilePath,
		EncryptedDBFile:   encryptedDbFilePath,
		LogFilePath:       logFilePath,
		SaltFile:          saltFilePath,
	}, nil
}

type StorageType string

// determine storage type and storage paths from env vars
func loadConfigFromEnvVars() (*Config, error) {
	config, err := DefaultConfig() // sqlite
	if err != nil {
		return nil, fmt.Errorf("failed to use default config, err: %w", err)
	}

	if encryptedDbFilePath := os.Getenv("EXPENSE_ENCRYPTED_DB_PATH"); encryptedDbFilePath != "" {
		config.EncryptedDBFile = encryptedDbFilePath
	}

	if unencryptedDbFilePath := os.Getenv("EXPENSE_UNENCRYPTED_DB_PATH"); unencryptedDbFilePath != "" {
		config.UnencryptedDbFile = unencryptedDbFilePath
	}

	if logFilePath := os.Getenv("EXPENSE_LOG_PATH"); logFilePath != "" {
		config.LogFilePath = logFilePath
	}

	if saltFilePath := os.Getenv("EXPENSE_SALT_PATH"); saltFilePath != "" {
		config.SaltFile = saltFilePath
	}

	return config, nil
}

func createLogFileIfNotPresent(logFilePath string) (logFile *os.File, err error) {
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("unable to open log file for writing %s, err: %w", logFilePath, err)
	}
	return logFile, nil
}
