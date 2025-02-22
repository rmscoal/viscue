package prompt

import (
	"viscue/tui/entity"
	"viscue/tui/style"
	"viscue/tui/tool/cache"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

var (
	minimumTextInputWidth = 36

	textboxRenderer = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(style.ColorPurple).
			Padding(1).
			Render
	titleRenderer = lipgloss.NewStyle().Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			Padding(0, 2).
			BorderForeground(style.ColorPurplePale).
			Foreground(style.ColorPurplePale).
			MarginBottom(2).
			Render
)

// Model that displays modal for either editing or inserting
// a new password or category, or a confirmation modal used
// when deleting an entity.
type Model struct {
	db *sqlx.DB

	categories      []entity.Category
	fields          []textinput.Model
	button          lipgloss.Style
	payload         any // holds either Password or Category entity.
	title           string
	availableWidth  int
	availableHeight int
	pointer         int
	showPassword    bool
	isDeletion      bool
}

type Option func(*Model)

func IsDeletion(boolean bool) Option {
	return func(m *Model) {
		m.isDeletion = boolean
	}
}

func New(db *sqlx.DB, payload any, opts ...Option) Model {
	termWidth, appHeight := cache.Get[int](cache.TerminalWidth),
		style.CalculateAppHeight()-2

	m := Model{
		db:              db,
		payload:         payload,
		button:          style.ButtonStyle.SetString("Submit"),
		availableWidth:  termWidth,
		availableHeight: appHeight,
	}

	switch payload := payload.(type) {
	case entity.Category:
		m.fields = make([]textinput.Model, 1)
		m.fields[0] = textinput.New()
		m.fields[0].Prompt = "Name"
		m.fields[0].PromptStyle = style.TextInputPromptStyle
		m.fields[0].Cursor.SetMode(cursor.CursorBlink)
		m.fields[0].Focus()
		m.fields[0].SetValue(payload.Name)
		m.fields[0].Width = min(m.availableWidth-2, minimumTextInputWidth)
	case entity.Password:
		if err := m.getCategories(); err != nil {
			// 	TODO: Handler error
		}

		m.fields = make([]textinput.Model, 5)
		m.fields[0] = textinput.New()
		m.fields[1] = textinput.New()
		m.fields[2] = textinput.New()
		m.fields[3] = textinput.New()
		m.fields[4] = textinput.New()

		m.fields[0].Prompt = "Name"
		m.fields[0].SetValue(payload.Name)
		m.fields[1].Prompt = "Category"
		category, _ := lo.Find(m.categories, func(item entity.Category) bool {
			return item.Id == payload.Id
		})
		m.fields[1].SetValue(category.Name)
		m.fields[2].Prompt = "Email"
		m.fields[2].SetValue(payload.Email)
		m.fields[3].Prompt = "Username"
		m.fields[3].SetValue(payload.Username)
		m.fields[4].Prompt = "Password"
		m.fields[4].SetValue(payload.Password)
		m.fields[4].EchoMode = textinput.EchoPassword
		m.fields[4].EchoCharacter = '•'

		for i := range m.fields {
			if i == 0 {
				m.fields[i].Focus()
			}
			m.fields[i].PromptStyle = style.TextInputPromptStyle.Width(10)
			m.fields[i].Cursor.SetMode(cursor.CursorBlink)
			m.fields[i].Width = min(m.availableWidth-2, minimumTextInputWidth)
		}
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

func (m Model) Init() tea.Cmd {
	// TODO: I think we have to load categories here...
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Cycle):
			m.cycleFocus(msg)
			return m, nil
		case key.Matches(msg, Keys.Close):
			return m, m.Close
		case key.Matches(msg, Keys.TogglePasswordVisibility):
			m.togglePasswordVisibility()
			return m, nil
		}
		return m.updateTextInputs(msg)
	case cursor.BlinkMsg:
		return m.updateTextInputs(msg)
	}
	return m, nil
}

func (m Model) View() string {
	textbox := textboxRenderer(
		lipgloss.JoinVertical(
			lipgloss.Center,
			titleRenderer(m.title),
			lipgloss.JoinVertical(
				lipgloss.Left,
				lo.Map(m.fields, func(item textinput.Model, index int) string {
					return item.View()
				})...,
			),
			m.button.Render(),
		),
	)

	return lipgloss.Place(
		m.availableWidth,
		m.availableHeight,
		lipgloss.Center,
		lipgloss.Center,
		textbox,
	)
}
