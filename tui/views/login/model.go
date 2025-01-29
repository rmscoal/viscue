package login

import (
	"database/sql"
	"errors"
	fatalLog "log"
	"strings"

	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
		fatalLog.Fatalf(
			"failed querying username from sqlite when setting up app: %s",
			err.Error(),
		)
	}

	usernameInput := textinput.New()
	usernameInput.Prompt = "Username: "
	usernameInput.Cursor.SetMode(cursor.CursorBlink)
	usernameInput.Focus()

	passwordInput := textinput.New()
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.Prompt = "Password: "
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
	msg := "Please enter your credentials!"
	helpView := m.help.View(keys)
	view := msg +
		"\n" +
		m.usernameInput.View() +
		"\n" +
		m.passwordInput.View()

	if m.err != nil {
		view += strings.Repeat("\n", 2) + style.ErrorText(m.err.Error())
	}

	view += strings.Repeat("\n", 3) + helpView
	return view
}
