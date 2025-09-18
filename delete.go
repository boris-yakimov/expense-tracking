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

	txDetails := fmt.Sprintf("ID %s | Amount %.2f | Category %s | Description %s", tx.Id, tx.Amount, tx.Category, tx.Description)

	var form *tview.Form
	var frame *tview.Frame

	form = styleForm(tview.NewForm().
		AddButton("Delete", func() {
			if err := handleDeleteTransaction(transactionType, transactionId); err != nil {
				showErrorModal(fmt.Sprintf("failed to delete transaction:\n\n%s", err), frame, form)
				return
			}
			gridVisualizeTransactions("", "")
		}).
		AddButton("Cancel", func() {
			gridVisualizeTransactions("", "")
		}))

	form.SetTitle("Delete Transaction").
		SetTitleAlign(tview.AlignCenter).
		SetBorder(true)

	detailtsTextView := styleTextView(tview.NewTextView().
		SetText(txDetails).
		SetRegions(true))
	form.AddFormItem(detailtsTextView)

	form.SetButtonsAlign(tview.AlignCenter)

	// navigation help
	frame = tview.NewFrame(form).
		AddText(generateCombinedControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// horizontal centering
	modal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).   // left spacer
		AddItem(frame, 90, 1, true). // form width fixed to fit text
		AddItem(nil, 0, 1, false))   // right spacer

	// vertical centering (height = 0 lets it fit content)
	centeredModal := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).   // top spacer
		AddItem(modal, 15, 1, true). // enough to fit all the fields of the form on the screen
		AddItem(nil, 0, 1, false))   // bottom spacer

	tui.SetRoot(centeredModal, true).SetFocus(form)

	// back to transactions list on ESC or q key press
	form.SetInputCapture(exitShortcuts)

	tui.SetRoot(centeredModal, true).SetFocus(form)
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
