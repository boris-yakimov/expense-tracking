package main

import (
	"fmt"
	"strconv"

	"github.com/rivo/tview"
)

const (
	descriptionMaxCharLength = 40
)

func formAddTransaction() error {
	allowedTransactionTypes, err := listOfAllowedTransactionTypes()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	var transactionType string
	var category string
	var categoryDropdown *tview.DropDown

	typeDropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction Type").
		SetOptions(allowedTransactionTypes, func(selectedOption string, index int) {
			transactionType = selectedOption
			if categoryDropdown != nil {
				opts, err := listOfAllowedCategories(transactionType)
				if err != nil {
					fmt.Println(err)
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
			return fmt.Errorf("%w", err)
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

	var period string
	periodDropdown := styleDropdown(tview.NewDropDown().
		SetLabel("Month/Year"))
	{
		opts, err := getMonthsWithTransactions()
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		periodDropdown.SetOptions(opts, func(selectedOption string, index int) {
			period = selectedOption
		})

		// j/k navigation inside dropdown
		periodDropdown.SetInputCapture(vimNavigation)
	}
	periodDropdown.SetCurrentOption(0)

	// TODO: this seems like a messy approach to get month & year from what was selected in the dropdown, there should be a better way
	month := period[:len(period)-5]
	year := period[len(period)-4:]

	form := styleForm(tview.NewForm().
		AddFormItem(typeDropdown).
		AddFormItem(amountField).
		AddFormItem(categoryDropdown).
		AddFormItem(descriptionField).
		AddFormItem(periodDropdown).
		AddButton("Add", func() {
			amount := amountField.GetText()

			description := descriptionField.GetText()

			if err := handleAddTransaction(transactionType, amount, category, description, month, year); err != nil {
				// TODO: figure out how to better handle these errors
				fmt.Printf("failed to add transaction: %s", err)
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
	frame := tview.NewFrame(form).
		AddText(generateControlsFooter(), false, tview.AlignCenter, theme.FieldTextColor)

	// back to mainMenu on ESC or q key press
	form.SetInputCapture(exitShortcuts)

	tui.SetRoot(frame, true).SetFocus(form)
	return nil
}

func handleAddTransaction(transactionType, amount, category, description, month, year string) error {
	txType, err := normalizeTransactionType(transactionType)
	if err != nil {
		return fmt.Errorf("transaction type error: %w", err)
	}

	txAmount, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("\ninvalid amount: %w\n", err)
	}

	updatedCategory := category
	if _, ok := allowedTransactionCategories[txType][updatedCategory]; !ok {
		fmt.Printf("\ninvalid transaction category: \"%s\"", updatedCategory)
		return fmt.Errorf("\n\nplease pick a valid transaction category from the list above.")
	}

	if !validDescriptionInputFormat(description) {
		return fmt.Errorf("\ninvalid character in description, should contain only letters, numbers, spaces, commas, or dashes")
	}

	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	var transactionId string
	if transactionId, err = generateTransactionId(); err != nil {
		return fmt.Errorf("unable to generate transaction id: %w", err)
	}

	// make sure nested structure exists
	if _, ok := transactions[year]; !ok {
		transactions[year] = make(map[string]map[string][]Transaction)
	}

	if _, ok := transactions[year][month]; !ok {
		transactions[year][month] = make(map[string][]Transaction)
	}

	if _, ok := transactions[year][month][txType]; !ok {
		transactions[year][month][txType] = []Transaction{}
	}

	// make sure only unique IDs are used
	for {
		var duplicateIdFound bool
		for txType := range transactions[year][month] {
			for _, t := range transactions[year][month][txType] {
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

	if len(transactionId) > 8 {
		return fmt.Errorf("transcation id should have a maximum of 8 chars, current id %s with length of %v", transactionId, len(transactionId))
	}

	newTransaction := Transaction{
		Id:          transactionId,
		Amount:      txAmount,
		Category:    category,
		Description: description,
	}

	transactions[year][month][txType] = append(transactions[year][month][txType], newTransaction)
	if saveTransactionErr := saveTransactions(transactions); saveTransactionErr != nil {
		return fmt.Errorf("Error saving transaction: %w", saveTransactionErr)
	}

	fmt.Printf("\n successfully added %s €%.2f | %s | %s\n", txType, txAmount, category, description)
	return nil
}
