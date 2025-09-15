package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Creates a TUI form with prompt for logi. On initial login provides a form set a password. On subsequent attempts, prompts for password to login with. The same password is also used to generate an encryption key that is then used for encrypting/decrypting the database.
func loginForm() error {
	// first-run for encryption: if no encrypted DB exists, prompt to set a password
	if _, err := os.Stat(encFile); os.IsNotExist(err) {
		setNewPasswordForm()
		return nil // i.e. don't proceed to build the login form in the event of a first login
	}

	passwordInputField := styleInputField(tview.NewInputField().
		SetLabel("Enter Password: ").
		SetMaskCharacter('*'))

	var formWithMessage *tview.Flex

	message := styleTextView(tview.NewTextView().
		SetText("").
		SetTextAlign(tview.AlignCenter))

	form := styleForm(tview.NewForm().
		AddFormItem(passwordInputField).
		AddButton("Login", func() {
			entered := passwordInputField.GetText()

			// store password in memory to derive an encryption key from it
			setUserPassword(entered)

			// if encrypted file exists, decrypt with provided password
			if _, err := os.Stat(encFile); err == nil {

				if err := decryptDatabase(globalConfig.SQLitePath); err != nil {

					if errors.Is(err, ErrWrongPassword) {
						// wrong password, stay on login prompt
						message.SetText("Wrong password. Try again.")
						passwordInputField.SetText("")
						clearUserPassword() // remove pass from memory on error
						return
					}

					// some other unexpected error occured - corrupted file, permision issues, etc
					showErrorModal(fmt.Sprintf("Decryption failed: %s", err), formWithMessage, passwordInputField)
					clearUserPassword()
					return
				}
			}

			// initialize DB connection now that the DB is decrypted or already plaintext
			if err := initDb(globalConfig.SQLitePath); err != nil {
				showErrorModal(fmt.Sprintf("failed to initialize DB: %s\n", err), formWithMessage, passwordInputField)
				clearUserPassword() // remove pass from memory on error
				return
			}

			// optional migration from JSON to SQLite (runs only if env var is set)
			if os.Getenv("MIGRATE_TRANSACTION_DATA") == "true" {
				if globalConfig.StorageType != StorageSQLite {
					showErrorModal("migration requires sqlite storage", formWithMessage, passwordInputField)
					return
				}
				if err := migrateJsonToDb(); err != nil {
					showErrorModal(fmt.Sprintf("migration failed: %v", err), formWithMessage, passwordInputField)
					return
				}
			}

			// TODO: login to go directly to list transactions
			// if err := mainMenu(); err != nil {
			// 	showErrorModal(fmt.Sprintf("failed to initialize main menu: %s\n", err), formWithMessage, passwordInputField)
			// 	clearUserPassword() // remove pass from memory on error
			// 	return
			// }
			if err := gridVisualizeTransactions("", ""); err != nil {
				showErrorModal(fmt.Sprintf("list transactions error:\n\n%s", err), formWithMessage, passwordInputField)
				clearUserPassword() // remove pass from memory on error
				return
			}

		}).
		AddButton("Quit", func() {
			// allow main() to run post-Run() cleanup (encrypt + remove plaintext)
			tui.Stop()
		}))

	form.SetButtonsAlign(tview.AlignCenter)

	// just a spacer that can be used to structure the UI, using this instead of nil because it also inherits theme styling
	topSpacer := tview.NewBox()

	// form + message - vertical alignment
	formWithMessage = styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(topSpacer, 1, 0, false).
		// AddItem(infoMsg, 1, 1, false).
		AddItem(form, 0, 1, true).
		AddItem(message, 1, 0, false))

	formWithMessage.SetBorder(true).
		SetTitle("Expense Tracking Tool").
		SetTitleAlign(tview.AlignCenter)

	// horizontal centering
	initialModal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).             // left spacer
		AddItem(formWithMessage, 50, 1, true). // modal width fixed at 40
		AddItem(nil, 0, 1, false))             // right spacer

	// vertical centering
	centeredModal := styleFlex(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).         // top spacer
		AddItem(initialModal, 9, 1, true). // form box automatic height
		AddItem(nil, 0, 1, false))         // bottom spacer

	root := styleFlex(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(centeredModal, 0, 1, true))

	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			// allow main() to run post-Run() cleanup (encrypt + remove plaintext)
			tui.Stop()
		}
		return event
	})

	tui.SetRoot(root, true).SetFocus(passwordInputField)
	return nil
}

