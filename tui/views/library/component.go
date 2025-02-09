package library

import (
	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Title styles
var (
	titleStyle = lipgloss.NewStyle().
			Background(style.ColorPurple).
			Foreground(style.ColorBlack).
			MarginBottom(1).
			Padding(0, 1)
	unfocusedTitleStyle = titleStyle.Background(style.ColorGray)
)

// Table component
func newTableColumns(width int) []table.Column {
	width = (width - 8) / 3
	return []table.Column{
		{Title: "Id", Width: 0},
		{Title: "CategoryId", Width: 0},
		{Title: "Name", Width: width},
		{Title: "Email", Width: width},
		{Title: "Username", Width: width},
		{Title: "Password", Width: 0},
	}
}

func newTableStyle(focus bool) table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorWhite).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Bold(false).
		Foreground(style.ColorWhite).
		Background(style.ColorGray)
	if focus {
		s.Selected = s.Selected.
			Foreground(style.ColorBlack).
			Background(style.ColorPurple)
	}

	return s
}

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
