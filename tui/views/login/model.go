package login

import (
	"database/sql"
	"errors"

	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
)

type keyMap struct {
	Tab    key.Binding
	Quit   key.Binding
	Submit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Quit, k.Submit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.Quit, k.Submit},
	}
}

var keys = keyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "tab"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("q", "quit"),
	),
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
}

type login struct {
	db *sqlx.DB

	help          help.Model
	usernameInput textinput.Model
	passwordInput textinput.Model

	shouldCreateAccount bool
	err                 error
}

func New(db *sqlx.DB) tea.Model {
	// Retrieve existing username from database.
	var username string
	query := `SELECT value FROM configurations WHERE key = ?`
	err := db.QueryRowx(query, "username").Scan(&username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Fatal(
			"failed querying username from sqlite when setting up app",
			"err", err,
		)
	}

	usernameInput := textinput.New()
	usernameInput.Prompt = "Username"
	usernameInput.Cursor.SetMode(cursor.CursorBlink)
	usernameInput.PromptStyle = style.TextInputPromptStyle.Width(10)
	usernameInput.Focus()

	passwordInput := textinput.New()
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.Prompt = "Password"
	passwordInput.PromptStyle = style.TextInputPromptStyle.Width(10)
	passwordInput.Cursor.SetMode(cursor.CursorBlink)

	if username != "" {
		usernameInput.SetValue(username)
		usernameInput.Blur()
		passwordInput.Focus()
	}

	return &login{
		db:                  db,
		help:                help.New(),
		usernameInput:       usernameInput,
		passwordInput:       passwordInput,
		shouldCreateAccount: username == "",
	}
}

func (m *login) Init() tea.Cmd {
	return textinput.Blink
}

func (m *login) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Tab):
			if m.usernameInput.Focused() {
				m.usernameInput.Blur()
				m.passwordInput.Focus()
			} else {
				m.usernameInput.Focus()
				m.passwordInput.Blur()
			}
		case key.Matches(msg, keys.Submit):
			return m, m.submit
		default:
			var commands [2]tea.Cmd
			m.usernameInput, commands[0] = m.usernameInput.Update(msg)
			m.passwordInput, commands[1] = m.passwordInput.Update(msg)
			if len(m.usernameInput.Value()) >= 34 ||
				len(m.passwordInput.Value()) >= 34 {
				m.usernameInput.Width = 36
				m.passwordInput.Width = 36
			}
			return m, tea.Batch(commands[:]...)
		}
	case cursor.BlinkMsg:
		var commands [2]tea.Cmd
		m.usernameInput, commands[0] = m.usernameInput.Update(msg)
		m.passwordInput, commands[1] = m.passwordInput.Update(msg)
		return m, tea.Batch(commands[:]...)
	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m *login) View() string {
	height := style.CalculateAppHeight()
	loginContainer := lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Height(height).
		Render

	form := lipgloss.JoinVertical(lipgloss.Left,
		m.usernameInput.View(),
		m.passwordInput.View(),
	)

	if m.err != nil {
		form = lipgloss.JoinVertical(
			lipgloss.Center,
			form,
			style.ErrorText(m.err.Error()),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		loginContainer(form),
		style.HelpContainer(m.help.View(keys)),
	)
}
