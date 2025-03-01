package warning

import (
	"fmt"

	"viscue/tui/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	text   string
	width  int
	height int
}

func NewScreenSize(width, height int) tea.Model {
	return Model{
		width:  width,
		height: height,
		text:   "Screen size is too small. Dimension: %dx%d",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	return lipgloss.Place(
		m.width, m.height, lipgloss.Center, lipgloss.Center,
		style.ErrorText(fmt.Sprintf(m.text, m.width, m.height)),
	)
}
