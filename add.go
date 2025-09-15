package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/rivo/tview"
)

type AddTransactionRequest struct {
	Type        string
	Amount      string
	Category    string
	Description string
	Month       string
	Year        string
}

// TODO: evaluate if category should not become = "income", "expenses", investments
//       and transaction type should not be     = "food", "groceries", "insurance"
// come to think of it, it sounds more adequate for categories to be the larger thing and types to be a subset of category rather than the other way around, like it is now

// TODO: pressing a when expenses are selected still defaults to income being the selected option from the dropdown menu
// need to make sure that on whichever table we press a we will get this is the selected option in tx Types

// creates a TUI form with required fiields to add a new transaction
func formAddTransaction(currentTableType string) error {
	var transactionType string = currentTableType // prefill with currently selected table
	var category string
	var categoryDropdown *tview.DropDown

	var form *tview.Form
	var frame *tview.Frame

	allowedTransactionTypes, err := listOfAllowedTransactionTypes()
	if err != nil {
		showErrorModal(fmt.Sprintf("list allowed transaction types: %s, err:\n\n%s", transactionType, err), frame, form)
		return err
	}

	typeDropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction Type").
		SetOptions(allowedTransactionTypes, func(selectedOption string, index int) {
			transactionType = selectedOption
			if categoryDropdown != nil {
				opts, err := listOfAllowedCategories(transactionType)
				if err != nil {
					// Set empty options on error to prevent crashes
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
	typeDropdown.SetInputCapture(vimMotions)

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
			showErrorModal(fmt.Sprintf("list allowed categories for transaction type: %s, err:\n\n%s", transactionType, err), frame, form)
			return err
		}
		categoryDropdown.SetOptions(opts, func(selectedOption string, index int) {
			category = selectedOption
		})

		// j/k navigation inside dropdown
		categoryDropdown.SetInputCapture(vimMotions)
	}
	categoryDropdown.SetCurrentOption(0)

	descriptionField := styleInputField(tview.NewInputField().
		SetLabel("Description").
		SetAcceptanceFunc(enforceCharLimit),
	)

	var monthAndYear string
	periodDropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Month/Year"))
	{
		opts, err := getMonthsWithTransactions()
		if err != nil {
			showErrorModal(fmt.Sprintf("unable to get months with transactions: err:\n\n%s", err), frame, form)
			return err
		}
		periodDropdown.SetOptions(opts, func(selectedOption string, index int) {
			monthAndYear = selectedOption
		})

		// j/k navigation inside dropdown
		periodDropdown.SetInputCapture(vimMotions)
	}
	periodDropdown.SetCurrentOption(0)

	// parse the selected month and year
	parts := strings.SplitN(monthAndYear, " ", 2)
	if len(parts) != 2 {
		showErrorModal(fmt.Sprintf("invalid period format: %s", monthAndYear), frame, form)
		return fmt.Errorf("invalid month or year %s", monthAndYear)
	}
	month := parts[0]
	year := parts[1]

	form = styleForm(tview.NewForm().
		AddFormItem(typeDropdown).
		AddFormItem(amountField).
		AddFormItem(categoryDropdown).
		AddFormItem(descriptionField).
		AddFormItem(periodDropdown).
		AddButton("Add", func() {
			amount := amountField.GetText()

			description := descriptionField.GetText()

			var addReq = AddTransactionRequest{
				Type:        transactionType,
				Amount:      amount,
				Category:    category,
				Description: description,
				Month:       month,
				Year:        year,
			}

			if err := handleAddTransaction(addReq); err != nil {
				showErrorModal(fmt.Sprintf("failed to add transaction:\n\n%s", err), frame, form)
				return
			}

			gridVisualizeTransactions("", "") // go back to list of transactions
		}).
		AddButton("Clear", func() {
			typeDropdown.SetCurrentOption(0)
			amountField.SetText("")
			categoryDropdown.SetCurrentOption(0)
			descriptionField.SetText("")
			transactionType = "expense"
		}).
		AddButton("Cancel", func() {
			gridVisualizeTransactions("", "") // go back to list of transactions
		}))

	form.SetBorder(true).SetTitle("Add Transaction").SetTitleAlign(tview.AlignCenter)

	// navigation help
	frame = tview.NewFrame(form).
		AddText(generateControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// back to mainMenu on ESC or q key press
	form.SetInputCapture(exitShortcuts)

	// center the modal
	modal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(frame, 60, 1, true). // width fixed
		AddItem(nil, 0, 1, false))

	centeredModal := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(modal, 20, 1, true). // enough to fit all the fields of the form on the screen
		AddItem(nil, 0, 1, false))

	tui.SetRoot(centeredModal, true).SetFocus(form)
	// tui.SetRoot(frame, true).SetFocus(form)
	return nil
}

// handles adding a new transaction to storage (db or json)
func handleAddTransaction(req AddTransactionRequest) error {
	txType, err := normalizeTransactionType(req.Type)
	if err != nil {
		return fmt.Errorf("transaction type error: %w", err)
	}

	txAmount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		return fmt.Errorf("\ninvalid amount: %w\n", err)
	}

	updatedCategory := req.Category
	if _, ok := allowedTransactionCategories[txType][updatedCategory]; !ok {
		return fmt.Errorf("invalid transaction category: %s", updatedCategory)
	}

	if !validDescriptionInputFormat(req.Description) {
		return fmt.Errorf("invalid character in description, should contain only letters, numbers, spaces, commas, or dashes: %s", req.Description)
	}

	transactions, loadFileErr := LoadTransactions()
	if loadFileErr != nil {
		return fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	var transactionId string
	if transactionId, err = generateTransactionId(); err != nil {
		return fmt.Errorf("unable to generate transaction id: %w", err)
	}

	// make sure nested structure exists
	if _, ok := transactions[req.Year]; !ok {
		transactions[req.Year] = make(map[string]map[string][]Transaction)
	}

	if _, ok := transactions[req.Year][req.Month]; !ok {
		transactions[req.Year][req.Month] = make(map[string][]Transaction)
	}

	if _, ok := transactions[req.Year][req.Month][txType]; !ok {
		transactions[req.Year][req.Month][txType] = []Transaction{}
	}

	// make sure only unique IDs are used
	for {
		var duplicateIdFound bool
		for txType := range transactions[req.Year][req.Month] {
			for _, t := range transactions[req.Year][req.Month][txType] {
				if transactionId == t.Id {
					duplicateIdFound = true
					break
				}
			}
			if duplicateIdFound {
				break // id is already in use
			}
		}

		if !duplicateIdFound {
			break // id is unique
		}

		if transactionId, err = generateTransactionId(); err != nil {
			return fmt.Errorf("unable to generate transaction id: %w", err)
		}
	}

	if len(transactionId) > TransactionIDLength {
		return fmt.Errorf("transcation id should have a maximum of %v chars, current id %s with length of %v", TransactionIDLength, transactionId, len(transactionId))
	}

	newTransaction := Transaction{
		Id:          transactionId,
		Amount:      txAmount,
		Category:    req.Category,
		Description: req.Description,
	}

	transactions[req.Year][req.Month][txType] = append(transactions[req.Year][req.Month][txType], newTransaction)
	if saveTransactionErr := SaveTransactions(transactions); saveTransactionErr != nil {
		return fmt.Errorf("Error saving transaction: %w", saveTransactionErr)
	}

	return nil
}

// creates a randomly generated transaction id that will be assined on each new transaction
func generateTransactionId() (id string, err error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err = rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("error generating transaction id: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
