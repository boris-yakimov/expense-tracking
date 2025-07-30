package main

import (
	"encoding/json"
	"os"
)

// TODO: auto complete or show list of allowed expense categories
var allowedTranscationCategories = map[string]map[string]string{
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
		"stocks":        "direct stock ownership in public comapnies",
		"bonds":         "government, corporrate, etc",
		"funds":         "ETFs or mutual funds",
		"insurance":     "only insurance with an investment element (such as a fund that buys assets)",
		"privateEquity": "direct ownership in private companies",
		"realEstate":    "property",
		"deposits":      "certificate of deposit (CD)",
		"retirement":    "retirement fund contributions",
		"p2p":           "peer-to-peer lending",
		"crypto":        "bitcoin, ethereum, etc",
		"forex":         "oreign currency investments",
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

var validTranscationTypes = map[string]struct{}{
	"expenses":    {},
	"expense":     {},
	"investments": {},
	"investment":  {},
	"income":      {},
}

// minimal expense without year and date
type Transaction struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	// TODO: is description a better term for this ?
	Note string `json:"note"`
}

// structrure year -> month -> transcation type (expense, income, or investment) -> transaction
type TransactionHistory map[string]map[string]map[string][]Transaction

func loadTransactions() (TransactionHistory, error) {
	file, err := os.Open("data.json")
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

func saveTransactions(transactions TransactionHistory) error {
	file, err := os.Create("data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(transactions)
}
