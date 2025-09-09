package main

import (
	"fmt"
	"strconv"
	"strings"

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
func formUpdateTransaction() error {
	var transactionId string
	var transactionType string

	var form *tview.Form
	var frame *tview.Frame

	idDropDown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction To Update"))

	{
		// show detailed transaction information so user knows what they are deleting
		opts, err := getListOfDetailedTransactions()
		if err != nil {
			showErrorModal(fmt.Sprintf("get a list of detailed transactions err:\n\n%s", err), frame, form)
			return err
		}
		idDropDown.SetOptions(opts, func(selectedOption string, index int) {
			// extract ID from the selected option (format: "ID: 12345678 | ...")
			if len(selectedOption) > 4 {
				parts := strings.SplitN(selectedOption, "|", 2)
				if len(parts) > 0 {
					idPart := strings.TrimSpace(parts[0]) // "ID: 12345678"
					transactionId = strings.TrimPrefix(idPart, "ID: ")
				}
			}

			// get transaction type after user selects an ID
			var err error
			transactionType, err = getTransactionTypeById(transactionId)
			if err != nil {
				showErrorModal(fmt.Sprintf("error getting transaction type by id: %s, err:\n\n%s", transactionId, err), frame, form)
				return
			}
		})
		// j/k navigation inside dropdown
		idDropDown.SetInputCapture(vimNavigation)
	}

	var categoryDropdown *tview.DropDown
	var category string

	allowedTransactionTypes, err := listOfAllowedTransactionTypes()
	if err != nil {
		showErrorModal(fmt.Sprintf("get a list of allowed transaction types err:\n\n%s", err), frame, form)
		return err
	}

	typeDropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction Type").
		SetOptions(allowedTransactionTypes, func(selectedOption string, index int) {
			transactionType = selectedOption
			if categoryDropdown != nil {
				opts, err := listOfAllowedCategories(transactionType)
				if err != nil {
					showErrorModal(fmt.Sprintf("list allowed categories for transaction type: %s, err:\n\n%s", transactionType, err), frame, form)
					return
				}
				categoryDropdown.SetOptions(opts, func(selectedOption string, index int) {
					category = selectedOption
				})
				if len(opts) > 0 {
					categoryDropdown.SetCurrentOption(0)
					category = opts[0]
				} else {
					category = ""
				}
			}
		}))
	typeDropdown.SetCurrentOption(0)

	// j/k navigation inside dropdown
	typeDropdown.SetInputCapture(vimNavigation)

	if _, opt := typeDropdown.GetCurrentOption(); opt != "" {
		transactionType = opt
	}

	amountField := styleInputField(tview.NewInputField().SetLabel("Amount"))

	categoryDropdown = styleDropdown(tview.NewDropDown().
		SetLabel("Category"))
	// scope boundary to isolate opts and err from leaking in the rest of the function
	{
		opts, err := listOfAllowedCategories(transactionType)
		if err != nil {
			showErrorModal(fmt.Sprintf("failed to list categories err:\n\n%s", err), frame, form)
			return err
		}
		categoryDropdown.SetOptions(opts, func(selectedOption string, index int) {
			category = selectedOption
		})

		// j/k navigation inside dropdown
		categoryDropdown.SetInputCapture(vimNavigation)
	}
	categoryDropdown.SetCurrentOption(0)

	descriptionField := styleInputField(tview.NewInputField().
		SetLabel("Description").
		SetAcceptanceFunc(enforceCharLimit),
	)

	form = styleForm(tview.NewForm().
		AddFormItem(idDropDown).
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
				Category:    category,
				Description: description,
			}

			if err := handleUpdateTransaction(updateReq); err != nil {
				showErrorModal(fmt.Sprintf("failed to update transaction:\n\n%s", err), frame, form)
				return
			}

			mainMenu() // go back to menu
		}).
		AddButton("Clear", func() {
			typeDropdown.SetCurrentOption(0)
			amountField.SetText("")
			categoryDropdown.SetCurrentOption(0)
			descriptionField.SetText("")
			transactionType = "expense"
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

// handles updating an existing transaction in storage (db or json)
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
	// TODO: seems to appear in the frame next to the helper menu, figure out what is a better place for this to appear in
	fmt.Printf("transaction successully updated")

	return nil
}
