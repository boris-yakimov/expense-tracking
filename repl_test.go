package main

import (
	"testing"
)

// TODO: write tests for expnese add with a variety of inputs
// TODO: write tests for expense del
// TODO: tests for list and show-total
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
