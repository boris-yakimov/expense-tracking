package main

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/rivo/tview"
)

func TestSetupGracefulShutdown(t *testing.T) {
	// Create a test config
	testConfig := &Config{
		StorageType:  StorageSQLite,
		SQLitePath:   "test.db",
		JSONFilePath: "test.json",
	}

	// Test that setupGracefulShutdown doesn't panic
	setupGracefulShutdown(testConfig)

	// Test that signal handler is set up by sending a signal
	// Note: This is a basic test - in a real scenario, we'd need to test
	// the actual signal handling behavior more thoroughly
}

func TestSetupGracefulShutdownWithJSONStorage(t *testing.T) {
	// Create a test config with JSON storage
	testConfig := &Config{
		StorageType:  StorageJSONFile,
		SQLitePath:   "test.db",
		JSONFilePath: "test.json",
	}

	// Test that setupGracefulShutdown doesn't panic with JSON storage
	setupGracefulShutdown(testConfig)
}

func TestSetupGracefulShutdownWithNilConfig(t *testing.T) {
	// Test that setupGracefulShutdown handles nil config gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("setupGracefulShutdown panicked with nil config: %v", r)
		}
	}()

	setupGracefulShutdown(nil)
}

func TestSignalHandling(t *testing.T) {
	// Test that signal channel is created and signal is notified
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Test that we can receive signals
	go func() {
		time.Sleep(10 * time.Millisecond)
		c <- os.Interrupt
	}()

	select {
	case sig := <-c:
		if sig != os.Interrupt {
			t.Errorf("Expected to receive Interrupt signal, got %v", sig)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Expected to receive signal within timeout")
	}
}

func TestMainFunctionDependencies(t *testing.T) {
	// Test that main function dependencies are available
	// This is a basic test to ensure the main function can be called
	// without panicking due to missing dependencies

	// Test loadConfigFromEnvVars
	config, err := loadConfigFromEnvVars()
	if err != nil {
		t.Errorf("Failed to load config from env var, err %v", err)
	}
	if config == nil {
		t.Errorf("Expected loadConfigFromEnvVars to return non-nil config")
	}

	// Test SetGlobalConfig
	SetGlobalConfig(config)
	if globalConfig == nil {
		t.Errorf("Expected globalConfig to be set")
	}

	// Test that tui can be created
	tui = tview.NewApplication()
	if tui == nil {
		t.Errorf("Expected tui to be created")
	}
}
