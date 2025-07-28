package main

import (
	"encoding/json"
	"os"
)

// TODO: auto complete or show list of allowed expense categories

// format is category: description of category
var allowedExpenseCategories = map[string]string{
	"bills":          "utilities",
	"car":            "any expense around car ownership - insurance, gas, lease, etc",
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
}

var allowedIncomeCategories = map[string]string{
	"salary":          "any income from employer - includes wages, on-call overtime, business trips",
	"transfers":       "transfer in from other people - split bills, family support, etc",
	"dividends":       "stocks, mutual funds, private equity",
	"capital gains":   "sale of stocks, bonds, real estate",
	"rentals":         "real estate, property, equipment",
	"interest":        "savings accounts, bonds, loans and other interest-bearing investments",
	"self employment": "contractor work, gig economy, freelancing",
	"insurance":       "insurance claims",
	"refunds":         "tax refunds, product returns",
}

var allowedInvestmentCategories = map[string]string{
	"stocks":         "direct stock ownership in public comapnies",
	"bonds":          "government, corporrate, etc",
	"funds":          "ETFs or mutual funds",
	"insurance":      "only insurance with an investment element (such as a fund that buys assets)",
	"private equity": "direct ownership in private companies",
	"real estate":    "property",
	"deposits":       "certificate of deposit (CD)",
	"retirement":     "retirement fund contributions",
	// less likely to be used"
	"p2p":         "peer-to-peer lending",
	"crypto":      "bitcoin, ethereum, etc",
	"forex":       "oreign currency investments",
	"options":     "stock options",
	"commodities": "gold, silver, oil, etc",
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

// nestest structrure: year -> month -> transcation type (expense, income, investment) -> transaction
type NestedExpenses map[string]map[string]map[string][]Transaction

func loadTransactions() (NestedExpenses, error) {
	file, err := os.Open("data.json")
	if os.IsNotExist(err) {
		return make(NestedExpenses), nil
	}
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var nested NestedExpenses
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&nested)
	return nested, err
}

func saveTransactions(nested NestedExpenses) error {
	file, err := os.Create("data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(nested)
}
