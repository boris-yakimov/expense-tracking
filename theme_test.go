package main

import (
	"github.com/rivo/tview"
	"testing"
)

func TestStyleInputField(t *testing.T) {
	field := tview.NewInputField()
	styledField := styleInputField(field)

	if styledField == nil {
		t.Errorf("Expected styled field, got nil")
	}
}

func TestStyleDropdown(t *testing.T) {
	dropdown := tview.NewDropDown()
	styledDropdown := styleDropdown(dropdown)

	if styledDropdown == nil {
		t.Errorf("Expected styled dropdown, got nil")
	}
}

func TestStyleForm(t *testing.T) {
	form := tview.NewForm()
	styledForm := styleForm(form)

	if styledForm == nil {
		t.Errorf("Expected styled form, got nil")
	}
}

func TestStyleGrid(t *testing.T) {
	grid := tview.NewGrid()
	styledGrid := styleGrid(grid)

	if styledGrid == nil {
		t.Errorf("Expected styled grid, got nil")
	}
}

func TestStyleTable(t *testing.T) {
	table := tview.NewTable()
	styledTable := styleTable(table)

	if styledTable == nil {
		t.Errorf("Expected styled table, got nil")
	}
}

func TestStyleList(t *testing.T) {
	list := tview.NewList()
	styledList := styleList(list)

	if styledList == nil {
		t.Errorf("Expected styled list, got nil")
	}
}
