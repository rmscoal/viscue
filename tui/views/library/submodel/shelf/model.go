package shelf

import (
	"viscue/tui/component/table"
	"viscue/tui/entity"
	"viscue/tui/style"
	"viscue/tui/views/library/message"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"golang.design/x/clipboard"
)

type Model struct {
	db *sqlx.DB

	search             textinput.Model
	table              table.Model
	passwords          []entity.Password
	selectedCategoryId int64
}

func New(db *sqlx.DB) tea.Model {
	search := textinput.New()
	search.Prompt = "🔍: "
	search.Placeholder = "Search password..."
	search.Cursor.SetMode(cursor.CursorStatic)

	m := Model{
		db:     db,
		search: search,
		table: table.New(
			table.WithColumns(
				[]table.Column{
					{Title: "Id", Width: 0},
					{Title: "CategoryId", Width: 0},
					{Title: "Name", Width: 24},
					{Title: "Email", Width: 24},
					{Title: "Username", Width: 24},
					{Title: "Password", Width: 0},
				}),
			table.WithFocused(true),
		),
	}

	m.calculateDimension()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.LoadItems,
		func() tea.Msg {
			return clipboard.Init()
		},
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DataLoadedMsg:
		log.Debug("DataLoadedMsg received in shelf Model")
		m.passwords = msg.Data
		m.table.SetRows(
			lo.Map(m.passwords,
				func(password entity.Password, index int) table.Row {
					return password.ToTableRow()
				},
			),
		)
		return m, nil
	case message.CategorySelectedMsg:
		m.selectedCategoryId = int64(msg)
		m.sync()
		return m, nil
	case message.SwitchFocusMsg:
		if msg == message.ShelfFocused {
			m.table.Focus()
			return m, nil
		} else {
			m.table.Blur()
			return m, nil
		}
	case cursor.BlinkMsg:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if !m.table.Focused() {
			// Since our parent model passes msg to both
			// shelf and sidebar, especially for tea.KeyMsg
			// we ignore msg if our model is not focused
			return m, nil
		} else if m.search.Focused() {
			switch msg.String() {
			case "esc":
				m.search.SetValue("")
				m.search.Blur()
				return m, nil
			case "enter":
				m.search.Blur()
				return m, nil
			}
			defer m.filter()
			var searchCmd tea.Cmd
			m.search, searchCmd = m.search.Update(msg)
			return m, searchCmd
		} else {
			switch msg.String() {
			case "j", "down", "k", "up":
				var cmd tea.Cmd
				m.table, cmd = m.table.Update(msg)
				return m, cmd
			case "ctrl+h":
				m.table.Blur()
				m.search.Blur()
				return m, func() tea.Msg { return message.SidebarFocused }
			case "y":
				return m, m.CopyToClipboard
			case "a":
				return m, m.AddPasswordPromptMsg()
			case "e", "enter":
				return m, m.EditPasswordPromptMsg()
			case "d":
				return m, m.DeletePasswordPromptMsg()
			case "f":
				m.search.Focus()
				return m, textinput.Blink
			case "c":
				m.search.Blur()
				m.search.SetValue("")
				return m, nil
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	// Title
	titleStyle := style.ModelTitleStyle
	if m.table.Focused() {
		titleStyle = style.ModelTitleFocusedStyle
	}

	// Search Box
	searchBoxStyle := style.SearchBoxStyle.
		Width(m.table.Width()).
		BorderForeground(style.ColorGray)
	if m.search.Focused() {
		searchBoxStyle = searchBoxStyle.BorderForeground(style.ColorPurple)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Password"),
		searchBoxStyle.Render(m.search.View()),
		m.table.View(),
	)
}
