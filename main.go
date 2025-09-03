package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var tui *tview.Application

func main() {
	// load configuration
	config := loadConfigFromEnvVars()
	SetGlobalConfig(config)

	// initialize db only if using SQLite storage
	if config.StorageType == StorageSQLite {
		if err := initDb(config.SQLitePath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize DB with err: \n\n%v", err)
			os.Exit(1)
		}
		defer closeDb()
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

	tui = tview.NewApplication()
	tui.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		screen.Clear()
		screen.Fill(' ', tcell.StyleDefault.Background(theme.BackgroundColor))
		return false
	})

	// TODO: move os.Exit here
	// TODO: check is this a first login, i.e. no previous password has been set
	loginForm()

	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui failed: %v\n", err)
		os.Exit(1)
	}
}
