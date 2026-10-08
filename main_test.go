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
	// Test that setupGracefulShutdown doesn't panic
	setupGracefulShutdown()
	// stop delivering signals to the handler again, otherwise a ctrl+c during the test run would exit the test binary
	signal.Reset(os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
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

	// Set HOME for test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	testHome := "/tmp/test_home"
	os.Setenv("HOME", testHome)
	os.MkdirAll(testHome, 0755)

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
