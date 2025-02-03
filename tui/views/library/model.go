package library

import (
	"database/sql"
	"sort"
	"strings"

	"viscue/tui/entity"
	"viscue/tui/style"
	"viscue/tui/views/library/prompt"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	"github.com/sahilm/fuzzy"
	"github.com/samber/lo"
)

type library struct {
	db *sqlx.DB

	// Sub-model
	prompt tea.Model

	// Components
	list             list.Model
	listHelp         help.Model
	table            table.Model
	tableHelp        help.Model
	tableFilterInput textinput.Model
	spinner          spinner.Model
	delegate         extendedItemDelegate

	// Data
	categories  []entity.Category
	passwords   []entity.Password
	focusedPane lipgloss.Position
	err         error
	loaded      bool
}

func New(db *sqlx.DB) tea.Model {
	lst, delegate := newListDelegate(nil)
	return &library{
		db: db,
		// Components
		list:             lst,
		listHelp:         help.New(),
		table:            newTable(nil),
		tableHelp:        help.New(),
		tableFilterInput: newFilter(""),
		spinner:          style.NewSpinner(),
		delegate:         delegate,
		// Data
		focusedPane: lipgloss.Right,
	}
}

func (m *library) Init() tea.Cmd {
	return tea.Sequence(m.spinner.Tick, m.load)
}

func (m *library) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debugf("(*library).Update: received message type %T", msg)
	defer m.updateTableHeight()

	switch msg := msg.(type) {
	case DataLoadedMsg:
		m.categories = msg.Categories
		m.passwords = msg.Passwords
		m.loaded = true
		m.list.SetItems(
			lo.Map(m.categories,
				func(item entity.Category, index int) list.Item {
					return item
				},
			),
		)
		m.table.SetRows(
			// When data is emitted, the first selected category will be "All".
			// Hence, we just map all our passwords entity into row.
			lo.Map(m.passwords,
				func(item entity.Password, index int) table.Row {
					return item.ToTableRow()
				},
			),
		)
		return m, nil
	case prompt.DataSubmittedMsg[entity.Password]:
		defer m.closePrompt()
		m.updateTableRecord(msg.Model)
		return m, nil
	case prompt.DataSubmittedMsg[entity.Category]:
		defer m.closePrompt()
		m.updateListRecord(msg.Model)
		return m, nil
	case prompt.CloseMsg:
		m.closePrompt()
		return m, nil
	case spinner.TickMsg:
		if m.loaded {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case cursor.BlinkMsg:
		var cmd tea.Cmd
		switch {
		case m.prompt != nil:
			m.prompt, cmd = m.prompt.Update(msg)
		case m.list.FilterState() == list.Filtering:
			m.list.FilterInput, cmd = m.list.FilterInput.Update(msg)
		case m.tableFilterInput.Focused():
			m.tableFilterInput, cmd = m.tableFilterInput.Update(msg)
		}
		return m, cmd
	case tea.KeyMsg:
		if m.prompt != nil {
			var cmd tea.Cmd
			m.prompt, cmd = m.prompt.Update(msg)
			return m, cmd
		}

		switch m.focusedPane {
		case lipgloss.Left:
			switch {
			case m.list.FilterState() == list.Filtering:
				break
			case key.Matches(msg, listKeys.Add):
				m.prompt = prompt.NewCategory(m.db, entity.Category{})
				return m, m.prompt.Init()
			case key.Matches(msg, listKeys.ClearFilter):
				m.list.ResetFilter()
				m.setRows()
				return m, nil
			case key.Matches(msg, listKeys.FocusRight):
				m.focusPane(lipgloss.Right)
				return m, nil
			}
		case lipgloss.Right:
			if m.tableFilterInput.Focused() {
				switch {
				case key.Matches(msg, tableKeys.SubmitFilter):
					m.blurTableFilter()
					return m, nil
				case key.Matches(msg, tableKeys.CancelFilter):
					m.blurTableFilter()
					return m, nil
				}
			} else {
				switch {
				case key.Matches(msg, tableKeys.Add):
					m.newPasswordPrompt()
					return m, m.prompt.Init()
				case key.Matches(msg, tableKeys.Edit):
					m.editPasswordPrompt()
					return m, m.prompt.Init()
				case key.Matches(msg, tableKeys.FocusLeft):
					m.focusPane(lipgloss.Left)
					return m, nil
				case key.Matches(msg, tableKeys.Filter):
					m.focusTableFilter()
					return m, textinput.Blink
				case key.Matches(msg, tableKeys.ClearFilter):
					m.tableFilterInput.Reset()
					m.filterTable()
					return m, nil
				}
			}
		}
	}

	// Pass another type of message to either table or list component
	switch m.focusedPane {
	case lipgloss.Left:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		m.setRows()
		return m, cmd
	case lipgloss.Right:
		var cmd tea.Cmd
		// When table filter input is focused,
		// we want to pass the message there instead...
		if m.tableFilterInput.Focused() {
			m.tableFilterInput, cmd = m.tableFilterInput.Update(msg)
			m.filterTable()
		} else {
			m.table, cmd = m.table.Update(msg)
		}
		return m, cmd
	}

	// TODO: Error view...

	return m, nil
}

