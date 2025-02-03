package prompt

import (
	"strings"

	"viscue/tui/entity"
	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

var textInputWidth = 24

var categoryMode int8 = 0
var passwordMode int8 = 1

type prompt struct {
	db *sqlx.DB

	fields       []textinput.Model
	help         help.Model
	title        string
	model        any
	mode         int8
	showPassword bool
	needConfirm  bool
}

func NewPassword(db *sqlx.DB, password entity.Password) tea.Model {
	title := "Add Password"
	if password.Id != 0 {
		title = "Edit Password"
	}

	fields := make([]textinput.Model, 4)
	fields[0] = textinput.New()
	fields[0].Cursor.SetMode(cursor.CursorBlink)
	fields[0].Width = 24
	fields[0].Prompt = "Name: "
	fields[0].SetValue(password.Name)
	fields[0].Focus()

	fields[1] = textinput.New()
	fields[1].Cursor.SetMode(cursor.CursorBlink)
	fields[1].Prompt = "Email: "
	fields[1].SetValue(password.Email)

	fields[2] = textinput.New()
	fields[2].Cursor.SetMode(cursor.CursorBlink)
	fields[2].Prompt = "Username: "
	fields[2].SetValue(password.Username)

	fields[3] = textinput.New()
	fields[3].Cursor.SetMode(cursor.CursorBlink)
	fields[3].Prompt = "Password: "
	fields[3].EchoMode = textinput.EchoPassword
	fields[3].EchoCharacter = '•'
	fields[3].SetValue(password.Password)

	for i := range fields {
		fields[i].Width = textInputWidth
	}

	h := help.New()
	h.ShowAll = true
	return &prompt{
		db:     db,
		fields: fields,
		help:   h,
		title:  title,
		model:  password,
		mode:   passwordMode,
	}
}

func NewCategory(db *sqlx.DB, category entity.Category) tea.Model {
	title := "Add Category"
	if category.Id != 0 {
		title = "Edit Category"
	}

	fields := make([]textinput.Model, 1)
	fields[0] = textinput.New()
	fields[0].Width = textInputWidth
	fields[0].Cursor.SetMode(cursor.CursorBlink)
	fields[0].Focus()
	fields[0].SetValue(category.Name)
	fields[0].Prompt = "Name: "

	h := help.New()
	h.ShowAll = true
	return &prompt{
		db:     db,
		fields: fields,
		help:   h,
		title:  title,
		model:  category,
		mode:   categoryMode,
	}
}

func (m *prompt) Init() tea.Cmd {
	return textinput.Blink
}

func (m *prompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, baseKeys.Tab):
			m.cycleFocus(msg)
			return m, nil
		case key.Matches(msg, baseKeys.Escape):
			log.Debug("(*prompt).Update: closing prompt")
			return m, m.close
		case key.Matches(msg, baseKeys.Enter):
			return m, m.submit
		case key.Matches(msg, passwordKeys.ShowPassword):
			m.showHidePassword()
			return m, nil
		}
		return m.updateFields(msg)
	case cursor.BlinkMsg:
		return m.updateFields(msg)
	}

	return m, nil
}

func (m *prompt) View() string {
	textInputs := lipgloss.JoinVertical(lipgloss.Center, lo.Map(m.fields,
		func(item textinput.Model, index int) string {
			return item.View()
		})...)

	helpView := strings.Repeat("\n", 2)
	if m.mode == categoryMode {
		helpView += m.help.View(baseKeys)
	} else {
		helpView += m.help.View(passwordKeys)
	}

	// TODO: Error view...

	return style.Dialog(
		lipgloss.JoinVertical(
			lipgloss.Center,
			m.title,
			textInputs,
			helpView,
		),
	)
}

//////////////////////////////////
//////////// PRIVATE /////////////
//////////////////////////////////

func (m *prompt) cycleFocus(msg tea.KeyMsg) {
	length := len(m.fields) - 1
	_, index, _ := lo.FindIndexOf(m.fields,
		func(item textinput.Model) bool {
			return item.Focused()
		})
	m.fields[index].Blur()
	if msg.String() == "tab" {
		index++
	} else {
		index--
	}
	if index > length {
		index = 0
	} else if index < 0 {
		index = length
	}
	m.fields[index].Focus()
}

func (m *prompt) showHidePassword() {
	if m.mode != passwordMode {
		return
	}

	m.showPassword = !m.showPassword
	if m.showPassword {
		m.fields[3].EchoMode = textinput.EchoNormal
	} else {
		m.fields[3].EchoMode = textinput.EchoPassword
	}
}

func (m *prompt) updateFields(msg tea.Msg) (tea.Model, tea.Cmd) {
	var commands []tea.Cmd
	var cmd tea.Cmd
	for i := range m.fields {
		m.fields[i], cmd = m.fields[i].Update(msg)
		commands = append(commands, cmd)
	}
	return m, tea.Batch(commands...)
}
