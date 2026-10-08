package main

import (
	"strings"
	"testing"

	"github.com/rivo/tview"
)

// every line of a logo block must have the same visible width, otherwise center alignment makes it look jagged
func TestSynthwaveLogoLineWidths(t *testing.T) {
	width := func(line string) int { return tview.TaggedStringWidth(line) }

	title := renderLogoWord("EXPENSE")
	if len(title) != 6 {
		t.Fatalf("expected 6 title rows, got %d", len(title))
	}
	for i, row := range title {
		if width(row) != width(title[0]) {
			t.Errorf("title row %d has width %d, expected %d", i, width(row), width(title[0]))
		}
	}

	scene := append(renderLogoSun(17), renderLogoGrid()...)
	for i, line := range scene {
		if w := width(line); w != logoSceneWidth {
			t.Errorf("scene line %d has width %d, expected %d", i, w, logoSceneWidth)
		}
	}

	if got := strings.Count(renderSynthwaveLogo(), "\n") + 1; got != logoHeight() {
		t.Errorf("logoHeight() = %d, logo has %d lines", logoHeight(), got)
	}
}

// the logo is shown when there is room for it and dropped on small terminals so the login form stays visible
func TestLoginLayoutHidesLogoWhenTooSmall(t *testing.T) {
	cases := []struct {
		name     string
		w, h     int
		wantLogo bool
	}{
		{"large terminal", 120, 45, true},
		{"too short", 120, 20, false},
		{"too narrow", 50, 45, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			screen := startSimulatedTui(t)
			screen.SetSize(c.w, c.h)

			var layout *loginLayout
			tui.QueueUpdateDraw(func() {
				layout = newLoginLayout(tview.NewBox(), 50, 9)
				pages.AddPage("login", layout, true, true)
			})
			// the layout decides on the logo while drawing, so wait until it has been drawn at the simulated size
			waitFor(t, "layout drawn", func() bool {
				_, _, w, _ := layout.GetRect()
				return w == c.w
			})

			var logoHeight int
			onTui(func() { _, _, _, logoHeight = layout.logo.GetRect() })
			if shown := logoHeight > 0; shown != c.wantLogo {
				t.Errorf("logo shown = %v (height %d), expected %v", shown, logoHeight, c.wantLogo)
			}
		})
	}
}
