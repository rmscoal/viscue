package list

import (
	"viscue/tui/style"

	"github.com/charmbracelet/lipgloss"
)

var (
	DefaultItemStyle = lipgloss.NewStyle().
				Foreground(style.ColorWhite).
				PaddingLeft(1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				Bold(true)
	DefaultSelectedItemStyle = DefaultItemStyle.Background(style.ColorPurple)

	_defaultViewportHeight = 10
	_defaultViewportWidth  = 20
)

type Styles struct {
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Item:         DefaultItemStyle,
		SelectedItem: DefaultSelectedItemStyle,
	}
}
