package library

import (
	"viscue/tui/entity"
	"viscue/tui/style"
	"viscue/tui/views/library/message"
	"viscue/tui/views/library/submodel/prompt"
	"viscue/tui/views/library/submodel/shelf"
	"viscue/tui/views/library/submodel/sidebar"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jmoiron/sqlx"
)

type Model struct {
	db *sqlx.DB

	// Submodels
	prompt  tea.Model
	sidebar tea.Model
	shelf   tea.Model

	// focusedSubmodel indicates what submodel is currently focused
	// `0` indicates sidebar
	// `1` indicates shelf
	// `2` indicates prompt
	focusedSubmodel int8
}

func New(db *sqlx.DB) tea.Model {
	m := Model{
		db:              db,
		shelf:           shelf.New(db),
		sidebar:         sidebar.New(db),
		focusedSubmodel: 1,
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return message.ShelfFocused
		},
		m.shelf.Init(),
		m.sidebar.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case message.SwitchFocusMsg:
		m.focusedSubmodel = int8(msg)
	case message.OpenPromptMsg[entity.Password]:
		m.prompt = prompt.New(m.db, msg.Payload)
	case message.OpenPromptMsg[entity.Category]:
		m.prompt = prompt.New(m.db, msg.Payload)
	case message.ClosePromptMsg[entity.Category]:
		m.prompt = nil
		return m, func() tea.Msg {
			return message.SidebarFocused
		}
	case message.ClosePromptMsg[entity.Password]:
		m.prompt = nil
		return m, func() tea.Msg {
			return message.ShelfFocused
		}
	}

	cmds := make([]tea.Cmd, 3)
	if m.prompt != nil {
		m.prompt, cmds[0] = m.prompt.Update(msg)
	}
	m.sidebar, cmds[1] = m.sidebar.Update(msg)
	m.shelf, cmds[2] = m.shelf.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var helpView string
	switch m.focusedSubmodel {
	case 0:
		helpView = style.HelpContainer(help.New().View(sidebar.Keys))
	case 1:
		helpView = style.HelpContainer(help.New().View(shelf.Keys))
	case 2:
		helpView = style.HelpContainer(help.New().View(prompt.Keys))
	}
	if m.prompt != nil {
		return lipgloss.JoinVertical(
			lipgloss.Top,
			m.prompt.View(),
			helpView,
		)
	}

	shelfView := m.shelf.View()
	sidebarView := m.sidebar.View()
	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			sidebarView,
			shelfView,
		),
		helpView,
	)
}
