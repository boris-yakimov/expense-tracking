package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/crypto/bcrypt"
)

// TODO: fix temp pass with a better approach
var passwordHash string = fmt.Sprintf("%x", sha256.Sum256([]byte("secret123")))

func tuiLogin() {
	// TODO: add a check if a password has been set previously
	// TODO: if it has prompt for login
	// TODO: if it has not, prompt to set password
	passwordInputField := styleInputField(tview.NewInputField().
		SetLabel("Enter Password: ").
		SetMaskCharacter('*'))

	message := styleTextView(tview.NewTextView().
		SetText("").
		SetTextAlign(tview.AlignCenter))

	form := styleForm(tview.NewForm().
		AddFormItem(passwordInputField).
		AddButton("Login", func() {
			entered := passwordInputField.GetText()
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(entered)))

			if hash == passwordHash {
				if err := mainMenu(); err != nil {
					fmt.Fprintf(os.Stderr, "failed to initialize main menu: %v\n", err)
					os.Exit(1)
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

	// form + message - vertical alignment
	formWithMessage := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(message, 1, 0, false))

	formWithMessage.SetBorder(true).
		SetTitle("Expense Tracking Tool").
		SetTitleAlign(tview.AlignCenter)

	// horizontal centering
	initialModal := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).             // left spacer
		AddItem(formWithMessage, 40, 1, true). // modal width fixed at 40
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
	fmt.Printf("successfully added new password")
	return nil
}

func validatePassword(providedPass string, storedHash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedPass)); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	return nil
}

func getHashedPassword() (hashedPassword string, err error) {
	sqlTx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("db connection during password validation failed, err: %w", err)
	}

	// TODO: to be reviewed
	err = sqlTx.QueryRow(`
			SELECT password_hash
			FROM authentication
			LIMIT 1
		`).Scan(&hashedPassword)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve hashed password from db, err: %w", err)
	}

	return hashedPassword, nil
}
