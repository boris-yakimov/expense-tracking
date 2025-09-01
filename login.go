package main

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var passwordHash string = fmt.Sprintf("%x", sha256.Sum256([]byte("secret123")))

func tuiLogin() {
	passwordInputField := styleInputField(tview.NewInputField().
		SetLabel("Enter Password: ").
		SetMaskCharacter('*'))

	// TODO: check if theme is applied to this
	message := tview.NewTextView().SetText("")

	flex := styleFlex(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(passwordInputField, 3, 1, true).
		AddItem(message, 1, 1, false))

	passwordInputField.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
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

		case tcell.KeyEsc:
			tui.Stop()
			os.Exit(0)
		}
	})

	tui.SetRoot(flex, true).SetFocus(passwordInputField)
}