func (m *library) View() string {
	if m.err != nil {
		return "Error: " + style.ErrorText(m.err.Error())
	}

	if !m.loaded {
		return m.spinner.View() + " " + "Opening vault... Please wait"
	}

	if m.prompt != nil {
		return m.prompt.View()
	}

	// Category
	categoryView := m.list.View()

	// Passwords
	passwordView := tableTitle() + "\n"
	passwordView += m.tableFilterInput.View() + strings.Repeat("\n", 2)
	tableView := m.table.View()
	if len(m.table.Rows()) == 0 {
		tableView = lipgloss.JoinVertical(lipgloss.Center,
			tableView, "No data")
	}
	passwordView += tableView

	libraryView := lipgloss.JoinHorizontal(lipgloss.Top, categoryView,
		passwordView)

	// Help
	var helpView string
	if m.focusedPane == lipgloss.Left {
		helpView = m.listHelp.View(listKeys)
	} else {
		helpView = m.listHelp.View(tableKeys)
	}

	return lipgloss.JoinVertical(lipgloss.Center, libraryView, helpView)
}

//////////////////////////////////
//////////// PRIVATE /////////////
//////////////////////////////////

// Table

func (m *library) setRows() {
	category := m.list.SelectedItem().(entity.Category)

	var rows []table.Row
	switch category.Id {
	case 0:
		rows = lo.Map(m.passwords,
			func(item entity.Password, index int) table.Row {
				return item.ToTableRow()
			})
	case -1:
		rows = lo.FilterMap(m.passwords,
			func(item entity.Password, index int) (table.Row, bool) {
				return item.ToTableRow(), !item.CategoryId.Valid
			})
	default:
		rows = lo.FilterMap(m.passwords,
			func(item entity.Password, index int) (table.Row, bool) {
				return item.ToTableRow(),
					item.CategoryId.Int64 == category.Id
			})
	}
	m.table.SetRows(rows)
	m.table.SetCursor(0)
}

func (m *library) updateTableHeight() {
	if len(m.table.Rows()) == 0 {
		// I don't know why setting this to 0 causes error when data is empty...
		m.table.SetHeight(2)
	} else {
		m.table.SetHeight(10)
	}
}

func (m *library) updateTableRecord(pw entity.Password) {
	_, index, found := lo.FindIndexOf(m.passwords,
		func(item entity.Password) bool {
			return item.Id == pw.Id
		})
	if !found {
		m.passwords = append(m.passwords, pw)
		sort.SliceStable(m.passwords, func(i, j int) bool {
			return m.passwords[i].Id < m.passwords[j].Id
		})
	} else {
		m.passwords[index] = pw
	}
	m.setRows()
}

// List

func (m *library) updateListRecord(cat entity.Category) {
	_, index, found := lo.FindIndexOf(m.categories,
		func(item entity.Category) bool {
			return item.Id == cat.Id
		})
	if !found {
		all := m.categories[0]
		uncategorized := m.categories[len(m.categories)-1]
		m.categories = m.categories[1 : len(m.categories)-1]
		m.categories = append(m.categories, cat)
		sort.SliceStable(m.categories, func(i, j int) bool {
			return m.categories[i].Name < m.categories[j].Name
		})
		m.categories = append([]entity.Category{all}, m.categories...)
		m.categories = append(m.categories, uncategorized)
	} else {
		m.categories[index] = cat
	}
	m.list.SetItems(lo.Map(m.categories,
		func(item entity.Category, index int) list.Item {
			return item
		},
	))
}

// Pane

func (m *library) focusPane(pos lipgloss.Position) {
	m.focusedPane = pos
	switch pos {
	case lipgloss.Left:
		m.table.Blur()
		m.delegate.SetFocus(true)
		m.table.SetStyles(newTableStyle(false))
		m.blurTableFilter() // Disable filter input
	case lipgloss.Right:
		m.table.Focus()
		m.delegate.SetFocus(false)
		m.table.SetStyles(newTableStyle(true))
	}
}

// Filters

func (m *library) focusTableFilter() {
	m.tableFilterInput.Focus()
	m.tableFilterInput.Cursor.SetMode(cursor.CursorBlink)
}

func (m *library) blurTableFilter() {
	m.tableFilterInput.Blur()
	m.tableFilterInput.Cursor.SetMode(cursor.CursorStatic)
}

func (m *library) filterTable() {
	m.setRows() // TODO: Optimize this, perhaps store the current rows in *library...
	if m.tableFilterInput.Value() == "" {
		return
	}

	ranks := fuzzy.Find(m.tableFilterInput.Value(),
		lo.Map(m.table.Rows(), func(row table.Row, _ int) string {
			return row[2] // This is the name...
		}))
	sort.Stable(ranks)
	log.Debugf("(*library).filterTable: ranks %+v", ranks)

	filteredIndexes := lo.Map(ranks, func(match fuzzy.Match, index int) int {
		return match.Index
	})
	filteredRows := lo.Filter(m.table.Rows(),
		func(item table.Row, index int) bool {
			return lo.Contains(filteredIndexes, index)
		})
	log.Debugf("(*library).filterTable: filteredRows %+v", filteredRows)
	m.table.SetRows(filteredRows)
	m.table.SetCursor(0)
}

// Prompt

func (m *library) closePrompt() {
	m.prompt = nil
}

func (m *library) newPasswordPrompt() {
	categoryId := sql.NullInt64{
		Int64: m.list.SelectedItem().(entity.Category).Id,
		Valid: true,
	}
	if categoryId.Int64 == 0 || categoryId.Int64 == -1 {
		categoryId = sql.NullInt64{Int64: 0, Valid: false}
	}
	m.prompt = prompt.NewPassword(m.db, entity.Password{
		CategoryId: categoryId,
	})
}

func (m *library) editPasswordPrompt() {
	password, err := entity.NewPasswordFromTableRow(m.table.SelectedRow())
	if err != nil {
		log.Fatal(err)
	}
	m.prompt = prompt.NewPassword(m.db, password)
}
