// The list model is the minimal usage of a list component
// that suits the needs of Viscue. It renders a set of items
// vertically and is able to scroll up and down direction
// viewing each item correctly.

package list

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

type Model struct {
	Styles Styles

	vp      viewport.Model
	items   []Item
	content string
	currIdx int
}

var _ tea.Model = (*Model)(nil)

func New() *Model {
	return &Model{
		Styles: DefaultStyles(),
		vp: viewport.New(
			_defaultViewportWidth,
			_defaultViewportHeight,
		),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *Model) View() string {
	// Render our items
	var content strings.Builder
	for idx, item := range m.items {
		if idx == m.currIdx {
			content.WriteString(m.Styles.SelectedItem.Render(item.String()))
		} else {
			content.WriteString(m.Styles.Item.Render(item.String()))
		}
		content.WriteRune('\n')
	}
	m.vp.SetContent(content.String())

	return m.vp.View()
}

type Item interface {
	String() string
}

func (m *Model) SetItems(items []Item) {
	m.items = items
	m.currIdx = 0
	m.vp.SetYOffset(0)
}

func (m *Model) Items() []Item {
	return m.items
}

func (m *Model) SetIndex(idx int) {
	m.currIdx = idx
}

func (m *Model) Index() int {
	return m.currIdx
}

func (m *Model) SelectedItem() Item {
	return m.items[m.currIdx]
}

// Up scrolls the list model upward
func (m *Model) Up() {
	length := len(m.items)
	if length == 0 || m.currIdx <= 0 {
		return
	}

	m.currIdx--
	currItem := m.items[m.currIdx]
	currItemHeight := lipgloss.Height(m.Styles.SelectedItem.Render(currItem.String()))
	currItemYPos := lo.Sum(lo.FilterMap(m.items,
		func(item Item, index int) (int, bool) {
			fn := m.Styles.Item
			if index == m.currIdx {
				fn = m.Styles.SelectedItem
			}
			return lipgloss.Height(fn.Render(item.String())), index <= m.currIdx
		}))
	for currItemYPos <= m.vp.YOffset {
		m.vp.SetYOffset(m.vp.YOffset - currItemHeight)
	}
}

// Down scrolls the list model downward
func (m *Model) Down() {
	length := len(m.items)
	if length == 0 || m.currIdx >= length-1 {
		return
	}

	m.currIdx++
	currItem := m.items[m.currIdx]
	currItemHeight := lipgloss.Height(m.Styles.SelectedItem.Render(currItem.String()))
	currItemYPos := lo.Sum(lo.FilterMap(m.items,
		func(item Item, index int) (int, bool) {
			fn := m.Styles.Item
			if index == m.currIdx {
				fn = m.Styles.SelectedItem
			}
			return lipgloss.Height(fn.Render(item.String())), index <= m.currIdx
		}))
	for currItemYPos > m.vp.Height+m.vp.YOffset {
		m.vp.SetYOffset(m.vp.YOffset + currItemHeight)
	}
}

func (m *Model) itemHeight(idx int) int {
	if m.items == nil {
		return lipgloss.Height(m.Styles.Item.Render(" "))
	}

	return lipgloss.Height(m.Styles.Item.Render(m.items[0].String()))
}
