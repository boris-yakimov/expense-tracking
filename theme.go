package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TuiTheme struct {
	LabelColor            tcell.Color
	FieldTextColor        tcell.Color
	FieldBackgroundColor  tcell.Color
	ButtonTextColor       tcell.Color
	ButtonBackgroundColor tcell.Color
	BorderColor           tcell.Color
	TitleColor            tcell.Color
}

// TODO: configure a tokyo equivalent scheme as default
var theme = TuiTheme{
	LabelColor:            tcell.ColorYellow,
	FieldTextColor:        tcell.ColorWhite,
	FieldBackgroundColor:  tcell.ColorBlack,
	ButtonTextColor:       tcell.ColorBlack,
	ButtonBackgroundColor: tcell.ColorGreen,
	BorderColor:           tcell.ColorDarkCyan,
	TitleColor:            tcell.ColorAqua,
}

// helper to style input fields in TUI
func styleInputField(field *tview.InputField) *tview.InputField {
	return field.SetLabelColor(theme.LabelColor).
		SetFieldTextColor(theme.FieldTextColor).
		SetFieldBackgroundColor(theme.FieldBackgroundColor)
}

// helper to style drop downs in TUI
func styleDropdown(dropdown *tview.DropDown) *tview.DropDown {
	return dropdown.SetLabelColor(theme.LabelColor).
		SetFieldTextColor(theme.FieldTextColor).
		SetFieldBackgroundColor(theme.FieldBackgroundColor)
}

// helper to style forms in TUI
func styleForm(form *tview.Form) *tview.Form {
	form.SetButtonTextColor(theme.ButtonTextColor).
		SetButtonBackgroundColor(theme.ButtonBackgroundColor).
		SetLabelColor(theme.LabelColor)

	form.SetBorderColor(theme.BorderColor)
	form.SetTitleColor(theme.TitleColor)

	return form
}
