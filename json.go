package main

import (
	"encoding/json"
	"os"
)

func loadTransactionsFromJsonFile() (TransactionHistory, error) {
	file, err := os.Open(globalConfig.JSONFilePath)
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
	file, err := os.Create(globalConfig.JSONFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(transactions)
}
