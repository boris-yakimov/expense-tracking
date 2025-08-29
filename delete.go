package main

import (
	"fmt"

	"github.com/rivo/tview"
)

func formDeleteTransaction() error {
	var transactionId string
	var transactionType string

	var frame *tview.Frame
	var form *tview.Form

	idDropDown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction List"))

	{
		// show detailed transaction information so user knows what they are deleting
		opts, err := getListOfDetailedTransactions()
		if err != nil {
			showErrorModal(fmt.Sprintf("get detailed transactions err: \n\n%s", err), frame, form)
			return err
		}
		idDropDown.SetOptions(opts, func(selectedOption string, index int) {
			// extract ID from the selected option (format: "ID: 12345678 | ...")
			if len(selectedOption) > 4 {
				transactionId = selectedOption[4:12] // extract ID from position 4-12
			}
			// get transaction type after user selects an ID
			var err error
			transactionType, err = getTransactionTypeById(transactionId)
			if err != nil {
				showErrorModal(fmt.Sprintf("error getting transaction type:\n\n%s", err), frame, form)
				return
			}
		})

		// j/k navigation inside dropdown
		idDropDown.SetInputCapture(vimNavigation)
	}

	form = styleForm(tview.NewForm().
		AddFormItem(idDropDown).
		AddButton("Delete", func() {
			if err := handleDeleteTransaction(transactionType, transactionId); err != nil {
				showErrorModal(fmt.Sprintf("failed to delete transaction:\n\n%s", err), frame, form)
				return
			}
		}).
		AddButton("Cancel", func() {
			mainMenu()
		}))

	form.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	// navigation help
	frame = tview.NewFrame(form).
		AddText(generateControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// back to mainMenu on ESC or q key press
	form.SetInputCapture(exitShortcuts)

	tui.SetRoot(frame, true).SetFocus(form)
	return nil
}

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
					// TODO: seems to appear in the frame next to the helper menu, figure out what is a better place for this to appear in
					fmt.Printf("successfully removed transaction with id %s\n\n", transactionId)

					fmt.Printf("%s for %s %s\n", txType, month, year)
					_, err = listTransactionsByMonth(txType, month, year)
					if err != nil {
						return fmt.Errorf("unable to list remaining transactions: %w", err)
					}

					return nil
				}
			}
		}
	}

	return fmt.Errorf("\ndid not match any transaction by id %s, please run list %s or show-total and confirm the transaction id that you want to delete\n", transactionId, txType)
}

func removeTransactionAtIndex(transactions []Transaction, index int) []Transaction {
	if index < 0 || index >= len(transactions) {
		return transactions // index out of range return original
	}
	return append(transactions[:index], transactions[index+1:]...)
}
