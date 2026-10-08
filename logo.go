package main

import (
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// "ANSI Shadow" style block letters used for the login screen title
var logoGlyphs = map[rune][]string{
	'E': {
		"███████╗",
		"██╔════╝",
		"█████╗  ",
		"██╔══╝  ",
		"███████╗",
		"╚══════╝",
	},
	'X': {
		"██╗  ██╗",
		"╚██╗██╔╝",
		" ╚███╔╝ ",
		" ██╔██╗ ",
		"██╔╝ ██╗",
		"╚═╝  ╚═╝",
	},
	'P': {
		"██████╗ ",
		"██╔══██╗",
		"██████╔╝",
		"██╔═══╝ ",
		"██║     ",
		"╚═╝     ",
	},
	'N': {
		"███╗   ██╗",
		"████╗  ██║",
		"██╔██╗ ██║",
		"██║╚██╗██║",
		"██║ ╚████║",
		"╚═╝  ╚═══╝",
	},
	'S': {
		"███████╗",
		"██╔════╝",
		"███████╗",
		"╚════██║",
		"███████║",
		"╚══════╝",
	},
}

// synthwave palette (tview hex color tags)
var (
	// top to bottom gradient of the title letters: hot pink -> purple -> neon blue
	logoTitleGradient = []string{"[#ff71ce]", "[#ff5fd2]", "[#e05cf0]", "[#b967ff]", "[#8b7bff]", "[#01cdfe]"}
	// top to bottom gradient of the sun: yellow -> orange -> pink -> magenta
	logoSunGradient = []string{"[#fff100]", "[#ffd319]", "[#ffb01f]", "[#ff901f]", "[#ff6b4a]", "[#ff2975]", "[#f222ff]"}
)

const (
	logoSubtitleColor = "[#01cdfe::b]" // neon cyan
	logoAccentColor   = "[#b967ff]"    // purple
	logoHorizonColor  = "[#ff2975::b]" // neon pink
	logoGridColor     = "[#f222ff]"    // magenta
	logoGridFarColor  = "[#7a1fa2]"    // dimmer magenta for the far horizontal lines
)

// width of the sun / horizon / grid scene under the title
const logoSceneWidth = 58

// renders a word with logoGlyphs, one string per glyph row (unknown runes are skipped)
func renderLogoWord(word string) []string {
	rows := make([]string, 6)
	for _, r := range word {
		glyph, ok := logoGlyphs[r]
		if !ok {
			continue
		}
		for i := range rows {
			rows[i] += glyph[i]
		}
	}
	return rows
}

// upper half of a sun sitting on the horizon, one gradient color per row
// the lower rows use half blocks so the dark gaps between them read as the classic synthwave stripes
func renderLogoSun(radius int) []string {
	height := len(logoSunGradient)
	lines := make([]string, height)
	for y := 0; y < height; y++ {
		// vertical distance of the middle of this row from the horizon, doubled because terminal cells are ~2x taller than wide
		dy := (float64(height-y) - 0.5) * 2
		half := int(math.Round(math.Sqrt(math.Max(0, float64(radius*radius)-dy*dy))))
		ch := "█"
		if y >= height-3 {
			ch = "▀"
		}
		pad := (logoSceneWidth - 2*half) / 2
		lines[y] = strings.Repeat(" ", pad) + logoSunGradient[y] + strings.Repeat(ch, 2*half) + Reset +
			strings.Repeat(" ", logoSceneWidth-pad-2*half)
	}
	return lines
}

// neon floor below the horizon: lines converging on the center, crossed by horizontal lines that get
// further apart towards the viewer; the rows nearest the horizon are dimmer
func renderLogoGrid() []string {
	type gridRow struct {
		spacing    float64 // distance between the converging lines on this row
		horizontal bool    // row is also a horizontal grid line
		color      string
	}
	rows := []gridRow{
		{3, true, logoGridFarColor},
		{5, false, logoGridFarColor},
		{7, true, logoGridColor},
		{9.5, false, logoGridColor},
		{12, true, logoGridColor},
	}

	center := float64(logoSceneWidth-1) / 2
	lines := make([]string, len(rows))
	for i, row := range rows {
		fill := " "
		if row.horizontal {
			fill = "─"
		}
		cells := []rune(strings.Repeat(fill, logoSceneWidth))
		for k := -20; k <= 20; k++ {
			x := int(math.Round(center + float64(k)*row.spacing))
			if x < 0 || x >= logoSceneWidth {
				continue
			}
			switch {
			case k < 0:
				cells[x] = '╱'
			case k > 0:
				cells[x] = '╲'
			default:
				cells[x] = '│'
			}
		}
		lines[i] = row.color + string(cells) + Reset
	}
	return lines
}

// the full synthwave logo: gradient title, subtitle, sun on the horizon and the neon grid
func renderSynthwaveLogo() string {
	var lines []string
	for i, row := range renderLogoWord("EXPENSE") {
		lines = append(lines, logoTitleGradient[i]+row+Reset)
	}
	lines = append(lines,
		logoAccentColor+"━━━━━━━━━━  "+Reset+logoSubtitleColor+"T  R  A  C  K  I  N  G"+Reset+logoAccentColor+"  ━━━━━━━━━━"+Reset,
		"",
	)
	lines = append(lines, renderLogoSun(17)...)
	lines = append(lines, logoHorizonColor+strings.Repeat("━", logoSceneWidth)+Reset)
	lines = append(lines, renderLogoGrid()...)
	return strings.Join(lines, "\n")
}

// number of rows the logo takes up
func logoHeight() int {
	return strings.Count(renderSynthwaveLogo(), "\n") + 1
}

// text view with the logo, centered and on the app background
func newLogoView() *tview.TextView {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(tview.AlignCenter).
		SetText(renderSynthwaveLogo())
	view.SetBackgroundColor(theme.BackgroundColor)
	return view
}

// login screen layout: logo above the form, both centered vertically
// the logo is dropped (instead of pushing the form off screen) when the terminal is too short or narrow for both
type loginLayout struct {
	*tview.Flex
	logo       *tview.TextView
	logoRows   int
	formHeight int
	formWidth  int
}

func newLoginLayout(form tview.Primitive, formWidth, formHeight int) *loginLayout {
	logo := newLogoView()
	l := &loginLayout{logo: logo, logoRows: logoHeight(), formHeight: formHeight, formWidth: formWidth}

	// horizontal centering of the form
	centeredForm := styleFlex(tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(form, formWidth, 1, true).
		AddItem(nil, 0, 1, false))

	l.Flex = styleFlex(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).                  // top spacer
		AddItem(logo, l.logoRows, 0, false).        // logo
		AddItem(nil, 1, 0, false).                  // gap between logo and form
		AddItem(centeredForm, formHeight, 1, true). // form
		AddItem(nil, 0, 1, false))                  // bottom spacer
	return l
}

// shows or hides the logo depending on the available space, then draws the layout
func (l *loginLayout) Draw(screen tcell.Screen) {
	_, _, width, height := l.GetInnerRect()
	if height >= l.logoRows+1+l.formHeight && width >= logoSceneWidth {
		l.ResizeItem(l.logo, l.logoRows, 0)
	} else {
		l.ResizeItem(l.logo, 0, 0)
	}
	l.Flex.Draw(screen)
}
