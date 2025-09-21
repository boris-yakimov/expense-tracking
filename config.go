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

func createLogFileIfNotPresent(logFilePath string) (logFile *os.File, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user's home directory: %w", err)
	}

	// TODO: update this to work on windows as well

	// log file is created in the user's home directory
	logFilePath = filepath.Join(homeDir, logFilePath)

	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("unable to open log file for writing %s, err: %w", logFilePath, err)
	}
	return logFile, nil
}
