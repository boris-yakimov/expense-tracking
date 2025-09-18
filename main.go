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

	// TODO: the transactions.db/.enc should be created in the user's home dir as .transactions.db/.enc
	// TODO: log file to be creaeted in user's home dir .expense-tracking.log
	// TODO: path to database file and log should be configurable
	// TODO: update all places where we do a print or error modal, or success modal, etc to generate an event in the log as well
	// TODO: evaluate option to either keep and encrypt JSON or remove the JSON option altogether

	// log file is created in the user's home directory
	if err := createLogFileIfNotPresent(config.LogFilePath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log file: %v\n", err)
		os.Exit(1)
	}

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
func setupGracefulShutdown(config *Config) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // Interrupt = ctrl+c ; SIGTERM when process is killed

	// since this runs inside the goroutine, the shutdown logic happens asynchronously when triggered.
	go func() {
		<-c // blocks until a signal is received

		// close db and re-encrypt database before exiting
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
		clearUserPassword() // clear password from memory
		os.Exit(0)
	}()
}
