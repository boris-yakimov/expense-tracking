package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"crypto/rand"
	"encoding/hex"
	"text/tabwriter"

	"github.com/gdamore/tcell/v2"
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

func cleanTerminalInput(cmdArgs string) []string {
	var sanitizedText = strings.Trim(strings.ToLower(cmdArgs), " ")
	return strings.Split(sanitizedText, " ")
}

func validDescriptionInputFormat(description string) bool {
	// only letters, numbers, commas, spaces or dashes
	pattern := `^[a-zA-Z0-9,' '-]+$`
	matched, err := regexp.MatchString(pattern, description)
	if err != nil {
		return false
	}

	return matched
}

func normalizeTransactionType(t string) (string, error) {
	switch t {

	case "expense", "expenses", "Expenses", "Expense":
		return "expense", nil

	case "investment", "investments", "Investments", "Investment":
		return "investment", nil

	case "income", "Income":
		return "income", nil

	default:
		return "", fmt.Errorf("\ninvalid transaction type %s - supported transactions types are income, expense, and investment", t)
	}
}

func generateTransactionId() (id string, err error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err = rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("error generating transaction id: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

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

func listOfAllowedCategories(transactionType string) (categories []string, err error) {
	for c := range allowedTransactionCategories[transactionType] {
		categories = append(categories, c)
	}

	if len(categories) <= 0 {
		return categories, fmt.Errorf("something went wrong with getting list of allowed categories for transaction type %s", transactionType)
	}

	return categories, nil
}

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

func getListOfDetailedTransactions() (listOfTransactions []string, err error) {
	transactions, loadFileErr := loadTransactions()
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

func getTransactionTypeById(txId string) (txType string, err error) {
	transactions, loadFileErr := loadTransactions()
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

func vimNavigation(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j': // move down
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k': // move up
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
	}
	return event
}

func exitShortcuts(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
		mainMenu()
		return nil
	}
	return event
}

func enforceCharLimit(textToCheck string, lastChar rune) bool {
	return len(textToCheck) <= descriptionMaxCharLength
}

func generateControlsFooter() string {
	return "[yellow]ESC[-]/[yellow]q[-]: back   [green]TAB[-]: next   [cyan]j/k[-] or [cyan]↑/↓[-]: navigate"
}

func listTransactionsByMonth(transactionType, month, year string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
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

func showAllowedCategories(transactionType string) error {
	fmt.Println("\nallowed categories are: ")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\nCategory\tDescription")
	fmt.Fprintln(w, "--------\t-----------")

	txType, err := normalizeTransactionType(transactionType)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	for key, val := range allowedTransactionCategories[txType] {
		fmt.Fprintf(w, "%s\t%s\n", key, val)
	}
	w.Flush()
	return nil
}

// make sure these are sorted with newest to oldest month/year
func getMonthsWithTransactions() (months []string, err error) {
	transactions, loadFileErr := loadTransactions()
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
