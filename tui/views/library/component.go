package library

import (
	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Title styles
var (
	titleStyle = lipgloss.NewStyle().
			Background(style.ColorPurple).
			Foreground(style.ColorNormal).
			MarginBottom(1).
			Padding(0, 1)
	unfocusedTitleStyle = titleStyle.Background(style.ColorGray)
)

// Search styles
var (
	searchBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorGray).
		Align(lipgloss.Left, lipgloss.Center).
		PaddingLeft(1).
		PaddingRight(1)
)

// Search component
func newSearch(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = "🔍: "
	ti.Placeholder = placeholder
	ti.Cursor.SetMode(cursor.CursorStatic)
	return ti
}
