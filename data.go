package main

import (
	"fmt"

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
	Id          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
}

// helper to build a table for a specific transaction type for visualization in the TUI
func createTransactionsTable(txType, month, year string, transactions TransactionHistory) *tview.Table {
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

	// populate table
	for r, tx := range txList {
		table.SetCell(r+1, 0, tview.NewTableCell(fmt.Sprintf("%s    ", tx.Id)))
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

// load transactions from storage (db or json)
func LoadTransactions() (TransactionHistory, error) {
	if globalConfig == nil {
		globalConfig = DefaultConfig()
	}

	switch globalConfig.StorageType {
	case StorageJSONFile:
		return loadTransactionsFromJsonFile()
	case StorageSQLite:
		return loadTransactionsFromDb()
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", globalConfig.StorageType)
	}
}

// load transactions to storage (db or json)
func SaveTransactions(transactions TransactionHistory) error {
	if globalConfig == nil {
		globalConfig = DefaultConfig()
	}

	switch globalConfig.StorageType {
	case StorageJSONFile:
		return saveTransactionsToJsonFile(transactions)
	case StorageSQLite:
		return saveTransactionsToDb(transactions)
	default:
		return fmt.Errorf("unsupported storage type: %s", globalConfig.StorageType)
	}
}
