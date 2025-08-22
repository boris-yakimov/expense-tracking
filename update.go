package main

import (
	"fmt"
	"strconv"

	"github.com/rivo/tview"
)

func formUpdateTransaction() error {
	var transactionId string
	var transactionType string

	idDropDown := styleDropdown(tview.NewDropDown().
		SetLabel("Transaction To Update"))

	{
		// show detailed transaction information so user knows what they are deleting
		opts, err := getListOfDetailedTransactions()
		if err != nil {
			return fmt.Errorf("%w", err)
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
				fmt.Printf("Error getting transaction type: %s\n", err)
			}
		})
		// j/k navigation inside dropdown
		idDropDown.SetInputCapture(vimNavigation)
	}

	var categoryDropdown *tview.DropDown
	var category string

	allowedTransactionTypes, err := listOfAllowedTransactionTypes()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

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

	form := styleForm(tview.NewForm().
		AddFormItem(idDropDown).
		AddFormItem(typeDropdown).
		AddFormItem(amountField).
		AddFormItem(categoryDropdown).
		AddFormItem(descriptionField).
		AddButton("Update", func() {
			amount := amountField.GetText()

			description := descriptionField.GetText()

			if err := handleUpdateTransaction(transactionType, transactionId, amount, category, description); err != nil {
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

	// back to mainMenu on ESC or q key press
	form.SetInputCapture(exitShortcuts)

	tui.SetRoot(form, true).SetFocus(form)

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
