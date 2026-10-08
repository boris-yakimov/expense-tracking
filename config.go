package main

import (
	"fmt"
	"log"
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
	StorageType StorageType
	// plaintext db of older versions, never written anymore; only imported on first run if present and then deleted
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
	if info, err := os.Stat(expenseToolDir); err != nil {
		if os.IsNotExist(err) { // directory doesn't exist, create it, only accessible by the current user
			if err := os.Mkdir(expenseToolDir, 0700); err != nil {
				return nil, fmt.Errorf("failed to create %s dir, err: %w", expenseToolDir, err)
			}
		} else { // other errors, like permission denied, etc
			return nil, fmt.Errorf("failed to check if %s dir exists, err: %w", expenseToolDir, err)
		}
	} else if info.IsDir() && info.Mode().Perm()&0077 != 0 {
		// created by an older version (0755) or by hand, don't let other users on the machine list or read it
		if err := os.Chmod(expenseToolDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to restrict permissions of %s dir, err: %w", expenseToolDir, err)
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
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("unable to open log file for writing %s, err: %w", logFilePath, err)
	}
	return logFile, nil
}

// makes the encrypted db, salt and log files readable only by the current user
// they are often copied in by hand (backups, other machines) and keep whatever permissions the copy gave them
// the db is encrypted, but a readable .enc + .salt still allows other local users to brute force the password offline
func restrictDataFilePermissions(config *Config) {
	if config == nil {
		return
	}
	for _, path := range []string{config.EncryptedDBFile, config.SaltFile, config.LogFilePath, config.UnencryptedDbFile} {
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 == 0 {
			continue
		}
		if err := os.Chmod(path, 0600); err != nil {
			log.Printf("warning: failed to restrict permissions of %s: %s\n", path, err)
		}
	}
}
