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
)

type Config struct {
	StorageType  StorageType
	SQLitePath   string
	JSONFilePath string
	LogFilePath  string
}

func SetGlobalConfig(config *Config) {
	globalConfig = config
}

var globalConfig *Config

func DefaultConfig() *Config {
	return &Config{
		StorageType:  StorageSQLite,
		SQLitePath:   "db/transactions.db",
		JSONFilePath: "db/transactions.json",
		LogFilePath:  ".expense-tracking.log",
	}
}

type StorageType string

// determine storage type and storage paths from env vars
func loadConfigFromEnvVars() *Config {
	config := DefaultConfig() // sqlite

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

	return config
}

func createLogFileIfNotPresent(logFilePath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user's home directory: %w", err)
	}

	// log file is created in the user's home directory
	logFilePath = filepath.Join(homeDir, logFilePath)

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		// log file does not exist, so create it
		logFile, err := os.Create(logFilePath)
		if err != nil {
			return fmt.Errorf("unable to create log file %s , err: %w", logFilePath, err)
		}

		defer logFile.Close()
	}

	return nil
}
