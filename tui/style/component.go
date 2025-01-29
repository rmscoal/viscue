package style

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// Spinner

func NewSpinner() spinner.Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Jump),
		spinner.WithStyle(lipgloss.NewStyle().
			Foreground(ColorTerminalGreen),
		),
	)
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorTerminalGreen)

	return s
}

// Button

var ButtonStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFF7DB")).
	Background(lipgloss.Color("#888B7E")).
	Padding(0, 3).
	MarginTop(1)
var Button = ButtonStyle.Render

var ActiveButtonStyle = ButtonStyle.
	Foreground(lipgloss.Color("#FFF7DB")).
	Background(lipgloss.Color("#F25D94")).
	MarginRight(2).
	Underline(true)
var ActiveButton = ActiveButtonStyle.Render

// Dialog

var DialogStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Padding(1, 0)
var Dialog = DialogStyle.Render
