package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	// files copied in manually (backups, another machine) keep their original, often world readable, permissions
	restrictDataFilePermissions(config)

	// make sure the tui is stopped (and the cleanup below runs) if the process gets killed or the terminal closed
	setupGracefulShutdown()

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
		shutdown()
		os.Exit(1)
	}

	shutdown()
	log.Printf("Exit Expense Tracking Tool")
}

// every change is already encrypted to disk when it is saved, this is only a last safety net before the in-memory db is dropped
// runs once, both a normal exit and the signal handler can call it
func shutdown() {
	shutdownOnce.Do(func() {
		if db != nil && userPassword != "" {
			if err := persistDb(); err != nil {
				log.Printf("failed to encrypt database on shutdown: %s\n", err)
			}
		}
		closeDb()
		clearUserPassword() // clear password from memory
	})
}

var shutdownOnce sync.Once

// stops the tui on ctrl+c, kill (SIGTERM) or a closed terminal (SIGHUP), so main() can restore the terminal and run shutdown()
// the decrypted data only lives in memory, so even a hard kill (SIGKILL, power loss) leaves nothing readable on disk
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		sig := <-c // blocks until a signal is received
		log.Printf("received %s, shutting down\n", sig)

		if tui != nil {
			go tui.Stop()
			// tui.Stop may never return if the terminal is already gone (SIGHUP), don't wait for it forever
			time.Sleep(3 * time.Second)
		}
		shutdown()
		os.Exit(1)
	}()
}
