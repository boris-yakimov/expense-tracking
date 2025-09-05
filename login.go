package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/crypto/bcrypt"
)

func loginForm() error {
	passHashInDb, err := getHashedPassword()
	if err != nil {
		return fmt.Errorf("failed to get hashed password from db: %w", err)
	}

	// if no previous pass has been set, switch to the setPasswordForm()
	if passHashInDb == "" {
		setPasswordForm()
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

			if isValid := validatePassword(entered, passHashInDb); isValid {
				// store password in memory to derive an encryption key from it
				setUserPassword(entered)

				// TODO: why are we storing the decrypted db as a file, shouldn't it be only in memory ?
				// decrypt the database before proceeding
				if err := decryptDatabase(globalConfig.SQLitePath); err != nil {
					showErrorModal(fmt.Sprintf("failed to decrypt database: %s\n", err), formWithMessage, passwordInputField)
					clearUserPassword() // remove pass from memory on error
					return
				}

				if err := mainMenu(); err != nil {
					showErrorModal(fmt.Sprintf("failed to initialize main menu: %s\n", err), formWithMessage, passwordInputField)
					clearUserPassword() // remove pass from memory on error
				}
			} else {
				message.SetText("Wrong password. Try again.")
				passwordInputField.SetText("")
			}
		}).
		AddButton("Quit", func() {
			tui.Stop()
			os.Exit(0)
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
			tui.Stop()
			os.Exit(0)
		}
		return event
	})

	tui.SetRoot(root, true).SetFocus(passwordInputField)
	return nil
}

func setPasswordForm() {
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

				message.SetText("New password has been set")
				loginForm() // switch back to login screen
			} else {
				message.SetText("Passwords do not match. Try Again.")
				passwordInputField.SetText("")
				repeatPasswordField.SetText("")
			}
		}).
		AddButton("Quit", func() {
			tui.Stop()
			os.Exit(0)
		}))
	form.SetButtonsAlign(tview.AlignCenter)

	infoMsg := styleTextView(tview.NewTextView().
		SetText("Set a password").
		SetTextAlign(tview.AlignCenter))

	// just a spacer that can be used to structure the UI, using this instead of nil because it also inherits theme styling
	topSpacer := tview.NewBox()

	// form + message - vertical alignment
	formWithMessage = styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(topSpacer, 1, 0, false).
		AddItem(infoMsg, 1, 1, false).
		AddItem(form, 0, 1, true).
		AddItem(message, 1, 0, false))

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
			tui.Stop()
			os.Exit(0)
		}
		return event
	})

	tui.SetRoot(root, true).SetFocus(passwordInputField)
}

func addInitialPassword(providedPass string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(providedPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("unable to generate password, err: %w", err)
	}

	sqlTx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("db connection during new password hashing failed, err: %w", err)
	}
	sqlStatement, err := sqlTx.Prepare(`
		INSERT INTO authentication
		(password_hash, created_at)
		VALUES(?, ?)
	`)
	if err != nil {
		sqlTx.Rollback()
		return fmt.Errorf("prepare insert of new hashed password failed, err: %w", err)
	}
	defer sqlStatement.Close()

	creationTimestamp := time.Now().Format("200601021504") // year, month, day, hour, minute

	_, err = sqlStatement.Exec(
		hashedPass,
		creationTimestamp,
	)
	if err != nil {
		sqlTx.Rollback()
		return fmt.Errorf("insert of new hashed password failed, err: %w", err)
	}

	// TODO: audit log
	// TODO: maybe this should be a modal in the TUI instead
	if err = sqlTx.Commit(); err != nil {
		return fmt.Errorf("unable to commit db transaction when adding new password hash, err: %w", err)
	}

	return nil
}

func validatePassword(providedPass string, storedHash string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedPass)); err != nil {
		return false
	}
	return true
}

func getHashedPassword() (hashedPassword string, err error) {
	err = db.QueryRow(`
			SELECT password_hash
			FROM authentication
		`).Scan(&hashedPassword)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// no password has been set yet
			return "", nil
		}
		return "", fmt.Errorf("unable to retrieve hashed password: %w", err)
	}

	return hashedPassword, nil
}
