package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"text/tabwriter"
)

var monthOrder = map[string]int{
	"january":   1,
	"february":  2,
	"march":     3,
	"april":     4,
	"may":       5,
	"june":      6,
	"july":      7,
	"august":    8,
	"september": 9,
	"october":   10,
	"november":  11,
	"december":  12,
}

// helper to validate the the format of the description field - only letters, numbers, commas, spaces or dashes
func validDescriptionInputFormat(description string) bool {
	pattern := `^[a-zA-Z0-9,' '-]+$`
	matched, err := regexp.MatchString(pattern, description)
	if err != nil {
		return false
	}

	return matched
}

// TODO: do i still need this ?
// helper to make sure transaction types are standardized - lowercase and matching the expected name as in the db
func normalizeTransactionType(t string) (string, error) {
	switch t {

	case "expense", "expenses", "Expenses", "Expense":
		return "expense", nil

	case "investment", "investments", "Investments", "Investment":
		return "investment", nil

	case "income", "Income":
		return "income", nil

	default:
		return "", fmt.Errorf("invalid transaction type %s, supported transactions types are income, expense, and investment", t)
	}
}

// helper to capitalize some words, mainly used for months and transaction types when we visualize them in the TUI, e.g. July, August, etc
func capitalize(word string) string {
	if len(word) == 0 {
		return ""
	}

	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])

	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}

// helper to provide a list of allowed transaction categories
func listOfAllowedCategories(transactionType string) (categories []string, err error) {
	for c := range allowedTransactionCategories[transactionType] {
		categories = append(categories, c)
	}

	if len(categories) <= 0 {
		return categories, fmt.Errorf("something went wrong with getting list of allowed categories for transaction type %s", transactionType)
	}

	return categories, nil
}

// TODO: this seems to be duplicated bello, check which of the 2 functions is used in the latest tui
// helper to provide a list of allowed transaction types
func listOfAllowedTransactionTypes() (categories []string, err error) {
	var transactionTypes []string
	for t := range allowedTransactionTypes {
		transactionTypes = append(transactionTypes, t)
	}

	if len(transactionTypes) <= 0 {
		return transactionTypes, fmt.Errorf("something went wrong with getting list of allowed transaction types")
	}

	return transactionTypes, nil
}

// helper to build a list of transactions for visualization in the TUI
func getListOfDetailedTransactions() (listOfTransactions []string, err error) {
	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return listOfTransactions, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	// build a list of transaction details for display
	for year := range transactions {
		for month := range transactions[year] {
			for txType := range transactions[year][month] {
				for _, tx := range transactions[year][month][txType] {
					detail := fmt.Sprintf("ID: %s | Amount: €%.2f | Category: %s | Description: %s | Type: %s | %s %s",
						tx.Id, tx.Amount, tx.Category, tx.Description, txType, month, year)
					listOfTransactions = append(listOfTransactions, detail)
				}
			}
		}
	}

	return listOfTransactions, nil
}

// helper to get a specific transaction by its ID
func getTransactionTypeById(txId string) (txType string, err error) {
	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return "", fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	// find what type of transaction is the particular id related to
	for year := range transactions {
		for month := range transactions[year] {
			for txType := range transactions[year][month] {
				for _, tx := range transactions[year][month][txType] {
					if txId == tx.Id {
						return txType, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("transaction ID %s could not be found in transaction list", txId)
}

// helper to enforce the character limit of the description field
func enforceCharLimit(textToCheck string, lastChar rune) bool {
	return len(textToCheck) <= DescriptionMaxCharLength
}

// helper to build a list of transactions for a specific month
func listTransactionsByMonth(transactionType, month, year string) (success bool, err error) {
	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	if len(transactions) == 0 {
		fmt.Println("\nno transactions found")
		return true, nil
	}

	transactionType, err = normalizeTransactionType(transactionType)
	if err != nil {
		return false, fmt.Errorf("transaction type error: %w", err)
	}

	// transaction type header
	fmt.Println()
	fmt.Printf("  %s\n", capitalize(transactionType))
	fmt.Printf("  %s\n", strings.Repeat("-", len(transactionType)))

	// transaction table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "    ID\tAmount\tCategory\tDescription")
	fmt.Fprintln(w, "    --\t------\t--------\t-----------")

	for _, t := range transactions[year][month][transactionType] {
		fmt.Fprintf(w, "    %s\t€%.2f\t%s\t%s\n", t.Id, t.Amount, t.Category, t.Description)
	}

	w.Flush()
	fmt.Println()

	return true, nil
}

// TODO: this seems function seems to be duplicated above check which one is used
// helper to show list of allowed categories
func showAllowedCategories(transactionType string) error {
	fmt.Println("\nallowed categories are: ")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\nCategory\tDescription")
	fmt.Fprintln(w, "--------\t-----------")

	txType, err := normalizeTransactionType(transactionType)
	if err != nil {
		return fmt.Errorf(" show allowed categories err: %w", err)
	}

	for key, val := range allowedTransactionCategories[txType] {
		fmt.Fprintf(w, "%s\t%s\n", key, val)
	}
	w.Flush()
	return nil
}

// helper to get a list of months that have transactions - also  make sure these are sorted with newest to oldest month/year
func getMonthsWithTransactions() (months []string, err error) {
	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return months, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	for y := range transactions {
		for m := range transactions[y] {
			months = append(months, fmt.Sprintf("%s %s", m, y))
		}
	}

	return months, nil
}

// helper to determine what is the latest month that contains transactions
func determineLatestMonthAndYear() (month, year string, err error) {
	transactions, err := LoadTransactions()
	if err != nil {
		return "", "", fmt.Errorf("unable to load transactions file: %w", err)
	}

	// determine latest year
	var latestYear string
	for y := range transactions {
		if latestYear == "" || y > latestYear {
			latestYear = y
		}
	}

	// determine latest month for the year
	var latestMonth string
	if latestYear != "" {
		for m := range transactions[latestYear] {
			if latestMonth == "" || monthOrder[m] > monthOrder[latestMonth] {
				latestMonth = m
			}
		}
	}

	return latestMonth, latestYear, nil
}
