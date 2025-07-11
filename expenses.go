package main

import (
	"encoding/json"
	"os"
)

type Expense struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Note     string  `json:"note"`
}

func loadExpenses() ([]Expense, error) {
	file, err := os.Open("data.json")
	if os.IsNotExist(err) {
		return []Expense{}, nil
	}
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var expenses []Expense
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&expenses)
	return expenses, err
}

func saveExpenses(expenses []Expense) error {
	file, err := os.Create("data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(expenses)
}
