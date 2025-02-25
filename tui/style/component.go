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
			Foreground(ColorPurple),
		),
	)

	return s
}

// Button

var ButtonStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFF7DB")).
	Background(lipgloss.Color("#888B7E")).
	Padding(0, 2).
	MarginTop(1)

var ActiveButtonStyle = ButtonStyle.
	Foreground(lipgloss.Color("#FFF7DB")).
	Background(lipgloss.Color("#F25D94")).
	MarginRight(2).
	Underline(true)

// Text Input / Search Box

var TextInputPromptStyle = lipgloss.NewStyle().
	PaddingRight(1).
	AlignHorizontal(lipgloss.Left).
	BorderStyle(lipgloss.Border{Right: ":"}).
	BorderRight(true).
	MarginRight(1)

var SearchBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorPurple).
	Align(lipgloss.Left, lipgloss.Center).
	PaddingLeft(1).
	PaddingRight(1)

// Pane Border

var PaneBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Align(lipgloss.Top, lipgloss.Top).
	Padding(1)
