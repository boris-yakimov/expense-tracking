package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
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
		"housing":        "rent, mortgage, etc",
		"taxes":          "property, capital gains tax, personal income tax, etc (excluding anything related to car)",
		"renovation":     "construction, renovations, home improvements (structural/contractor work)",
		"education":      "courses, certificates, books for learning, tuition",
		"kids":           "daycare, school fees, baby supplies",
		"pets":           "vet, pet food, toys, etc",
		"donations":      "charity, crowdfunding support",
		"fees":           "bank fees, late fees, penalties, subscriptions that don't fall under entertainment",
		"services":       "cleaners, repairs, movers, consultants, etc",
		"cash":           "money withdrawn from ATM and harder to track down under the separate categories, can just be expensed together under this category",
		"sports":         "gym, tennis, swimming ,etc",
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

// color (tview tag) used for a transaction type across the TUI, matching the year overview palette
func txTypeColor(txType string) string {
	switch txType {
	case "income":
		return incomeColor
	case "expense":
		return expenseColor
	case "investment":
		return investmentColor
	}
	return Reset
}

// transactions of a type that match the search filter (case-insensitive, any field); no filter returns all of them
func filterTransactions(txList []Transaction, filter string) []Transaction {
	if filter == "" {
		return txList
	}
	filterLower := strings.ToLower(filter)
	var filtered []Transaction
	for _, tx := range txList {
		if strings.Contains(strings.ToLower(tx.Id), filterLower) ||
			strings.Contains(strings.ToLower(fmt.Sprintf("%.2f", tx.Amount)), filterLower) ||
			strings.Contains(strings.ToLower(tx.Category), filterLower) ||
			strings.Contains(strings.ToLower(tx.Description), filterLower) {
			filtered = append(filtered, tx)
		}
	}
	return filtered
}

// clears and fills a transactions table (header, rows and a title with count and total) and returns the rows shown
// every cell of a row carries the transaction id as its reference, used by update and delete
func populateTransactionsTable(table *tview.Table, txType, month, year string, transactions TransactionHistory, filter string) []Transaction {
	table.Clear()
	table.SetSelectable(true, false). // enable row selection
						SetFixed(1, 0) // make header row fixed
	styleTable(table)
	table.SetBackgroundColor(theme.FieldBackgroundColor) // panel look, same as the breakdown and year summary pages
	table.SetBorder(true)

	color := txTypeColor(txType)
	typeColor := tagColor(color)
	headerColor := tagColor("[#56b6c2]")
	muted := tagColor(mutedColor)

	// empty states replace the header row, a long message in the first column would otherwise stretch it
	message := func(text string) {
		table.SetCell(0, 0, tview.NewTableCell(text).SetTextColor(muted).SetSelectable(false))
	}

	var txList []Transaction
	if year != "" && month != "" {
		txList = transactions[year][month][txType]
	}
	filtered := filterTransactions(txList, filter)

	var total float64
	for _, tx := range filtered {
		total += tx.Amount
	}
	count := fmt.Sprintf("%d", len(txList))
	if filter != "" {
		count = fmt.Sprintf("%d of %d", len(filtered), len(txList))
	}
	table.SetTitle(fmt.Sprintf(" %s%s%s  %s%s ·%s %s€%.2f%s ",
		color, capitalize(txType), Reset, mutedColor, count, Reset, color, total, Reset))

	switch {
	case len(txList) == 0:
		message("no transactions")
		return nil
	case len(filtered) == 0:
		message(fmt.Sprintf("no matches for \"%s\"", tview.Escape(filter)))
		return nil
	}

	headers := []string{"ID", "Amount €", "Category", "Description"}
	for c, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(headerColor).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		if c == 1 {
			cell.SetAlign(tview.AlignRight)
		}
		table.SetCell(0, c, cell)
	}

	for r, tx := range filtered {
		cells := []*tview.TableCell{
			tview.NewTableCell(tx.Id + "  ").SetTextColor(muted),
			tview.NewTableCell(fmt.Sprintf("%.2f", tx.Amount)).SetTextColor(typeColor).SetAlign(tview.AlignRight),
			tview.NewTableCell(" " + tview.Escape(capitalize(tx.Category))).SetTextColor(theme.LabelColor),
			tview.NewTableCell(" " + tview.Escape(tx.Description)).SetTextColor(theme.FieldTextColor).SetExpansion(1),
		}
		for c, cell := range cells {
			fg, _, _ := cell.Style.Decompose()
			cell.SetReference(tx.Id). // used to match specific transaction IDs during update and delete operations
							SetSelectedStyle(tcell.StyleDefault.Foreground(fg).Background(selectedRowColor).Bold(true))
			table.SetCell(r+1, c, cell)
		}
	}
	return filtered
}

// helper to build a table for a specific transaction type for visualization in the TUI
// with a search filter only transactions matching it in any field are shown (vim like search)
func createTransactionsTable(txType, month, year string, transactions TransactionHistory, filter string) *tview.Table {
	table := tview.NewTable()
	populateTransactionsTable(table, txType, month, year, transactions, filter)

	// make sure selection always starts on the first row
	if table.GetRowCount() > 1 {
		table.Select(1, 0)
	}
	return table
}

// helper to update an existing table with filtered transactions, keeping the selected transaction selected if still shown
func updateTransactionsTable(table *tview.Table, txType, month, year string, transactions TransactionHistory, filter string) {
	var selectedTxId string
	if row, col := table.GetSelection(); row > 0 && col >= 0 {
		if cell := table.GetCell(row, col); cell != nil {
			selectedTxId, _ = cell.GetReference().(string)
		}
	}

	filtered := populateTransactionsTable(table, txType, month, year, transactions, filter)
	if len(filtered) == 0 {
		return
	}

	selectedRow := 1 // default to first
	for r, tx := range filtered {
		if tx.Id == selectedTxId {
			selectedRow = r + 1
			break
		}
	}
	table.Select(selectedRow, 0)
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
