package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	descriptionMaxCharLength = 40
)

// TODO: add option for month year - default shows current, but if you start typing a previous month or year it is available based on the data you have
func formAddTransaction() error {
	var transactionType string
	var category string
	var description string
	var categoryDropdown *tview.DropDown

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
	}
	categoryDropdown.SetCurrentOption(0)

	descriptionField := styleInputField(tview.NewInputField().SetLabel("Description"))
	// TODO: can I not just set a limit on the field when users type it ?
	if len(descriptionField.GetText()) > descriptionMaxCharLength {
		return fmt.Errorf("\ndescription should be a maximum of %v characters, provided %v", descriptionMaxCharLength, len(description))
	}

	// TODO: by default this and than provide the option to add for a specific month or year by selecting it from dropdown
	year := strings.ToLower(strconv.Itoa(time.Now().Year()))
	month := strings.ToLower(time.Now().Month().String())

	// TODO: display footer that shows ESC or 'q' can be pressed to go back to menu
	form := styleForm(tview.NewForm().
		AddFormItem(typeDropdown).
		AddFormItem(amountField).
		AddFormItem(categoryDropdown).
		AddFormItem(descriptionField).
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

func handleAddTransaction(transactionType, amount, category, description, month, year string) error {
	txType, err := normalizeTransactionType(transactionType)
	if err != nil {
		return fmt.Errorf("transaction type error: %w", err)
	}

	txAmount, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("\ninvalid amount: %w\n", err)
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

	fmt.Printf("\n successfully added %s €%.2f | %s | %s\n", txType, amount, category, description)
	return nil
}
