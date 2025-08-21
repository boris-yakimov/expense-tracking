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
	descriptionMaxLength = 40 // chars
)

// TODO: add option for month year - default shows current, but if you start typing a previous month or year it is available based on the data you have
func formAddTransaction() error {
	var transactionType string
	var category string
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
			fmt.Println(err)
		}
		categoryDropdown.SetOptions(opts, func(selectedOption string, index int) {
			category = selectedOption
		})
	}

	categoryDropdown.SetCurrentOption(0)

	descriptionField := styleInputField(tview.NewInputField().SetLabel("Description"))

	// TODO: display footer that shows ESC or 'q' can be pressed to go back to menu
	form := styleForm(tview.NewForm().
		AddFormItem(typeDropdown).
		AddFormItem(amountField).
		AddFormItem(categoryDropdown).
		AddFormItem(descriptionField).
		AddButton("Add", func() {
			amount := amountField.GetText()
			description := descriptionField.GetText()

			// TODO: refactor add transactions to no longer expect cli args so we can just pass these cleanly
			if _, err := addTransaction([]string{transactionType, amount, category, description}); err != nil {
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

	// back to mainMenu on ESC or q key press
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

// add <transaction_type> <amount> <category> <description>
func addTransaction(args []string) (success bool, err error) {
	if len(args) < 4 {
		return false, fmt.Errorf("usage: add <transcation type> <amount> <category> <description>")
	}

	transactionType, err := normalizeTransactionType(args[0])
	if err != nil {
		return false, fmt.Errorf("transaction type error: %w", err)
	}

	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return false, fmt.Errorf("\ninvalid amount: %w\n", err)
	}

	category := args[2]
	if _, ok := allowedTransactionCategories[transactionType][category]; !ok {
		fmt.Printf("\ninvalid transaction category: \"%s\"", category)
		showAllowedCategories(transactionType) // expense, income, investment
		return false, fmt.Errorf("\n\nPlease pick a valid transaction category from the list above.")
	}

	description := strings.Join(args[3:], " ")
	if len(description) > descriptionMaxLength {
		return false, fmt.Errorf("\ndescription should be a maximum of %v characters, provided %v", descriptionMaxLength, len(description))
	}

	if !validDescriptionInputFormat(description) {
		return false, fmt.Errorf("\ninvalid character in description, should contain only letters, numbers, spaces, commas, or dashes")
	}

	// TODO: extend this to support adding transactions for a specific month and not only the current one
	year := strings.ToLower(strconv.Itoa(time.Now().Year()))
	month := strings.ToLower(time.Now().Month().String())

	return handleTransactionAdd(transactionType, amount, category, description, month, year)
}

func handleTransactionAdd(transactionType string, amount float64, category, description, month, year string) (success bool, err error) {
	transactions, loadFileErr := loadTransactions()
	if loadFileErr != nil {
		return false, fmt.Errorf("unable to load transactions file: %w", loadFileErr)
	}

	if _, ok := transactions[year]; !ok {
		transactions[year] = make(map[string]map[string][]Transaction)
	}

	if _, ok := transactions[year][month]; !ok {
		transactions[year][month] = make(map[string][]Transaction)
	}

	if _, ok := transactions[year][month][transactionType]; !ok {
		transactions[year][month][transactionType] = []Transaction{}
	}

	var transactionId string
	if transactionId, err = generateTransactionId(); err != nil {
		return false, fmt.Errorf("unable to generate transaction id: %w", err)
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
			return false, fmt.Errorf("unable to generate transaction id: %w", err)
		}
	}

	if len(transactionId) > 8 {
		return false, fmt.Errorf("transcation id should have a maximum of 8 chars, current id %s with length of %v", transactionId, len(transactionId))
	}

	newTransaction := Transaction{
		Id:          transactionId,
		Amount:      amount,
		Category:    category,
		Description: description,
	}

	transactions[year][month][transactionType] = append(transactions[year][month][transactionType], newTransaction)
	if saveTransactionErr := saveTransactions(transactions); saveTransactionErr != nil {
		return false, fmt.Errorf("Error saving transaction: %w", saveTransactionErr)
	}

	fmt.Printf("\n successfully added %s €%.2f | %s | %s\n", transactionType, amount, category, description)
	return true, nil
}
