package library

import (
	"fmt"
	"io"
	"strings"

	"viscue/tui/entity"
	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	listPagination            = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	listItem                  = lipgloss.NewStyle().PaddingLeft(4).Foreground(style.ColorGray)
	selectedListItem          = lipgloss.NewStyle().PaddingLeft(2).Foreground(style.ColorPurple)
	unfocusedSelectedListItem = selectedListItem.Foreground(style.ColorWhite)
)

func NewCategoryItemDelegate() ExtendedItemDelegate {
	return &categoryItemDelegate{focused: false}
}

type ExtendedItemDelegate interface {
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

	str := " " + category.Name
	fn := listItem.Render

	if index == m.Index() {
		fn = func(s ...string) string {
			render := unfocusedSelectedListItem.Render
			if delegate.focused {
				render = selectedListItem.Render
			}
			return render("➡ " + strings.Join(s, " "))
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

func (delegate *categoryItemDelegate) SetFocus(focus bool) {
	delegate.focused = focus
}

func newList(items []list.Item) list.Model {
	lst := list.New(items, NewCategoryItemDelegate(), 15, 10)
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(true)
	lst.SetShowHelp(false)
	lst.SetShowPagination(true)
	lst.Title = "Categories"
	lst.Styles.PaginationStyle = listPagination

	return lst
}

var (
	tableTitle = lipgloss.NewStyle().
			Background(style.ColorPurple).
			Foreground(style.ColorBlack).
			MarginBottom(1).
			Padding(0, 1).
			SetString("Passwords").
			Render
	tableBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(style.ColorWhite).
			Render
)

func newTable(rows []table.Row) table.Model {
	columns := []table.Column{
		{Title: "Id", Width: 0},
		{Title: "CategoryId", Width: 0},
		{Title: "Name", Width: 10},
		{Title: "Email", Width: 24},
		{Title: "Username", Width: 24},
		{Title: "Password", Width: 24},
		{Title: "Actual Password", Width: 0},
	}

	return table.New(
		table.WithColumns(columns), table.WithRows(rows),
		table.WithFocused(true), table.WithHeight(10),
		table.WithStyles(newTableStyle(true)))
}

func newTableStyle(focus bool) table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorWhite).
		BorderBottom(true).
		Bold(true)

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
