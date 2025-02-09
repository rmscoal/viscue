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

const (
	tableMaxHeight    = 20
	tableColumnHeight = 2

	maxListHeight      = 12
	maxTableHeight     = 20
	columnDefaultWidth = 24
	maxItemWidth       = 48
)

// List styles
var (
	listPagination = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	listItem       = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(style.ColorWhite).
			MaxWidth(maxItemWidth).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(style.ColorGray)
	selectedListItem          = listItem.Background(style.ColorPurple)
	unfocusedSelectedListItem = listItem.Background(style.ColorGray)
	noListItem                = lipgloss.NewStyle().
					Width(maxItemWidth).
					Align(lipgloss.Center).
					Underline(true)
)

// Title styles
var (
	titleStyle = lipgloss.NewStyle().
			Background(style.ColorPurple).
			Foreground(style.ColorBlack).
			MarginBottom(2).
			Padding(0, 1)
	unfocusedTitleStyle = titleStyle.Background(style.ColorGray)
)

// Category delegate
type (
	extendedItemDelegate interface {
		list.ItemDelegate
		SetFocus(focus bool)
	}

	categoryItemDelegate struct {
		focused bool
	}
)

func newCategoryItemDelegate() extendedItemDelegate {
	return &categoryItemDelegate{focused: false}
}

func (d *categoryItemDelegate) Height() int                             { return 1 }
func (d *categoryItemDelegate) Spacing() int                            { return 0 }
func (d *categoryItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d *categoryItemDelegate) SetFocus(focus bool)                     { d.focused = focus }

func (d *categoryItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	category, ok := item.(entity.Category)
	if !ok {
		return
	}

	str := category.Name
	if len(str) > m.Width() {
		str = str[:m.Width()-3] + "…"
	}

	st := listItem.Width(m.Width())
	if index == m.Index() {
		if d.focused {
			st = selectedListItem.Width(m.Width())
		} else {
			st = unfocusedSelectedListItem.Width(m.Width())
		}
	}

	fmt.Fprint(w, st.Render(str))
}

// List component
func newListDelegate(items []list.Item) (list.Model, extendedItemDelegate) {
	delegate := newCategoryItemDelegate()
	lst := list.New(items, delegate, maxItemWidth, maxListHeight)

	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.SetShowHelp(false)
	lst.SetShowTitle(false)
	lst.SetShowPagination(false)
	lst.DisableQuitKeybindings()
	lst.SetStatusBarItemName("category", "categories")

	lst.Styles.NoItems = noListItem
	lst.Styles.PaginationStyle = listPagination

	return lst, delegate
}

// Table component
func newTableColumns(width int) []table.Column {
	width = (width - 8) / 3 // Account for borders and padding
	return []table.Column{
		{Title: "Id", Width: 0},
		{Title: "CategoryId", Width: 0},
		{Title: "Name", Width: width},
		{Title: "Email", Width: width},
		{Title: "Username", Width: width},
		{Title: "Password", Width: 0},
	}
}

func newTable(rows []table.Row) table.Model {
	columns := newTableColumns(columnDefaultWidth * 3)
	return table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithStyles(newTableStyle(true)),
	)
}

func newTableStyle(focus bool) table.Styles {
	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorWhite).
		BorderBottom(true).
		Bold(true)

	s.Selected = s.Selected.
		Bold(false).
		Foreground(style.ColorWhite).
		Background(style.ColorGray)

	if focus {
		s.Selected = s.Selected.
			Foreground(style.ColorBlack).
			Background(style.ColorPurple)
	}

	return s
}

// Search component
func newSearch(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = "🔍: "
	ti.Placeholder = placeholder
	ti.Cursor.SetMode(cursor.CursorStatic)
	return ti
}
