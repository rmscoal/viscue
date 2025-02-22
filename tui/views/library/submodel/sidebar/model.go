package sidebar

import (
	"viscue/tui/component/list"
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
)

type Model struct {
	db *sqlx.DB

	search     textinput.Model
	list       list.Model
	categories []entity.Category
}

func New(db *sqlx.DB) tea.Model {
	search := textinput.New()
	search.Prompt = "🔍: "
	search.Placeholder = "Search category..."
	search.Cursor.SetMode(cursor.CursorStatic)

	m := Model{
		db:     db,
		search: search,
		list:   list.New(list.WithFocused(false)),
	}

	m.calculateDimension()
	return m
}

func (m Model) Init() tea.Cmd {
	return m.LoadItems
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DataLoadedMsg:
		log.Debug("DataLoadedMsg received for Sidebar")
		m.categories = msg.Data
		m.list.SetItems(
			lo.Map(m.categories,
				func(category entity.Category, _ int) list.Item {
					return category
				},
			),
		)
		return m, nil
	case message.SwitchFocusMsg:
		if msg == message.SidebarFocused {
			m.list.Focus()
			return m, nil
		} else {
			m.list.Blur()
			return m, nil
		}
	case cursor.BlinkMsg:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if !m.list.Focused() {
			// Since our parent model passes msg to both
			// shelf and sidebar, especially for tea.KeyMsg
			// we ignore msg if our model is not focused
			return m, nil
		} else if m.search.Focused() {
			switch keypress := msg.String(); keypress {
			case "enter":
				m.search.Blur()
				return m, m.CategorySelectedMsg
			case "esc":
				m.search.SetValue("")
				m.search.Blur()
				return m, m.CategorySelectedMsg
			}
			defer m.filter()
			var searchCmd tea.Cmd
			m.search, searchCmd = m.search.Update(msg)
			return m, tea.Batch(searchCmd, m.CategorySelectedMsg)
		} else {
			switch keypress := msg.String(); keypress {
			case "k", "up", "j", "down":
				var listCmd tea.Cmd
				m.list, listCmd = m.list.Update(msg)
				return m, tea.Sequence(listCmd, m.CategorySelectedMsg)
			case "ctrl+l":
				m.search.Blur()
				m.list.Blur()
				return m, func() tea.Msg { return message.ShelfFocused }
			case "a":
				return m, m.AddCategoryPromptMsg()
			case "e", "enter":
				return m, m.EditCategoryPromptMsg()
			case "d":
				return m, m.DeleteCategoryPromptMsg()
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
	if m.list.Focused() {
		titleStyle = style.ModelTitleFocusedStyle
	}

	// Search Box
	searchBoxStyle := style.SearchBoxStyle.
		Width(m.list.Width()).
		BorderForeground(style.ColorGray)
	if m.search.Focused() {
		searchBoxStyle = searchBoxStyle.BorderForeground(style.ColorPurple)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Category"),
		searchBoxStyle.Render(m.search.View()),
		m.list.View(),
	)
}
