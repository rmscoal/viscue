package library

import (
	"database/sql"
	"sort"

	"viscue/tui/entity"
	"viscue/tui/event"
	"viscue/tui/style"
	"viscue/tui/views/library/prompt"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

type library struct {
	db *sqlx.DB

	prompt tea.Model

	spinner   spinner.Model
	list      list.Model
	table     table.Model
	tableHelp help.Model
	listHelp  help.Model

	categories  []entity.Category
	passwords   []entity.Password
	focusedPane lipgloss.Position
	err         error
	loaded      bool
}

func New(db *sqlx.DB) tea.Model {
	return &library{
		db:          db,
		spinner:     style.NewSpinner(),
		list:        newList(nil),
		table:       newTable(nil),
		tableHelp:   help.New(),
		listHelp:    help.New(),
		focusedPane: lipgloss.Right,
	}
}

func (m *library) Init() tea.Cmd {
	return tea.Sequence(m.spinner.Tick, m.load)
}

func (m *library) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debugf("(*library).Update: received message of type %T", msg)
	defer func() {
		if len(m.table.Rows()) == 0 {
			m.table.SetHeight(2) // I don't know why setting this to 0 causes error when data is empty...
		} else {
			m.table.SetHeight(10)
		}
	}()

	switch msg := msg.(type) {
	case loadedData:
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
	case spinner.TickMsg:
		if m.loaded {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case cursor.BlinkMsg:
		if m.prompt != nil {
			var cmd tea.Cmd
			m.prompt, cmd = m.prompt.Update(msg)
			return m, cmd
		}
	case tea.KeyMsg:
		if m.prompt != nil {
			var cmd tea.Cmd
			m.prompt, cmd = m.prompt.Update(msg)
			return m, cmd
		}
		switch m.focusedPane {
		case lipgloss.Left:
			switch {
			case key.Matches(msg, listKeys.Add):
				m.prompt = prompt.NewCategory(m.db, entity.Category{})
				return m, m.prompt.Init()
			case key.Matches(msg, listKeys.FocusRight):
				m.focusedPane = lipgloss.Right
				m.table.Focus()
				return m, nil
			default:
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				m.setRows()
				return m, cmd
			}
		case lipgloss.Right:
			switch {
			case key.Matches(msg, tableKeys.Add):
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
				return m, m.prompt.Init()
			case key.Matches(msg, tableKeys.Edit):
				password, err := entity.NewPasswordFromTableRow(m.table.SelectedRow())
				if err != nil {
					log.Fatal(err)
				}
				m.prompt = prompt.NewPassword(m.db, password)
				return m, m.prompt.Init()
			case key.Matches(msg, tableKeys.FocusLeft):
				m.focusedPane = lipgloss.Left
				m.table.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.table, cmd = m.table.Update(msg)
				return m, cmd
			}
		}
	case event.LibraryMessage:
		switch msg {
		case event.ClosePrompt:
			m.prompt = nil
			return m, nil
		}
	case prompt.SubmitData:
		switch model := msg.(type) {
		case entity.Category:
			_, index, found := lo.FindIndexOf(m.categories,
				func(item entity.Category) bool {
					return item.Id == model.Id
				})
			if !found {
				all := m.categories[0]
				uncategorized := m.categories[len(m.categories)-1]
				m.categories = m.categories[1 : len(m.categories)-1]
				m.categories = append(m.categories, model)
				sort.SliceStable(m.categories, func(i, j int) bool {
					return m.categories[i].Name < m.categories[j].Name
				})
				m.categories = append([]entity.Category{all}, m.categories...)
				m.categories = append(m.categories, uncategorized)
			} else {
				m.categories[index] = model
			}
			m.list.SetItems(lo.Map(m.categories,
				func(item entity.Category, index int) list.Item {
					return item
				},
			))
			return m, m.closePrompt
		case entity.Password:
			_, index, found := lo.FindIndexOf(m.passwords,
				func(item entity.Password) bool {
					return item.Id == model.Id
				})
			if !found {
				m.passwords = append(m.passwords, model)
				sort.SliceStable(m.passwords, func(i, j int) bool {
					return m.passwords[i].Id < m.passwords[j].Id
				})
			} else {
				m.passwords[index] = model
			}
			m.setRows()
			return m, m.closePrompt
		}
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
}
