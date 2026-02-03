package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var tui *tview.Application
var logFile *os.File
var pages *tview.Pages

func main() {
	var err error
	// load configuration
	config, err := loadConfigFromEnvVars()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config from env var, err %v\n", err)
	}
	SetGlobalConfig(config)

	// log file is created in the user's home directory
	if logFile, err = createLogFileIfNotPresent(config.LogFilePath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log file: %v\n", err)
		os.Exit(1)
	}

	log.SetOutput(io.MultiWriter(logFile))
	log.SetFlags(log.LstdFlags | log.Lshortfile) // timestamps + file:line info

	// set up graceful shutdown handler to make sure database re-encryption happens even if the tui gets killed
	setupGracefulShutdown(config)

	tui = tview.NewApplication()
	tui.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		screen.Clear()
		screen.Fill(' ', tcell.StyleDefault.Background(theme.BackgroundColor))
		return false
	})

	// maintain a list of each page like add, delete, update transaction, login, etc instead of replacing the root every time we have to switch a screen because it was causing resizing issues
	// this way we add each page and we can easily switch between them on each funciton as needed
	pages = tview.NewPages()
	tui.SetRoot(pages, true)

	log.Printf("Start Expense Tracking Tool")

	if err := loginForm(); err != nil {
		log.Printf("login form failed to start: %s\n", err)
		os.Exit(1)
	}

	if err := tui.Run(); err != nil {
		log.Printf("tui failed to start: %s\n", err)
		os.Exit(1)
	}

	// on normal shutdown, close and re-encrypt DB if user was authenticated
	if config.StorageType == StorageSQLite {
		closeDb()
		if userPassword != "" {
			if err := encryptDatabase(config.UnencryptedDbFile); err != nil {
				log.Printf("failed to encrypt database on shutdown: %s\n", err)
			} else {
				// remove unencrypted database file after successful encryption
				if err := os.Remove(config.UnencryptedDbFile); err != nil {
					log.Printf("warning: failed to remove plaintext database: %s\n", err)
				}
			}
		}
	}

	log.Printf("Exit Expense Tracking Tool")
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
				if err := encryptDatabase(config.UnencryptedDbFile); err != nil {
					log.Printf("failed to encrypt database on shutdown: %s\n", err)
				} else {
					// remove unencrypted database file after successful encryption
					if err := os.Remove(config.UnencryptedDbFile); err != nil {
						log.Printf("warning: failed to remove plaintext database: %s\n", err)
					}
				}
			}
		}
		clearUserPassword() // clear password from memory
		if logFile != nil {
			if err := logFile.Sync(); err != nil {
				log.Printf("failed to sync log file: %s\n", err)
			}
			if err := logFile.Close(); err != nil {
				log.Printf("failed to close log file: %s\n", err)
			}
		}
		// os.Exit(0)
	}()
}
