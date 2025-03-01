package prompt

import (
	"errors"
	"fmt"

	"viscue/tui/component/list"
	"viscue/tui/entity"
	"viscue/tui/style"
	"viscue/tui/tool/cache"
	"viscue/tui/views/library/message"

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
	list            list.Model
	button          lipgloss.Style
	payload         any // holds either Password or Category entity.
	err             error
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

	for _, opt := range opts {
		opt(&m)
	}

	switch payload := payload.(type) {
	case entity.Category:
		if payload.Id != 0 {
			if m.isDeletion {
				m.title = "Delete Category"
				m.focusSubmitButton()
				break
			} else {
				m.title = "Edit Category"
			}
		} else {
			m.title = "Create Category"
		}
		m.fields = make([]textinput.Model, 1)
		m.fields[0] = textinput.New()
		m.fields[0].Prompt = "Name"
		m.fields[0].PromptStyle = style.TextInputPromptStyle
		m.fields[0].Cursor.SetMode(cursor.CursorBlink)
		m.fields[0].Focus()
		m.fields[0].SetValue(payload.Name)
		m.fields[0].Width = m.textInputWidth()
	case entity.Password:
		if payload.Id != 0 {
			if m.isDeletion {
				m.title = "Delete Password"
				m.focusSubmitButton()
				break
			} else {
				m.title = "Edit Password"
			}
		} else {
			m.title = "Create Password"
		}

		if err := m.getCategories(); err != nil {
			m.err = errors.New("failed building categories dropdown")
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
			return item.Id == payload.CategoryId.Int64
		})
		m.setCategoryField(category)
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
			m.fields[i].Width = m.textInputWidth()
		}
		m.list = list.New(list.WithFocused(false))
		m.list.SetHeight(4)
		m.list.SetWidth(m.fields[0].Width)
		m.list.SetItems(
			lo.Map(m.categories,
				func(item entity.Category, index int) list.Item {
					return item
				},
			),
		)
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SubmitError:
		m.err = msg
		return m, nil
	case tea.KeyMsg:
		switch {
		case m.isDeletion:
			switch {
			case key.Matches(msg, Keys.Close):
				return m, m.Close
			case key.Matches(msg, Keys.Submit):
				return m, m.Delete
			}
		default:
			switch {
			case m.list.Focused():
				switch {
				case key.Matches(msg, DropdownKeys.Up),
					key.Matches(msg, DropdownKeys.Down):
					var cmd tea.Cmd
					m.list, cmd = m.list.Update(msg)
					return m, cmd
				case key.Matches(msg, DropdownKeys.Select):
					category := m.list.SelectedItem().(entity.Category)
					m.setCategoryField(category)
					fallthrough
				case key.Matches(msg, DropdownKeys.Cancel):
					m.list.Blur()
					return m, func() tea.Msg {
						return message.SetHelpKeysMsg{
							Keys: Keys,
						}
					}
				}
			default:
				switch {
				case key.Matches(msg, Keys.Cycle):
					m.cycleFocus(msg)
					return m, nil
				case key.Matches(msg, Keys.Close):
					return m, m.Close
				case key.Matches(msg, Keys.Submit):
					if m.isButtonFocused() {
						return m, m.Submit
					}
					if m.isPasswordPrompt() && m.pointer == 1 {
						m.list.Focus()
						return m, func() tea.Msg {
							return message.SetHelpKeysMsg{
								Keys: DropdownKeys,
							}
						}
					}
				case key.Matches(msg, Keys.TogglePasswordVisibility):
					m.togglePasswordVisibility()
					return m, nil
				default:
					m.err = nil // Clear existing error on type
				}
			}
		}
		return m.updateTextInputs(msg)
	case cursor.BlinkMsg:
		return m.updateTextInputs(msg)
	case message.OpenPromptMsg[entity.Password]:
		return m, func() tea.Msg {
			return message.SetHelpKeysMsg{
				Keys: Keys,
			}
		}
	case message.OpenPromptMsg[entity.Category]:
		return m, func() tea.Msg {
			return message.SetHelpKeysMsg{
				Keys: Keys,
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var view string
	if m.isDeletion {
		var subtext string
		switch payload := m.payload.(type) {
		case entity.Category:
			subtext = fmt.Sprintf("Delete category: %s", payload.Name)
		case entity.Password:
			subtext = fmt.Sprintf("Delete password: %s", payload.Name)
		}
		view = textboxRenderer(
			lipgloss.JoinVertical(
				lipgloss.Center,
				subtext,
				m.button.Render(),
			),
		)
	} else {
		var textFields []string
		if m.isPasswordPrompt() && m.list.Focused() {
			textFields = lo.FilterMap(m.fields,
				func(item textinput.Model, index int) (string, bool) {
					return item.View(), index < 1
				})
			label := style.TextInputPromptStyle.Width(10).
				Render("Category")
			selectBox := lipgloss.JoinHorizontal(
				lipgloss.Left, label, m.list.View(),
			)
			textFields = append(textFields, selectBox)
		} else {
			textFields = lo.Map(m.fields,
				func(item textinput.Model, index int) string {
					return item.View()
				})
		}
		view = textboxRenderer(
			lipgloss.JoinVertical(
				lipgloss.Center,
				titleRenderer(m.title),
				lipgloss.JoinVertical(
					lipgloss.Left,
					textFields...,
				),
				m.button.Render(),
			),
		)
	}

	if m.err != nil {
		view = lipgloss.JoinVertical(
			lipgloss.Center,
			view,
			style.ErrorText(m.err.Error()),
		)
	}

	return lipgloss.Place(
		m.availableWidth,
		m.availableHeight,
		lipgloss.Center,
		lipgloss.Center,
		view,
	)
}
