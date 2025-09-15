package main

import (
	"fmt"

	"github.com/rivo/tview"
)

// creates a TUI form with required fiields to delete an existing transaction
func formDeleteTransaction(transactionId, transactionType string) error {
	tx, err := getTransactionById(transactionId)
	if err != nil {
		return fmt.Errorf("could not get transaction by id %s: %w", transactionId, err)
	}

	var frame *tview.Frame
	var modal *tview.Modal

	txDetails := fmt.Sprintf("ID %s | Amount %.2f | Category %s | Description %s", tx.Id, tx.Amount, tx.Category, tx.Description)

	modal = styleModal(tview.NewModal().
		SetText(fmt.Sprintf("Deleting transaction \n\n%s", txDetails)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "Delete":
				if err := handleDeleteTransaction(transactionType, transactionId); err != nil {
					showErrorModal(fmt.Sprintf("failed to delete transaction:\n\n%s", err), frame, modal)
					return
				}
				gridVisualizeTransactions("", "") // go back to list of transactions
			case "Cancel":
				gridVisualizeTransactions("", "") // go back to list of transactions
			}
		}))

	modal.SetTitle("Confirm deletion")

	// navigation help
	frame = tview.NewFrame(modal).
		AddText(generateControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// back to mainMenu on ESC or q key press
	modal.SetInputCapture(exitShortcuts)

	tui.SetRoot(frame, true).SetFocus(modal)
	return nil
}

// handles deleting an existing transaction to storage (db or json)
func handleDeleteTransaction(transactionType, transactionId string) error {
	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	txType, err := normalizeTransactionType(transactionType)
	if err != nil {
		return fmt.Errorf("transaction type error: %w", err)
	}

	if len(transactionId) != TransactionIDLength {
		return fmt.Errorf("invalid transaction id length, expected %v char id, got %v", TransactionIDLength, len(transactionId))
	}

	for year, months := range transactions {

		for month := range months {

			var txList = transactions[year][month][txType]
			for i, t := range txList {
				if t.Id == transactionId {
					transactions[year][month][txType] = removeTransactionAtIndex(txList, i)

					if saveTransactionErr := SaveTransactions(transactions); saveTransactionErr != nil {
						return fmt.Errorf("error saving transaction: %w", saveTransactionErr)
					}

					return nil
				}
			}
		}
	}

	return fmt.Errorf("\ndid not match any transaction by id %s, please run list %s or show-total and confirm the transaction id that you want to delete\n", transactionId, txType)
}

// helper to remove a transaction at a specific index
func removeTransactionAtIndex(transactions []Transaction, index int) []Transaction {
	if index < 0 || index >= len(transactions) {
		return transactions // index out of range return original
	}
	return append(transactions[:index], transactions[index+1:]...)
}
