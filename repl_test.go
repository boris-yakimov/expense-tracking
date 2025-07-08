package main

import (
	// "fmt"
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
			expected: []string{"hellowo, rld!"},
		},
		{
			input:    "hello, world, 123, test",
			expected: []string{"hello", "world", "123", "test"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		// fmt.Println(actual)
		actualSize := len(actual)
		expectedSize := len(c.expected)
		// fmt.Printf("expected len: %v\n", expectedSize)
		// fmt.Printf("actual size: %v\n", actualSize)
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
