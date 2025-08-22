package main

import (
	"fmt"
	"strconv"
)

func formUpdateTransaction() error {

	return nil
}

func handleUpdateTransaction(transactionType, transactionId, amount, category, description string) error {
	txType, err := normalizeTransactionType(transactionType)
	if err != nil {
		return fmt.Errorf("transaction type error: %w", err)
	}

	if len(transactionId) != 8 {
		return fmt.Errorf("invalid transaction id length, expected 8 char id, got %v", len(transactionId))
	}

	updatedAmount, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("\ninvalid amount: %w\n", err)
	}

	if _, ok := allowedTransactionCategories[txType][category]; !ok {
		fmt.Printf("\ninvalid transaction category: \"%s\"", category)
		return fmt.Errorf("\n\nplease pick a valid transaction category from the list above.")
	}

	if !validDescriptionInputFormat(description) {
		return fmt.Errorf("\ninvalid character in description, should contain only letters, numbers, spaces, commas, or dashes")
	}
	if len(description) > descriptionMaxCharLength {
		return fmt.Errorf("\ndescription should be a maximum of %v characters, provided %v", descriptionMaxCharLength, len(description))
	}

	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	// years
	for year, months := range transactions {

		// months
		for month := range months {

			for i, tx := range transactions[year][month][txType] {
				if tx.Id == transactionId {
					tx.Amount = updatedAmount
					tx.Description = description
					tx.Category = category

					transactions[year][month][txType][i] = tx
				}
			}
		}
	}

	if saveTransactionErr := saveTransactions(transactions); saveTransactionErr != nil {
		return fmt.Errorf("error saving transaction: %w", saveTransactionErr)
	}
	fmt.Printf("transaction successully updated")

	return nil
}
