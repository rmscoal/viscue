package list

import (
	"viscue/tui/style"

	"github.com/charmbracelet/lipgloss"
)

var (
	DefaultItemStyle = lipgloss.NewStyle().
				Foreground(style.ColorWhite).
				PaddingLeft(1).
				Bold(true)
	DefaultSelectedItemStyle        = DefaultItemStyle.Background(style.ColorPurple)
	DefaultBlurredItemStyle         = DefaultItemStyle.Foreground(style.ColorGray)
	DefaultBlurredSelectedItemStyle = DefaultItemStyle.Background(style.ColorGray)

	_defaultViewportHeight = 10
	_defaultViewportWidth  = 20
)

type Styles struct {
	Item                lipgloss.Style
	SelectedItem        lipgloss.Style
	BlurredItem         lipgloss.Style
	BlurredSelectedItem lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Item:                DefaultItemStyle,
		SelectedItem:        DefaultSelectedItemStyle,
		BlurredItem:         DefaultBlurredItemStyle,
		BlurredSelectedItem: DefaultBlurredSelectedItemStyle,
	}
}
