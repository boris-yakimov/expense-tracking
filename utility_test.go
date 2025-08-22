package main

import (
	"testing"
)

func TestCleanTerminalInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "HelloWo rld!",
			expected: []string{"hellowo", "rld!"},
		},
		{
			input:    "hello world 123 te!t",
			expected: []string{"hello", "world", "123", "te!t"},
		},
	}

	for _, c := range cases {
		actual := cleanTerminalInput(c.input)
		actualSize := len(actual)
		expectedSize := len(c.expected)
		if actualSize != expectedSize {
			t.Errorf("expected size: %v\n, got size: %v\n", expectedSize, actualSize)
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("expected: %s\n, got: %s\n", expectedWord, word)
			}
		}
	}
}

func TestValidDescriptionInputFormat(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{"correct description", true},
		{"Valid description 123", true},
		{"Another-description, with commas", true},
		{"dash-separated-description", true},
		{"description with 'single quotes'", true},
		{"contains_underscore", false},
		{"contains@symbol", false},
		{"contains/slash", false},
		{"", false},   // empty string is not valid based on the regex
		{"   ", true}, // spaces only, allowed by regex
		{"strings that is too long for what might be expected as a description, but is still valid also includes - and Capital letter", true},
	}

	for _, c := range cases {
		validFormat := validDescriptionInputFormat(c.input)
		if validFormat != c.expected {
			t.Errorf("validDescriptionInputFormat(%q) = %v; expected %v", c.input, validFormat, c.expected)
		}
	}
}

func TestNormalizeTransactionType(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"expense", "expense", false},
		{"expenses", "expense", false},
		{"Expense", "expense", false},
		{"Expenses", "expense", false},
		{"income", "income", false},
		{"Income", "income", false},
		{"investment", "investment", false},
		{"investments", "investment", false},
		{"Investment", "investment", false},
		{"Investments", "investment", false},
		{"invalid", "", true},
		{"", "", true},
		{"random", "", true},
		{"EXPENSE", "", true},    // uppercase not supported
		{"INCOME", "", true},     // uppercase not supported
		{"INVESTMENT", "", true}, // uppercase not supported
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			result, err := normalizeTransactionType(c.input)

			if (err != nil) != c.hasError {
				t.Errorf("normalizeTransactionType(%q) error = %v; expected error = %v", c.input, err, c.hasError)
			}

			if !c.hasError && result != c.expected {
				t.Errorf("normalizeTransactionType(%q) = %q; expected %q", c.input, result, c.expected)
			}
		})
	}
}

func TestGenerateTransactionId(t *testing.T) {
	// Test that IDs are generated and have correct length
	for i := 0; i < 10; i++ {
		id, err := generateTransactionId()
		if err != nil {
			t.Errorf("generateTransactionId() returned error: %v", err)
		}
		if len(id) != 8 {
			t.Errorf("generateTransactionId() returned ID of length %d; expected 8", len(id))
		}
		// Check that ID contains only alphanumeric characters
		for _, char := range id {
			if !((char >= '0' && char <= '9') || (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')) {
				t.Errorf("generateTransactionId() returned ID with invalid character: %c", char)
			}
		}
	}
}

func TestCapitalize(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"test", "Test"},
		{"a", "A"},
		{"", ""},
		{"already Capitalized", "Already capitalized"}, // function converts to lowercase after first char
		{"MIXED case", "Mixed case"},                   // function converts to lowercase after first char
		{"123", "123"},
		{"hello world", "Hello world"},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			result := capitalize(c.input)
			if result != c.expected {
				t.Errorf("capitalize(%q) = %q; expected %q", c.input, result, c.expected)
			}
		})
	}
}
