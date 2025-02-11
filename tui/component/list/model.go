// The list model is the minimal usage of a list component
// that suits the needs of Viscue. It renders a set of items
// vertically and is able to scroll up and down direction
// viewing each item correctly.

package list

import (
	"strings"

	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	DefaultItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Bold(true)
	DefaultSelectedItemStyle = DefaultItemStyle.Background(style.ColorPurple)
	DefaultBlurredItemStyle  = lipgloss.NewStyle().
					PaddingLeft(1).
					Foreground(style.ColorGray).
					Bold(true)
	DefaultBlurredSelectedItemStyle = DefaultBlurredItemStyle.Background(style.ColorGray)
)

type Styles struct {
	Item                lipgloss.Style
	SelectedItem        lipgloss.Style
	BlurredItem         lipgloss.Style
	BlurredSelectedItem lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Item:                DefaultItemStyle,
		SelectedItem:        DefaultSelectedItemStyle,
		BlurredItem:         DefaultBlurredItemStyle,
		BlurredSelectedItem: DefaultBlurredSelectedItemStyle,
	}
}

type Model struct {
	Styles Styles

	vp      viewport.Model
	items   []Item
	currIdx int
	focused bool
}

func New(opts ...Option) Model {
	m := Model{
		Styles: DefaultStyles(),
		vp:     viewport.New(0, 0),
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// Options ...

type Option func(*Model)

// WithHeight sets the height of the list viewport
func WithHeight(height int) Option {
	return func(m *Model) {
		m.SetHeight(height)
	}
}

// WithWidth sets the width of the list viewport
func WithWidth(width int) Option {
	return func(m *Model) {
		m.SetWidth(width)
	}
}

// WithItems set the items to the list
func WithItems(items []Item) Option {
	return func(m *Model) {
		m.SetItems(items)
	}
}

func WithFocused(focused bool) Option {
	return func(m *Model) {
		if focused {
			m.Focus()
		} else {
			m.Blur()
		}
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Ignore msg when blurred
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.Down()
			return m, nil
		case "k", "up":
			m.Up()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	m.vp.SetContent(m.renderItems())
	return m.vp.View()
}

type Item interface {
	String() string
}

func (m *Model) SetItems(items []Item) {
	m.items = items
	m.vp.SetContent(m.renderItems())
	m.currIdx = 0
	m.vp.SetYOffset(0)
}

func (m Model) Items() []Item {
	return m.items
}

func (m Model) renderItems() string {
	var content strings.Builder
	for idx, item := range m.items {
		str := item.String()
		width := lipgloss.Width(str)
		if width > m.vp.Width-2 {
			str = str[:m.vp.Width-2] + "…"
		}
		fn := m.Styles.BlurredItem.Render
		if m.focused {
			fn = m.Styles.Item.Render
		}
		if idx == m.currIdx {
			if m.focused {
				fn = m.Styles.SelectedItem.Render
			} else {
				fn = m.Styles.BlurredSelectedItem.Render
			}
		}
		content.WriteString(fn(str))
		content.WriteRune('\n')
	}
	return content.String()
}

func (m Model) Index() int {
	return m.currIdx
}

func (m Model) SelectedItem() Item {
	if len(m.items) == 0 {
		return nil
	}
	return m.items[m.currIdx]
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) Blur() {
	m.focused = false
}

func (m Model) Focused() bool {
	return m.focused
}

func (m *Model) SetHeight(height int) {
	m.vp.Height = height
}

func (m Model) Height() int {
	return m.vp.Height
}

func (m *Model) SetWidth(width int) {
	m.Styles.Item.UnsetWidth()
	m.Styles.SelectedItem.UnsetWidth()
	m.Styles.BlurredItem.UnsetWidth()
	m.Styles.BlurredSelectedItem.UnsetWidth()

	m.Styles.Item = m.Styles.Item.Width(width)
	m.Styles.SelectedItem = m.Styles.SelectedItem.Width(width)
	m.Styles.BlurredItem = m.Styles.BlurredItem.Width(width)
	m.Styles.BlurredSelectedItem = m.Styles.BlurredSelectedItem.Width(width)

	m.vp.Width = width
}

func (m Model) Width() int {
	return m.vp.Width
}

func (m *Model) Up() {
	length := len(m.items)
	if length == 0 || m.currIdx <= 0 {
		return
	}

	m.currIdx--
	// The height of the current row is currIdx + 1,
	// assuming all rows have height of 1
	currItemYPos := m.currIdx + 1
	for currItemYPos <= m.vp.YOffset && !m.vp.AtTop() {
		m.vp.LineUp(1)
	}
}

func (m *Model) Down() {
	length := len(m.items)
	if length == 0 || m.currIdx >= length-1 {
		return
	}

	m.currIdx++
	// The height of the current row is currIdx + 1,
	// assuming all rows have height of 1
	currItemYPos := m.currIdx + 1
	for currItemYPos > m.vp.Height+m.vp.YOffset && !m.vp.AtBottom() {
		m.vp.LineDown(1)
	}
}