// creates the TUI form for setting a new password
func setNewPasswordForm() {
	passwordInputField := styleInputField(tview.NewInputField().
		SetLabel("Enter Password: ").
		SetMaskCharacter('*'))
	repeatPasswordField := styleInputField(tview.NewInputField().
		SetLabel("Repeat Password: ").
		SetMaskCharacter('*'))

	var formWithMessage *tview.Flex

	message := styleTextView(tview.NewTextView().
		SetText("").
		SetTextAlign(tview.AlignCenter))

	form := styleForm(tview.NewForm().
		AddFormItem(passwordInputField).
		AddFormItem(repeatPasswordField).
		AddButton("Confirm", func() {
			entered := passwordInputField.GetText()
			repeat := repeatPasswordField.GetText()

			if entered == repeat {
				if err := addInitialPassword(entered); err != nil {
					showErrorModal(fmt.Sprintf("failed to set a new password: %v", err), formWithMessage, passwordInputField)
					return // interrupt here
				}

				// proceed directly to app using the newly set in-memory password
				if err := initDb(globalConfig.SQLitePath); err != nil {
					showErrorModal(fmt.Sprintf("failed to initialize DB: %s\n", err), formWithMessage, passwordInputField)
					clearUserPassword() // remove pass from memory on error
					return
				}

				// optional migration from JSON to SQLite (runs only if env var is set)
				if os.Getenv("MIGRATE_TRANSACTION_DATA") == "true" {
					if globalConfig.StorageType != StorageSQLite {
						showErrorModal("migration requires sqlite storage", formWithMessage, passwordInputField)
						return
					}
					if err := migrateJsonToDb(); err != nil {
						showErrorModal(fmt.Sprintf("migration failed: %v", err), formWithMessage, passwordInputField)
						return
					}
				}

				// TODO: login to go directly to list transactions
				// if err := mainMenu(); err != nil {
				// 	showErrorModal(fmt.Sprintf("failed to initialize main menu: %s\n", err), formWithMessage, passwordInputField)
				// 	clearUserPassword() // remove pass from memory on error
				// 	return
				// }
				if err := gridVisualizeTransactions("", ""); err != nil {
					showErrorModal(fmt.Sprintf("list transactions error:\n\n%s", err), formWithMessage, passwordInputField)
					clearUserPassword() // remove pass from memory on error
					return
				}

			} else {
				message.SetText("Passwords do not match. Try Again.")
				passwordInputField.SetText("")
				repeatPasswordField.SetText("")
			}
		}).
		AddButton("Quit", func() {
			// allow main() to run post-Run() cleanup (encrypt + remove plaintext)
			tui.Stop()
		}))
	form.SetButtonsAlign(tview.AlignCenter)

	infoMsg := styleTextView(tview.NewTextView().
		SetText("Set a password").
		SetTextAlign(tview.AlignCenter))

	// just a spacer that can be used to structure the UI, using this instead of nil because it also inherits theme styling
	formSpacer := tview.NewBox()
	disclaimerMsg := styleTextView(tview.NewTextView().
		SetText("A new transaction DB will be created and your password will be used to encrypt its contents.").
		SetWrap(true))

	// disclaimer + form + message - vertical alignment
	formWithMessage = styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(formSpacer, 1, 0, false).    // 1 row on top
		AddItem(disclaimerMsg, 2, 0, false). // 2 rows after disclaimer
		AddItem(formSpacer, 2, 0, false).    // 1 row before info message
		AddItem(infoMsg, 1, 1, false).       // 1 row after info message
		AddItem(form, 7, 0, true).           // the form spans 7 rows - including password and repeat password fields and buttons for ok and cancel bellow
		AddItem(message, 0, 1, false))       // dynamic size of field that contains error message (such as repeat password doesn't match)

	formWithMessage.SetBorder(true).
		SetTitle("Expense Tracking Tool").
		SetTitleAlign(tview.AlignCenter)

	// horizontal centering
	initialModal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).             // left spacer
		AddItem(formWithMessage, 50, 1, true). // modal fixed width
		AddItem(nil, 0, 1, false))             // right spacer

	// vertical centering
	centeredModal := styleFlex(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).          // top spacer
		AddItem(initialModal, 15, 1, true). // form box automatic height
		AddItem(nil, 0, 1, false))          // bottom spacer

	root := styleFlex(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(centeredModal, 0, 1, true))

	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			// allow main() to run post-Run() cleanup (encrypt + remove plaintext)
			tui.Stop()
		}
		return event
	})

	tui.SetRoot(root, true).SetFocus(passwordInputField)
}

// helper to check if newly set password is adequate and stores it in memory for later use in generating an encryption key
func addInitialPassword(providedPass string) error {
	if providedPass == "" {
		return fmt.Errorf("password cannot be empty")
	}
	setUserPassword(providedPass)
	return nil
}
