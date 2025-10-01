package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

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

// creates a TUI form with required fiields to add a new transaction
func formAddTransaction(currentTableType, selectedMonth, selectedYear string) error {
	var transactionType string
	var category string
	var categoryDropdown *tview.DropDown

	var form *tview.Form
	var frame *tview.Frame

	allowedTransactionTypes, err := listOfAllowedTransactionTypes()
	if err != nil {
		showErrorModal(fmt.Sprintf("list allowed transaction types: %s, err:\n\n%s", transactionType, err), frame, form)
		log.Printf("list allowed transaction types: %s, err:\n\n%s", transactionType, err)
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
					log.Printf("list allowed categories for transaction type: %s, err:\n\n%s", transactionType, err)
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

	// set dropdown to match the table you came from
	prefillIndex := 0
	for i, t := range allowedTransactionTypes {
		if t == currentTableType {
			prefillIndex = i
			transactionType = t
			break
		}
	}
	typeDropdown.SetCurrentOption(prefillIndex)

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
			log.Printf("list allowed categories for transaction type: %s, err:\n\n%s", transactionType, err)
			return err
		}
		categoryDropdown.SetOptions(opts, func(selectedOption string, index int) {
			category = selectedOption
		})
	}
	categoryDropdown.SetCurrentOption(0)

	descriptionField := styleInputField(tview.NewInputField().
		SetLabel(fmt.Sprintf("Description (0/%d)", DescriptionMaxCharLength)).
		SetAcceptanceFunc(enforceCharLimit),
	)
	// keep track of characters typed so far and char limit for description
	descriptionField.SetChangedFunc(func(text string) {
		descriptionField.SetLabel(fmt.Sprintf("Description (%d/%d)", len(text), DescriptionMaxCharLength))
	})

	var monthAndYear string
	periodDropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Month/Year"))
	{
		opts, err := getMonthsWithTransactions()
		if err != nil {
			showErrorModal(fmt.Sprintf("unable to get months with transactions: err:\n\n%s", err), frame, form)
			log.Printf("unable to get months with transactions: err:\n\n%s", err)
			return err
		}

		// make sure the current month is in the list
		now := time.Now()
		currentMonth := fmt.Sprintf("%s %s", strings.ToLower(now.Month().String()), strconv.Itoa(now.Year()))
		if !slices.Contains(opts, currentMonth) {
			opts = append(opts, currentMonth)
		}

		// make sure the previous month is also in the list
		prev := now.AddDate(0, -1, 0)
		previousMonth := fmt.Sprintf("%s %s", strings.ToLower(prev.Month().String()), strconv.Itoa(prev.Year()))
		if !slices.Contains(opts, previousMonth) {
			opts = append(opts, previousMonth)
		}

		periodDropdown.SetOptions(opts, func(selectedOption string, index int) {
			monthAndYear = selectedOption
		})

		//pre-select the month that we came from when we initiated formAddTransaction()
		prefillIndex := 0
		for i, period := range opts {
			if period == fmt.Sprintf("%s %s", selectedMonth, selectedYear) {
				prefillIndex = i
				monthAndYear = period
				break
			}
		}

		periodDropdown.SetCurrentOption(prefillIndex)

		// j/k navigation inside dropdown
		periodDropdown.SetInputCapture(vimMotions)
	}

	form = styleForm(tview.NewForm().
		AddFormItem(typeDropdown).
		AddFormItem(amountField).
		AddFormItem(categoryDropdown).
		AddFormItem(descriptionField).
		AddFormItem(periodDropdown).
		AddButton("Add", func() {
			amount := amountField.GetText()
			description := descriptionField.GetText()

			// parse the selected month and year
			parts := strings.SplitN(monthAndYear, " ", 2)
			if len(parts) != 2 {
				showErrorModal(fmt.Sprintf("invalid period format: %s", monthAndYear), frame, form)
				log.Printf("invalid period format: %s", monthAndYear)
				return
			}
			month := parts[0]
			year := parts[1]

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
				log.Printf("failed to add transaction:\n\n%s", err)
				return
			}

			_, err := gridVisualizeTransactions(month, year, transactionType, true) // go back to list of transactions for the same month and table type
			if err != nil {
				showErrorModal("failed to return back to transactions list from add form", frame, form)
				log.Printf("failed to return back to transactions list from add form")
			}
		}).
		AddButton("Clear", func() {
			typeDropdown.SetCurrentOption(0)
			amountField.SetText("")
			categoryDropdown.SetCurrentOption(0)
			descriptionField.SetText("")
			transactionType = "expense"
		}).
		AddButton("Cancel", func() {
			gridVisualizeTransactions(selectedMonth, selectedYear, currentTableType, true) // go back to the list of transactions (at the same month and year from where formDeleteTransaction was triggered)
		}))

	form.SetBorder(true).SetTitle("Add Transaction").SetTitleAlign(tview.AlignCenter)

	// navigation help
	frame = tview.NewFrame(form).
		AddText(generateCombinedControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// back to list of transactions on ESC or q key press
	form.SetInputCapture(exitShortcutsWithPeriod(selectedMonth, selectedYear, currentTableType))

	// center the modal
	modal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(frame, 60, 1, true). // width fixed
		AddItem(nil, 0, 1, false))

	centeredModal := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(modal, 19, 1, true). // enough to fit all the fields of the form on the screen
		AddItem(nil, 0, 1, false))

	tui.SetRoot(centeredModal, true).SetFocus(form)
	return nil
}

// handles adding a new transaction to storage
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
		return fmt.Errorf("invalid character in description, allowed: %s, got: %s", allowedCharsDescription, req.Description)
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
