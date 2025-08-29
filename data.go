package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

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
	table := tview.NewTable()
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

	for r, tx := range txList {
		table.SetCell(r+1, 0, tview.NewTableCell(fmt.Sprintf("%s    ", tx.Id)))
		table.SetCell(r+1, 1, tview.NewTableCell(fmt.Sprintf("€%.2f", tx.Amount)))
		table.SetCell(r+1, 2, tview.NewTableCell(tx.Category))
		table.SetCell(r+1, 3, tview.NewTableCell(tx.Description))
	}
	return table
}

// year -> month -> transcation type (expense, income, or investment) -> transaction
type TransactionHistory map[string]map[string]map[string][]Transaction

var transactionsFilePath = "data.json"

func loadTransactionsFromJsonFile() (TransactionHistory, error) {
	file, err := os.Open(transactionsFilePath)
	if os.IsNotExist(err) {
		return make(TransactionHistory), nil
	}
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var transactions TransactionHistory
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&transactions)
	return transactions, err
}

func saveTransactionsToJsonFile(transactions TransactionHistory) error {
	file, err := os.Create(transactionsFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(transactions)
}

func loadTransactionsFromDb() (TransactionHistory, error) {
	rows, err := db.Query(`
			SELECT id, amount, type, category, description, year, month
			FROM transactions
		`)
	if err != nil {
		return nil, fmt.Errorf("failed to execute load transactions sql query: %w", err)
	}
	defer rows.Close()

	transactions := make(TransactionHistory)

	for rows.Next() {
		var (
			id, txType, category, description string
			amount                            float64
			year, month                       int
		)

		if err := rows.Scan(&id, &amount, &txType, &category, &description, &year, &month); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		y := fmt.Sprintf("%d", year)
		m := fmt.Sprintf("%02d", month)

		if _, ok := transactions[y]; !ok {
			transactions[y] = make(map[string]map[string][]Transaction)
		}
		if _, ok := transactions[y][m]; !ok {
			transactions[y][m] = make(map[string][]Transaction)
		}

		transactions[y][m][txType] = append(transactions[y][m][txType], Transaction{
			Id:          id,
			Amount:      amount,
			Category:    category,
			Description: description,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed during transaction loading: %w", err)
	}

	return transactions, nil
}

func saveTransactionsToDb(transactions TransactionHistory) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin save transaction failed: %w", err)
	}

	// Clear existing data first
	_, err = tx.Exec("DELETE FROM transactions")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to clear transactions: %w", err)
	}

	sqlStatement, err := tx.Prepare(`
			INSERT INTO transactions
			(id, amount, type, category, description, year, month)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare insert during save transaction failed: %w", err)
	}
	defer sqlStatement.Close()

	for year, months := range transactions {
		y, err := strconv.Atoi(year)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("invalid year key %q: %w", year, err)
		}

		for month, types := range months {
			m, err := strconv.Atoi(month)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("invalid month key %q: %w", month, err)
			}

			for txType, list := range types {
				for _, tr := range list {
					_, err = sqlStatement.Exec(
						tr.Id,
						tr.Amount,
						txType,
						tr.Category,
						tr.Description,
						y,
						m,
					)
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("insert failed for transaction %s: %w", tr.Id, err)
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}
