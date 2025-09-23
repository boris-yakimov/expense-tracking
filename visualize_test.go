package main

import (
	"testing"
)

func TestShowAllowedCategories(t *testing.T) {
	cases := []struct {
		name            string
		transactionType string
		expectedError   bool
	}{
		{
			name:            "show expense categories",
			transactionType: "expense",
			expectedError:   false,
		},
		{
			name:            "show income categories",
			transactionType: "income",
			expectedError:   false,
		},
		{
			name:            "show investment categories",
			transactionType: "investment",
			expectedError:   false,
		},
		{
			name:            "invalid transaction type",
			transactionType: "invalidtype",
			expectedError:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := listOfAllowedCategories(c.transactionType)

			if (err != nil) != c.expectedError {
				t.Errorf("showAllowedCategories(%q) error = %v; expected error = %v",
					c.transactionType, err, c.expectedError)
			}
		})
	}
}
