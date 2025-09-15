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

// helper to style grid type objects in the TUI
func styleGrid(grid *tview.Grid) *tview.Grid {
	grid.SetBackgroundColor(theme.BackgroundColor)
	grid.SetBorderColor(theme.BorderColor)
	grid.SetTitleColor(theme.TitleColor)
	grid.SetBordersColor(theme.BorderColor)

	return grid
}

// helper to style tables in the TUI
func styleTable(table *tview.Table) *tview.Table {
	table.SetBackgroundColor(theme.BackgroundColor)
	table.SetBorderColor(theme.BorderColor)
	table.SetBordersColor(theme.BorderColor)
	table.SetTitleColor(theme.TitleColor)

	return table
}

// helper to style lists in the TUI
func styleList(list *tview.List) *tview.List {
	list.SetBackgroundColor(theme.BackgroundColor)
	list.SetBorderColor(theme.BorderColor)
	list.SetTitleColor(theme.TitleColor)
	list.SetMainTextColor(theme.FieldTextColor)
	list.SetSecondaryTextColor(theme.LabelColor)

	return list
}

// helper to style flex type objects in the TUI
func styleFlex(flex *tview.Flex) *tview.Flex {
	flex.SetBackgroundColor(theme.BackgroundColor)
	flex.SetBorderColor(theme.BorderColor)
	flex.SetTitleColor(theme.TitleColor)

	return flex
}

// helper to style textView objects in the TUI
func styleTextView(textView *tview.TextView) *tview.TextView {
	textView.SetBackgroundColor(theme.FieldBackgroundColor)
	textView.SetBorderColor(theme.BorderColor)
	textView.SetTitleColor(theme.TitleColor)
	textView.SetTextColor(theme.FieldTextColor)

	return textView
}

// helper to style modal objects in the TUI
func styleModal(modal *tview.Modal) *tview.Modal {
	modal.SetBorderColor(theme.BorderColor)
	modal.SetTitleColor(theme.TitleColor)
	modal.SetBackgroundColor(theme.BackgroundColor)
	modal.SetTextColor(theme.FieldTextColor)

	return modal
}
