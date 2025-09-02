package main

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TODO: fix temp pass with a better approach
var passwordHash string = fmt.Sprintf("%x", sha256.Sum256([]byte("secret123")))

func tuiLogin() {
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
