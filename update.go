package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	noteMaxLength = 42
)

func addTransaction(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: add <transcation typ> <amount> <category> <note>")
	}

	transactionType := args[0]
	if _, ok := validTranscationTypes[transactionType]; !ok {
		return fmt.Errorf("invalid transaction type %s, please use expense, income, or investment", transactionType)
	}

	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		// TODO: simulate this and validate the output
		return fmt.Errorf("invalid amount: %v", err)
	}

	category := args[2]
	// TODO: fix this so it works with other transaction types as well
	if _, ok := allowedExpenseCategories[category]; !ok {
		fmt.Printf("\ninvalid transaction category: \"%s\"", category)
		showAllowedCategories(transactionType) // expense, income, investment
		return fmt.Errorf("\n\nPlease pick a valid transaction category from the list above.")
	}

	note := strings.Join(args[3:], " ")
	if len(note) > noteMaxLength {
		return fmt.Errorf("\nnote should be a maximum of %v characters, provided %v", noteMaxLength, len(note))
	}

	if !validNoteInputFormat(note) {
		return fmt.Errorf("\ninvalid character in note, notes should contain only letters, numbers, spaces, commas, or dashes")
	}

	return handleTransactionAdd(transactionType, amount, category, note)
}

func handleTransactionAdd(transactionType string, amount float64, category, note string) error {
	transcations, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return fmt.Errorf("Unable to load transactions file: %s", loadFileErr)
	}

	if transactionType == "expense" || transactionType == "expenses" {
		transactionType = "Expenses"
	}

	if transactionType == "investment" || transactionType == "investments" {
		transactionType = "Expenses"
	}

	if transactionType == "income" {
		transactionType = "Income"
	}

	year := strconv.Itoa(time.Now().Year())
	month := time.Now().Month().String()

	// ensure nested structure exists
	if _, ok := transcations[year]; !ok {
		transcations[year] = make(map[string]map[string][]Transaction)
	}

	if _, ok := transcations[year][month]; !ok {
		transcations[year][month] = make(map[string][]Transaction)
	}

	if _, ok := transcations[year][month][transactionType]; !ok {
		transcations[year][month][transactionType] = []Transaction{}
	}

	newTransaction := Transaction{
		Amount:   amount,
		Category: category,
		Note:     note,
	}

	transcations[year][month][transactionType] = append(transcations[year][month][transactionType], newTransaction)
	if saveTransactionErr := saveTransactions(transcations); saveTransactionErr != nil {
		return fmt.Errorf("Error saving transaction: %s", saveTransactionErr)
	}

	fmt.Printf("\nadded $%.2f | %s | %s\n", amount, category, note)

	// TODO: figure out a better way to define cli callback funcs to avoid just passing aroungs args even in places where they are not mandatory
	var args []string
	showTotal(args)

	return nil
}
