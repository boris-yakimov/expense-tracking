package main

import (
	"encoding/json"
	"os"
)

type Expense struct {
	Year     string  `json:"year"`
	Month    string  `json:"month"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Note     string  `json:"note"`
}

// minimal expense without year and date
type ExpenseDetails struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Note     string  `json:"note"`
}

type NestedExpenses map[string]map[string][]ExpenseDetails

func loadExpenses() (NestedExpenses, error) {
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

func saveExpenses(nested NestedExpenses) error {
	file, err := os.Create("data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(nested)
}
