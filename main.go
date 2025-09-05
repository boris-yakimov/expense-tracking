package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var tui *tview.Application

func main() {
	// load configuration
	config := loadConfigFromEnvVars()
	SetGlobalConfig(config)

	// set up graceful shutdown handler to make sure database re-encryption happens even if the tui gets killed
	setupGracefulShutdown(config)

	tui = tview.NewApplication()
	tui.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		screen.Clear()
		screen.Fill(' ', tcell.StyleDefault.Background(theme.BackgroundColor))
		return false
	})

	if err := loginForm(); err != nil {
		fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
		os.Exit(1)
	}

	// option to migrate data from JSON to SQLite
	if os.Getenv("MIGRATE_TRANSACTION_DATA") == "true" {
		if config.StorageType != StorageSQLite {
			fmt.Fprintf(os.Stderr, "migration can only be performed when using SQLite storage")
			os.Exit(1)
		}
		if err := migrateJsonToDb(); err != nil {
			fmt.Fprintf(os.Stderr, "executed migration from json to db because MIGRATE_TRANSACTION_DATA=true was set, however migration failed with err: %v", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "successfully executed db migration from json to sqlite db\n")
	}

	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui failed: %v\n", err)
		os.Exit(1)
	}

	// on normal shutdown, close and re-encrypt DB if user was authenticated
	if config.StorageType == StorageSQLite {
		closeDb()
		if userPassword != "" {
			if err := encryptDatabase(config.SQLitePath); err != nil {
				fmt.Fprintf(os.Stderr, "failed to encrypt database on shutdown: %v\n", err)
			} else {
				// remove unencrypted database file after successful encryption
				if err := os.Remove(config.SQLitePath); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to remove plaintext database: %v\n", err)
				}
			}
		}
	}
}

// sets up signal handling to ensure database encryption on exit
// TODO: undertsand how this works
func setupGracefulShutdown(config *Config) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		// close db and encrypt database before exiting
		if config.StorageType == StorageSQLite {
			closeDb()
			if userPassword != "" {
				if err := encryptDatabase(config.SQLitePath); err != nil {
					fmt.Fprintf(os.Stderr, "failed to encrypt database on shutdown: %v\n", err)
				} else {
					// remove unencrypted database file after successful encryption
					if err := os.Remove(config.SQLitePath); err != nil {
						fmt.Fprintf(os.Stderr, "warning: failed to remove plaintext database: %v\n", err)
					}
				}
			}
		}
		clearUserPassword()
		os.Exit(0)
	}()
}
