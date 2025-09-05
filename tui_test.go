package main

import (
	"testing"

	"github.com/rivo/tview"
)

func TestMainMenu(t *testing.T) {
	// Initialize tui for testing
	tui = tview.NewApplication()

	// Set up a test config to avoid nil pointer dereference
	testConfig := &Config{
		StorageType:  StorageSQLite,
		SQLitePath:   "test.db",
		JSONFilePath: "test.json",
	}
	SetGlobalConfig(testConfig)

	// Test that mainMenu doesn't panic and returns no error
	err := mainMenu()
	if err != nil {
		t.Errorf("Expected no error from mainMenu, got %v", err)
	}
}
