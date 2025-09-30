package main

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

var allowedTransactionTypes = map[string]struct{}{
	"income":     {},
	"expense":    {},
	"investment": {},
}

var allowedTransactionCategories = map[string]map[string]string{
	"expense": {
		"bills":          "utilities (usually recurring) - electricity, water, gas, internet, phone, etc",
		"car":            "any expense around car ownership - insurance, fuel, lease, etc",
		"food":           "anything food and drink related, including groceries, coffee stops, etc",
		"entertainment":  "games, books, movies, subscriptions, events",
		"insurance":      "health, property, card (excluding anything related to car) (excluding investment grace insurance policies which should fall under the investments category)",
		"shopping":       "clothes, gifts, personal items, home goods",
		"travel":         "all travel including busines trip expenses",
		"transportation": "anything transportation related excluding personal car expenditures",
		"healthcare":     "hospital, pharmacy, supplements, etc",
		"transfers":      "transfer out to other people - split bills, family support, etc",
		"taxes":          "property, capital gains tax, personal income tax, etc (excluding anything related to car)",
		"renovation":     "construction, renovations, home improvements (structural/contractor work)",
		"education":      "courses, certificates, books for learning, tuition",
		"kids":           "daycare, school fees, baby supplies",
		"pets":           "vet, pet food, toys, etc",
		"donations":      "charity, crowdfunding support",
		"fees":           "bank fees, late fees, penalties, subscriptions that don't fall under entertainment",
		"services":       "cleaners, repairs, movers, consultants, etc",
		"cash":           "money withdrawn from ATM and harder to track down under the separate categories, can just be expensed together under this category",
	},

	"investment": {
		"stocks":        "direct stock ownership in public companies",
		"bonds":         "government, corporate, municipal, etc",
		"funds":         "ETFs or mutual funds",
		"insurance":     "only insurance with an investment element (such as a fund that buys assets)",
		"privateEquity": "direct ownership in private companies",
		"realEstate":    "property",
		"deposits":      "certificate of deposit (CD)",
		"retirement":    "retirement fund contributions",
		"p2p":           "peer-to-peer lending",
		"crypto":        "bitcoin, ethereum, etc",
		"forex":         "foreign currency investments",
		"options":       "stock options",
		"commodities":   "gold, silver, oil, etc",
	},

	"income": {
		"salary":         "any income from employer - includes wages, on-call overtime, business trips",
		"transfers":      "transfer in from other people - split bills, family support, etc",
		"dividends":      "stocks, mutual funds, private equity",
		"capitalGains":   "sale of stocks, bonds, real estate",
		"rentals":        "real estate, property, equipment",
		"interest":       "savings accounts, bonds, loans and other interest-bearing investments",
		"selfEmployment": "contractor work, gig economy, freelancing",
		"insurance":      "insurance claims",
		"refunds":        "tax refunds, product returns",
	},
}

// minimal expense without year and date
type Transaction struct {
	Id          string
	Amount      float64
	Category    string
	Description string
}

// helper to build a table for a specific transaction type for visualization in the TUI
func createTransactionsTable(txType, month, year string, transactions TransactionHistory, filter string) *tview.Table {
	table := tview.NewTable().
		SetSelectable(true, false). // enable row selection
		SetFixed(1, 0)              // make header row fixed
	table.SetBorder(false)
	table.SetTitle(capitalize(txType)).SetBorder(true)

	headers := []string{"ID", "Amount", "Category", "Description"}
	for c, h := range headers {
		table.SetCell(0, c, tview.NewTableCell(h).SetSelectable(false))
	}

	if year == "" || month == "" {
		table.SetCell(1, 0, tview.NewTableCell("no transactions"))
		return table
	}

	txList := transactions[year][month][txType]
	if len(txList) == 0 {
		table.SetCell(1, 0, tview.NewTableCell("no transactions"))
		return table
	}

	var filteredTxList []Transaction
	// if no search filter is provided just use the list of transactions for the selected month
	// this way we show all transactions initially and when we initiate a search we show only trasactions that match the pattern
	if filter == "" {
		filteredTxList = txList
	} else {
		// search through transactions if filter is provided (for vim like search functionality)
		filterLower := strings.ToLower(filter)
		for _, tx := range txList {
			// search for a pattern in any of the sections if present, append to the filtered list
			// filtered list will later be used to show only trasactions that match the search pattern during searching
			if strings.Contains(strings.ToLower(tx.Id), filterLower) ||
				strings.Contains(strings.ToLower(fmt.Sprintf("%.2f", tx.Amount)), filterLower) ||
				strings.Contains(strings.ToLower(tx.Category), filterLower) ||
				strings.Contains(strings.ToLower(tx.Description), filterLower) {
				filteredTxList = append(filteredTxList, tx)
			}
		}
	}

	if len(filteredTxList) == 0 {
		table.SetCell(1, 0, tview.NewTableCell("no transaction matches found"))
		return table
	}

	// populate a table with only the transactions that match the specific pattern that we are searching for
	for r, tx := range filteredTxList {
		table.SetCell(r+1, 0, tview.NewTableCell(fmt.Sprintf("%s    ", tx.Id)).
			SetReference(tx.Id)) // setting a reference for transaction IDs that will later be used when trying to match specific transaction IDs during update and delete operations
		table.SetCell(r+1, 1, tview.NewTableCell(fmt.Sprintf("€%.2f", tx.Amount)))
		table.SetCell(r+1, 2, tview.NewTableCell(tx.Category))
		table.SetCell(r+1, 3, tview.NewTableCell(tx.Description))

	}
	// make sure selection always starts on the first row
	if table.GetRowCount() > 1 {
		table.Select(1, 0)
	}

	return table
}

// year -> month -> transcation type (expense, income, or investment) -> transaction
type TransactionHistory map[string]map[string]map[string][]Transaction

// load transactions from storage
func LoadTransactions() (TransactionHistory, error) {
	var err error
	if globalConfig == nil {
		globalConfig, err = DefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load transactions, err: %w", err)
		}
	}

	// previously also supported JSON but was deprecated, leaving the current approach in case I want to extend with other storage options in the future
	switch globalConfig.StorageType {
	case StorageSQLite:
		return loadTransactionsFromDb()
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", globalConfig.StorageType)
	}
}

// load transactions to storage
func SaveTransactions(transactions TransactionHistory) error {
	var err error
	if globalConfig == nil {
		globalConfig, err = DefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to save transactions, err: %w", err)
		}
	}

	// previously also supported JSON but was deprecated, leaving the current approach in case I want to extend with other storage options in the future
	switch globalConfig.StorageType {
	case StorageSQLite:
		return saveTransactionsToDb(transactions)
	default:
		return fmt.Errorf("unsupported storage type: %s", globalConfig.StorageType)
	}
}
