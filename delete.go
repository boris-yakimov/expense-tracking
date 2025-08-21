package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func formDeleteTransaction() error {

	var transactionId string
	var transactionType string

	idDropDown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction ID"))

	{
		// TODO: show info what is behind this id
		opts, err := listOfTransactionIds() // TODO: to be implemented
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		idDropDown.SetOptions(opts, func(selectedOption string, index int) {
			transactionId = selectedOption
		})
	}

	// TODO: how to get type from id ?
	transactionType, err := getTxTypeById() // TODO: to be implemented
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// TODO: build out form
	form := styleForm(tview.NewForm().
		AddFormItem(idDropDown).
		AddButton("Delete", func() {
			if _, err := deleteTransaction(transactionType, transactionId); err != nil {
				fmt.Printf("failed to delete transaction: %s", err)
				return
			}
		}).
		AddButton("Cancel", func() {
			mainMenu()
		}))

	form.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)
	// TOOD: q / ESC also goes back to main menu
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// back to mainMenu on ESC or q key press
		if event.Key() == tcell.KeyEsc || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			mainMenu()
			return nil
		}
		return event
	})

	tui.SetRoot(form, true).SetFocus(form)
	return nil
}

// TODO: remove these after TUI approach is implemented

// delete <transaction_type> <transaction_id>
func deleteTransaction(transactionType, transactionId string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	txType, err := normalizeTransactionType(transactionType)
	if err != nil {
		return false, fmt.Errorf("transaction type error: %w", err)
	}

	if len(transactionId) != 8 {
		return false, fmt.Errorf("invalid transaction id length, expected 8 char id, got %s", len(transactionId))
	}

	for year, months := range transactions {

		for month := range months {

			var txList = transactions[year][month][txType]
			for i, t := range txList {
				if t.Id == transactionId {
					transactions[year][month][txType] = removeTransactionAtIndex(txList, i)

					if saveTransactionErr := saveTransactions(transactions); saveTransactionErr != nil {
						return false, fmt.Errorf("error saving transaction: %w", saveTransactionErr)
					}
					fmt.Printf("successfully removed transaction with id %s\n\n", transactionId)

					fmt.Printf("%s for %s %s\n", txType, month, year)
					_, err = listTransactionsByMonth(txType, month, year)
					if err != nil {
						return false, fmt.Errorf("unable to list remaining transactions: %s", err)
					}

					return true, nil
				}
			}
		}
	}

	return false, fmt.Errorf("\ndid not match any transaction by id %s, please run list %s or show-total and confirm the transaction id that you want to delete\n", transactionId, txType)
}

func removeTransactionAtIndex(transactions []Transaction, index int) []Transaction {
	if index < 0 || index >= len(transactions) {
		return transactions // index out of range return original
	}
	return append(transactions[:index], transactions[index+1:]...)
}
