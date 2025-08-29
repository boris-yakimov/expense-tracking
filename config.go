package main

import (
	"os"
)

const (
	StorageJSONFile StorageType = "json"
	StorageSQLite   StorageType = "sqlite"

	DescriptionMaxCharLength = 40
	TransactionIDLength      = 8
)

var globalConfig *Config

type Config struct {
	StorageType  StorageType
	SQLitePath   string
	JSONFilePath string
}

func SetGlobalConfig(config *Config) {
	globalConfig = config
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		StorageType:  StorageSQLite, // Default to SQLite for better performance
		SQLitePath:   "db/transactions.db",
		JSONFilePath: "db/transactions.json",
	}
}

type StorageType string

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
