package main

import (
	"testing"
)

func TestAddInitialPassword(t *testing.T) {
	// Test with empty password
	err := addInitialPassword("")
	if err == nil {
		t.Errorf("Expected error with empty password")
	}

	// Test with valid password
	err = addInitialPassword("testpassword")
	if err != nil {
		t.Errorf("Expected no error with valid password, got %v", err)
	}
	if userPassword != "testpassword" {
		t.Errorf("Expected userPassword to be set to 'testpassword', got %s", userPassword)
	}

	// Clean up
	clearUserPassword()
}

func TestAddInitialPasswordWithSpecialCharacters(t *testing.T) {
	// Test with password containing special characters
	testPassword := "test@password#123"
	err := addInitialPassword(testPassword)
	if err != nil {
		t.Errorf("Expected no error with special character password, got %v", err)
	}
	if userPassword != testPassword {
		t.Errorf("Expected userPassword to be set to '%s', got %s", testPassword, userPassword)
	}

	// Clean up
	clearUserPassword()
}

func TestAddInitialPasswordWithSpaces(t *testing.T) {
	// Test with password containing spaces
	testPassword := "test password with spaces"
	err := addInitialPassword(testPassword)
	if err != nil {
		t.Errorf("Expected no error with password containing spaces, got %v", err)
	}
	if userPassword != testPassword {
		t.Errorf("Expected userPassword to be set to '%s', got %s", testPassword, userPassword)
	}

	// Clean up
	clearUserPassword()
}

func TestAddInitialPasswordWithUnicode(t *testing.T) {
	// Test with password containing unicode characters
	testPassword := "test密码123"
	err := addInitialPassword(testPassword)
	if err != nil {
		t.Errorf("Expected no error with unicode password, got %v", err)
	}
	if userPassword != testPassword {
		t.Errorf("Expected userPassword to be set to '%s', got %s", testPassword, userPassword)
	}

	// Clean up
	clearUserPassword()
}
