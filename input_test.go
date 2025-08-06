package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
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
		actual := cleanInput(c.input)
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

func TestValidDescriptionFormat(t *testing.T) {
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

// TODO: tests for normalizeTransactionType
// TODO: tests for generateTransactionId - check for errors; check for exact length of id
