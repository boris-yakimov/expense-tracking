package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	StorageJSONFile StorageType = "json"
	StorageSQLite   StorageType = "sqlite"

	DescriptionMaxCharLength = 40
	TransactionIDLength      = 8

	defaultUnencryptedDb = ".transactions.db"
	defaultEncryptedDb   = ".transactions.enc"
	defaultSaltFile      = ".transactions.salt"
	defaultJsonFile      = ".transactions.json"
	defaultLogFile       = ".expense-tracking.log"

	// encryption configuration
	keyLen     = 32      // AES-256 key length
	iterations = 200_000 // PBKDF2 iterations for key derivation
	saltLen    = 16      // Salt length in bytes
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
	// TODO: should this file be moved somewhere more hidden sice it is for the temporary unencrypted file that exists only while the app is running
	unencryptedDbFilePath := filepath.Join(homeDir, defaultUnencryptedDb)
	encryptedDbFilePath := filepath.Join(homeDir, defaultEncryptedDb)
	jsonFilePath := filepath.Join(homeDir, defaultJsonFile)
	logFilePath := filepath.Join(homeDir, defaultLogFile)
	saltFilePath := filepath.Join(homeDir, defaultSaltFile)
	// TODO: check if those paths will also work on windows

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
