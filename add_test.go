package main

import ()

// TODO: tests have to be reworked to fir the new TUI approach
//
// func TestAddTransaction(t *testing.T) {
// 	tmpFile := "test_add_transactions.json"
// 	transactionsFilePath = tmpFile
//
// 	// Clean up after test
// 	defer os.Remove(tmpFile)
//
// 	cases := []struct {
// 		input           []string
// 		expectedSuccess bool
// 		expectedError   bool
// 	}{
// 		{
// 			// valid add expense
// 			input: []string{
// 				"expense",
// 				"54.30",
// 				"food",
// 				"test food description",
// 			},
// 			expectedSuccess: true,
// 			expectedError:   false,
// 		},
// 		{
// 			input: []string{
// 				"expense",
// 				"5423.87",
// 				"renovation",
// 				"fence",
// 			},
// 			expectedSuccess: true,
// 			expectedError:   false,
// 		},
// 		{
// 			// invalid order of arguments - category before amount
// 			input: []string{
// 				"expense",
// 				"food",
// 				"32.55",
// 				"test with invalid order of arguments",
// 			},
// 			expectedSuccess: false,
// 			expectedError:   true,
// 		},
// 		{
// 			//	super long description
// 			input: []string{
// 				"expense",
// 				"100",
// 				"food",
// 				"very long description on an expense to make sure that we go above character limit",
// 			},
// 			expectedError:   true,
// 			expectedSuccess: false,
// 		},
// 		{
// 			// invalid transaction type
// 			input: []string{
// 				"expAnse",
// 				"54.30",
// 				"food",
// 				"test food description",
// 			},
// 			expectedSuccess: false,
// 			expectedError:   true,
// 		},
// 		{
// 			// invalid category
// 			input: []string{
// 				"expense",
// 				"54.30",
// 				"madeUpCategory",
// 				"test food description",
// 			},
// 			expectedSuccess: false,
// 			expectedError:   true,
// 		},
// 	}
//
// 	for _, c := range cases {
// 		status, err := addTransaction(c.input)
//
// 		if status != c.expectedSuccess {
// 			t.Errorf("provided input: add %v\nexpected result: %v\nreceived result: %v\n", c.input, c.expectedSuccess, status)
// 		}
//
// 		if (err != nil) != c.expectedError {
// 			t.Errorf("input: add %v\nexpected error: %v\nactual erorr: %s", c.input, c.expectedError, err)
// 		}
// 	}
// }
