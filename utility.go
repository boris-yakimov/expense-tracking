package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
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

// helper to get all details of a single trasnaction by its ID and return it
func getTransactionById(id string) (*Transaction, error) {
	transactions, err := LoadTransactions()
	if err != nil {
		return nil, fmt.Errorf("unable to load transactions: %w", err)
	}

	for _, months := range transactions {
		for _, types := range months {
			for _, txs := range types {
				for i := range txs {
					if txs[i].Id == id {
						return &txs[i], nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("transaction with ID %s not found", id)
}

// helper to enforce the character limit of the description field
func enforceCharLimit(textToCheck string, lastChar rune) bool {
	return len(textToCheck) <= DescriptionMaxCharLength
}

// helper to get a list of months that have transactions - also make sure these are sorted with newest to oldest month/year
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

	// sort year (descending), than month (descending)
	sort.Slice(months, func(i, j int) bool {
		// split the parts of each element in the list of months "year month" becomes []string{"year", "month"}
		partsI := strings.Split(months[i], " ")
		partsJ := strings.Split(months[j], " ")
		if len(partsI) != 2 || len(partsJ) != 2 {
			// if the string doesn't split in exactly 2 parts, fallback to plain string comparison (this is not expected happen)
			return months[i] > months[j] // reverse order comparison, i.e. Sep 2025 will come before Aug 2025
		}

		// parse year and convert it to integer for comparison
		yearI, _ := strconv.Atoi(partsI[1])
		yearJ, _ := strconv.Atoi(partsJ[1])
		if yearI != yearJ {
			return yearI > yearJ // newest year first (if the years are different, whichever is larger should come first)
		}

		// compare month
		// make sure month is lowercase for lookup in the monthOrder map
		monthI := monthOrder[strings.ToLower(partsI[0])]
		monthJ := monthOrder[strings.ToLower(partsJ[0])]
		return monthI > monthJ // if the years are the same, compare months, larger month number comes earlier in the list - sep (9) comes before aug (8)
	})

	return months, nil
}

func getYearsWithTransactions() (years []string, err error) {
	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return years, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	for y := range transactions {
		years = append(years, y)
	}

	sort.Slice(years, func(i, j int) bool {
		yearI, _ := strconv.Atoi(years[i])
		yearJ, _ := strconv.Atoi(years[j])

		return yearI > yearJ
	})

	return years, nil
}

// helper to get a list of months for a specific year that have transactions
func getMonthsForYear(year string) (months []string, err error) {
	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return months, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	if _, exists := transactions[year]; !exists {
		return months, fmt.Errorf("no transactions for year %s", year)
	}

	for m := range transactions[year] {
		months = append(months, m)
	}

	// sort months by monthOrder (newest first)
	sort.Slice(months, func(i, j int) bool {
		monthI := monthOrder[strings.ToLower(months[i])]
		monthJ := monthOrder[strings.ToLower(months[j])]
		return monthI > monthJ
	})

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
