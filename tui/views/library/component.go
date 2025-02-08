package library

import (
	"fmt"
	"io"

	"viscue/tui/entity"
	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	listPagination = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	listItem       = lipgloss.NewStyle().PaddingLeft(2).
			Foreground(style.ColorWhite).
			Width(36).
			MaxWidth(36).
			Border(lipgloss.NormalBorder(), false, false, true, false). // Only the bottom...
			BorderForeground(style.ColorGray)

	selectedListItem          = listItem.Foreground(style.ColorWhite).Background(style.ColorPurple)
	unfocusedSelectedListItem = selectedListItem.
					Foreground(style.ColorWhite).
					Background(style.ColorGray)
	noListItem = lipgloss.NewStyle().Width(36).Align(lipgloss.Center).Underline(true)
)

func newCategoryItemDelegate() extendedItemDelegate {
	return &categoryItemDelegate{focused: false}
}

type extendedItemDelegate interface {
	list.ItemDelegate
	SetFocus(focus bool)
}

type categoryItemDelegate struct{ focused bool }

func (delegate *categoryItemDelegate) Height() int  { return 1 }
func (delegate *categoryItemDelegate) Spacing() int { return 0 }
func (delegate *categoryItemDelegate) Update(
	_ tea.Msg, _ *list.Model,
) tea.Cmd {
	return nil
}

func (delegate *categoryItemDelegate) Render(
	w io.Writer, m list.Model, index int, item list.Item,
) {
	category, ok := item.(entity.Category)
	if !ok {
		return
	}

	str := category.Name
	if len(str) > 36 {
		str = str[:33] + "…"
	}
	st := listItem

	if index == m.Index() {
		st = unfocusedSelectedListItem
		if delegate.focused {
			st = selectedListItem
		}
	}

	if index == 0 {
		st = st.BorderTop(true)
	}

	_, _ = fmt.Fprint(w, st.Render(str))
}

func (delegate *categoryItemDelegate) SetFocus(focus bool) {
	delegate.focused = focus
}

const (
	maxListHeight  = 15
	maxTableHeight = 20
)

func newListDelegate(items []list.Item) (list.Model, extendedItemDelegate) {
	delegate := newCategoryItemDelegate()
	lst := list.New(items, delegate, 36, 15)
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.SetShowHelp(false)
	lst.SetShowHelp(false)
	lst.SetShowTitle(false)
	lst.SetShowPagination(true)
	lst.DisableQuitKeybindings()
	lst.SetStatusBarItemName("category", "categories")
	lst.Styles.NoItems = noListItem
	lst.Styles.PaginationStyle = listPagination

	return lst, delegate
}

var (
	titleStyle = lipgloss.NewStyle().
			Background(style.ColorPurple).
			Foreground(style.ColorBlack).
			MarginBottom(2).
			Padding(0, 1)
	unfocusedTitleStyle = titleStyle.Background(style.ColorGray).
				Foreground(style.ColorBlack)
)

func newTableColumns(width int) []table.Column {
	return []table.Column{
		{Title: "Id", Width: 0},
		{Title: "CategoryId", Width: 0},
		{Title: "Name", Width: width / 3},
		{Title: "Email", Width: width / 3},
		{Title: "Username", Width: width / 3},
		{Title: "Password", Width: 0},
	}
}

func newTable(rows []table.Row) table.Model {
	columns := newTableColumns(24 * 3)

	return table.New(
		table.WithColumns(columns), table.WithRows(rows),
		table.WithFocused(true),
		table.WithStyles(newTableStyle(true)))
}

func newTableStyle(focus bool) table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorWhite).
		BorderBottom(true).
		Bold(true)

	s.Cell = s.Cell.AlignHorizontal(lipgloss.Right)

	s.Selected = s.Selected.Bold(false).
		Foreground(style.ColorWhite).
		Background(style.ColorGray)
	if focus {
		s.Selected = s.Selected.
			Foreground(style.ColorBlack).
			Background(style.ColorPurple)
	}

	return s
}

func newSearch(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = "🔍: "
	ti.CharLimit = 16
	ti.Placeholder = placeholder
	ti.Cursor.SetMode(cursor.CursorStatic)

	return ti
}
