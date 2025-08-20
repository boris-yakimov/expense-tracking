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
	BackgroundColor       tcell.Color
}

var theme = TuiTheme{
	LabelColor:            tcell.ColorLightCyan,             // cyan labels
	FieldTextColor:        tcell.ColorWhite,                 // light text
	FieldBackgroundColor:  tcell.NewRGBColor(40, 44, 52),    // deep gray for inputs
	ButtonTextColor:       tcell.ColorBlack,                 // black text on buttons
	ButtonBackgroundColor: tcell.NewRGBColor(152, 195, 121), // soft green buttons
	BorderColor:           tcell.ColorTeal,                  // teal borders
	TitleColor:            tcell.ColorAqua,                  // aqua titles
	BackgroundColor:       tcell.NewRGBColor(26, 27, 38),    // dark navy base
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
		SetLabelColor(theme.LabelColor).
		SetFieldBackgroundColor(theme.FieldBackgroundColor).
		SetFieldTextColor(theme.FieldTextColor)

	form.SetBorderColor(theme.BorderColor)
	form.SetTitleColor(theme.TitleColor)

	return form
}

func styleGrid(grid *tview.Grid) *tview.Grid {
	grid.SetBackgroundColor(theme.BackgroundColor)
	grid.SetBorderColor(theme.BorderColor)
	grid.SetTitleColor(theme.TitleColor)

	return grid
}
