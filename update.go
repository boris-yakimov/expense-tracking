package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/rivo/tview"
)

type UpdateTransactionRequest struct {
	Type        string
	Id          string
	Amount      string
	Category    string
	Description string
}

// creates a TUI form with required fields to update an existing transaction
func formUpdateTransaction(transactionId, transactionType, selectedMonth, selectedYear string) error {
	// fetch transaction data by id
	tx, err := getTransactionById(transactionId)
	if err != nil {
		return fmt.Errorf("could not get transaction by id %s: %w", transactionId, err)
	}

	var form *tview.Form
	var frame *tview.Frame

	// transaction type dropdown (pre-populated with currently selected type)
	allowedTransactionTypes, err := listOfAllowedTransactionTypes()
	if err != nil {
		showErrorModal(fmt.Sprintf("get a list of allowed transaction types err:\n\n%s", err), frame, form)
		log.Printf("get a list of allowed transaction types err:\n\n%s", err)
	}

	typeDropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction Type").
		SetOptions(allowedTransactionTypes, func(selectedOption string, index int) {
			transactionType = selectedOption
		}))

	// pre-select transaction type in dropdown
	for i, opt := range allowedTransactionTypes {
		if opt == transactionType {
			typeDropdown.SetCurrentOption(i)
			transactionType = opt
			break
		}
	}

	// j/k navigation inside dropdown
	typeDropdown.SetInputCapture(vimMotions)

	// pre-populated amount from selected transaction
	amountField := styleInputField(tview.NewInputField().
		SetLabel("Amount").
		SetText(fmt.Sprintf("%.2f", tx.Amount)))

	// category dropwon (pre-populated with current category)
	categoryDropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Category"))
	// scope boundary to isolate opts and err from leaking in the rest of the function
	{
		opts, err := listOfAllowedCategories(transactionType)
		if err != nil {
			showErrorModal(fmt.Sprintf("failed to list categories err:\n\n%s", err), frame, form)
			log.Printf("failed to list categories err:\n\n%s", err)
			return err
		}
		categoryDropdown.SetOptions(opts, func(selectedOption string, index int) {
			tx.Category = selectedOption
		})

		for i, opt := range opts {
			if opt == tx.Category {
				categoryDropdown.SetCurrentOption(i)
				break
			}
		}

		// j/k navigation inside dropdown
		categoryDropdown.SetInputCapture(vimMotions)
	}

	// description field (pre-populated with current description)
	descriptionField := styleInputField(tview.NewInputField().
		SetLabel("Description").
		SetText(tx.Description).
		SetAcceptanceFunc(enforceCharLimit),
	)

	form = styleForm(tview.NewForm().
		AddFormItem(typeDropdown).
		AddFormItem(amountField).
		AddFormItem(categoryDropdown).
		AddFormItem(descriptionField).
		AddButton("Update", func() {
			amount := amountField.GetText()
			description := descriptionField.GetText()

			var updateReq = UpdateTransactionRequest{
				Type:        transactionType,
				Id:          transactionId,
				Amount:      amount,
				Category:    tx.Category,
				Description: description,
			}

			if err := handleUpdateTransaction(updateReq); err != nil {
				showErrorModal(fmt.Sprintf("failed to update transaction:\n\n%s", err), frame, form)
				log.Printf("failed to update transaction:\n\n%s", err)
				return
			}

			gridVisualizeTransactions(selectedMonth, selectedYear) // go back to the list of transactions (at the same month and year from where formDeleteTransaction was triggered)

		}).
		AddButton("Clear", func() {
			typeDropdown.SetCurrentOption(0)
			amountField.SetText("")
			categoryDropdown.SetCurrentOption(0)
			descriptionField.SetText("")
		}).
		AddButton("Cancel", func() {
			gridVisualizeTransactions(selectedMonth, selectedYear) // go back to list of transactions
		}))

	form.SetBorder(true).SetTitle("Expense Tracking Tool").SetTitleAlign(tview.AlignCenter)

	// navigation help
	frame = tview.NewFrame(form).
		AddText(generateCombinedControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// back to list of transactions on ESC or q key press
	form.SetInputCapture(exitShortcuts)

	modal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(frame, 60, 1, true). // width fixed
		AddItem(nil, 0, 1, false))

	centeredModal := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(modal, 17, 1, true). // enough to fit all the fields of the form on the screen
		AddItem(nil, 0, 1, false))

	tui.SetRoot(centeredModal, true).SetFocus(form)

	return nil
}

// handles updating an existing transaction in storage
func handleUpdateTransaction(req UpdateTransactionRequest) error {
	txType, err := normalizeTransactionType(req.Type)
	if err != nil {
		return fmt.Errorf("transaction type error: %w", err)
	}

	if len(req.Id) != TransactionIDLength {
		return fmt.Errorf("invalid transaction id length, expected %v char id, got %v", TransactionIDLength, len(req.Id))
	}

	updatedAmount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		return fmt.Errorf("\ninvalid amount: %w\n", err)
	}

	if _, ok := allowedTransactionCategories[txType][req.Category]; !ok {
		return fmt.Errorf("\n\ninvalid transaction category: %s", req.Category)
	}

	if !validDescriptionInputFormat(req.Description) {
		return fmt.Errorf("\ninvalid character in description, should contain only letters, numbers, spaces, commas, or dashes")
	}

	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	// years
	var transactionFound bool
	for year, months := range transactions {

		// months
		for month := range months {

			for i, tx := range transactions[year][month][txType] {
				if tx.Id == req.Id {
					tx.Amount = updatedAmount
					tx.Description = req.Description
					tx.Category = req.Category

					transactions[year][month][txType][i] = tx
					transactionFound = true
				}
			}
		}
	}

	if !transactionFound {
		return fmt.Errorf("transaction with id %s not found", req.Id)
	}

	if saveTransactionErr := SaveTransactions(transactions); saveTransactionErr != nil {
		return fmt.Errorf("error saving transaction: %w", saveTransactionErr)
	}

	return nil
}
