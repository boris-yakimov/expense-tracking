package main

import (
	"github.com/rivo/tview"
	"testing"
)

func TestMainMenu(t *testing.T) {
	// Initialize tui for testing
	tui = tview.NewApplication()

	// Test that mainMenu doesn't panic and returns no error
	err := mainMenu()
	if err != nil {
		t.Errorf("Expected no error from mainMenu, got %v", err)
	}
}
