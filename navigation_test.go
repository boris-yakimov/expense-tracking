package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestVimNavigation(t *testing.T) {
	// Test 'j' key (move down)
	event := tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone)
	result := vimNavigation(event)
	if result.Key() != tcell.KeyDown {
		t.Errorf("Expected 'j' key to be converted to KeyDown, got %v", result.Key())
	}

	// Test 'k' key (move up)
	event = tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone)
	result = vimNavigation(event)
	if result.Key() != tcell.KeyUp {
		t.Errorf("Expected 'k' key to be converted to KeyUp, got %v", result.Key())
	}

	// Test other rune keys (should pass through unchanged)
	event = tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
	result = vimNavigation(event)
	if result.Key() != tcell.KeyRune || result.Rune() != 'a' {
		t.Errorf("Expected 'a' key to pass through unchanged")
	}

	// Test non-rune keys (should pass through unchanged)
	event = tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	result = vimNavigation(event)
	if result.Key() != tcell.KeyEnter {
		t.Errorf("Expected Enter key to pass through unchanged")
	}
}

func TestExitShortcuts(t *testing.T) {
	// Test ESC key
	event := tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone)
	result := exitShortcuts(event)
	if result != nil {
		t.Errorf("Expected ESC key to be consumed (return nil), got %v", result)
	}

	// Test 'q' key
	event = tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)
	result = exitShortcuts(event)
	if result != nil {
		t.Errorf("Expected 'q' key to be consumed (return nil), got %v", result)
	}

	// Test 'Q' key
	event = tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)
	result = exitShortcuts(event)
	if result != nil {
		t.Errorf("Expected 'Q' key to be consumed (return nil), got %v", result)
	}

	// Test other keys (should pass through unchanged)
	event = tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
	result = exitShortcuts(event)
	if result.Key() != tcell.KeyRune || result.Rune() != 'a' {
		t.Errorf("Expected 'a' key to pass through unchanged")
	}

	// Test non-rune keys (should pass through unchanged)
	event = tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	result = exitShortcuts(event)
	if result.Key() != tcell.KeyEnter {
		t.Errorf("Expected Enter key to pass through unchanged")
	}
}
