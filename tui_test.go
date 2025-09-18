package main

import (
	"strings"
	"testing"
)

func TestGenerateCombinedControlsFooter(t *testing.T) {
	footer := generateCombinedControlsFooter()
	expectedParts := []string{"ESC", "q", "back", "TAB", "next", "j/k", "↑/↓", "navigate"}
	for _, part := range expectedParts {
		if !strings.Contains(footer, part) {
			t.Errorf("Expected footer to contain '%s', got %s", part, footer)
		}
	}
}

func TestGenerateWindowNavigationFooter(t *testing.T) {
	footer := generateWindowNavigationFooter()
	expectedParts := []string{"ESC", "q", "back", "m", "select month", "TAB", "next table"}
	for _, part := range expectedParts {
		if !strings.Contains(footer, part) {
			t.Errorf("Expected footer to contain '%s', got %s", part, footer)
		}
	}
}

func TestGenerateTransactionCrudFooter(t *testing.T) {
	footer := generateTransactionCrudFooter()
	expectedParts := []string{"a", "add", "d", "delete", "e/u", "update"}
	for _, part := range expectedParts {
		if !strings.Contains(footer, part) {
			t.Errorf("Expected footer to contain '%s', got %s", part, footer)
		}
	}
}

func TestGenerateTransactionNavigationFooter(t *testing.T) {
	footer := generateTransactionNavigationFooter()
	expectedParts := []string{"j/k", "↑/↓", "move up and down", "h/l", "←/→", "move left and right"}
	for _, part := range expectedParts {
		if !strings.Contains(footer, part) {
			t.Errorf("Expected footer to contain '%s', got %s", part, footer)
		}
	}
}
