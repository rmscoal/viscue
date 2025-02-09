package library

import (
	"database/sql"
	"sort"

	"viscue/tui/entity"
	"viscue/tui/style"
	"viscue/tui/tool/cache"
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
	listFilterInput  textinput.Model
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
		listFilterInput:  newSearch("Search category..."),
		table:            newTable(nil),
		tableHelp:        help.New(),
		tableFilterInput: newSearch("Search password..."),
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
	defer m.setKeys()

	switch msg := msg.(type) {
	case DataLoadedMsg:
		m.categories = msg.Categories
		m.passwords = msg.Passwords
		m.loaded = true
		m.setItems()
		m.list.Select(0) // Select "All" category
		m.setRows()
		return m, nil
	case prompt.DataSubmittedMsg[entity.Password]:
		defer m.closePrompt()
		m.appendPassword(msg.Model)
		return m, nil
	case prompt.DataSubmittedMsg[entity.Category]:
		defer m.closePrompt()
		m.appendCategory(msg.Model)
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
		case m.listFilterInput.Focused():
			m.listFilterInput, cmd = m.listFilterInput.Update(msg)
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
			if m.listFilterInput.Focused() {
				switch {
				case key.Matches(msg, keys.Escape),
					key.Matches(msg, keys.Enter):
					m.blurSearch()
					return m, nil
				}
				var cmd tea.Cmd
				m.listFilterInput, cmd = m.listFilterInput.Update(msg)
				m.applyListSearch()
				return m, cmd
			} else {
				switch {
				case key.Matches(msg, keys.Add):
					m.prompt = prompt.NewCategory(m.db, entity.Category{})
					return m, m.prompt.Init()
				case key.Matches(msg, keys.Edit):
					selectedCategory := m.list.SelectedItem().(entity.Category)
					m.prompt = prompt.NewCategory(m.db, selectedCategory)
					return m, m.prompt.Init()
				case key.Matches(msg, keys.Delete):
				case key.Matches(msg, keys.Search):
					m.focusSearch()
					return m, nil
				case key.Matches(msg, keys.Clear):
					m.listFilterInput.Reset()
					m.setItems()
					m.setRows()
					return m, nil
				case key.Matches(msg, keys.Switch), key.Matches(msg,
					keys.Enter):
					m.switchFocus(lipgloss.Right)
					return m, nil
				}
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				m.setRows()
				return m, cmd
			}
		case lipgloss.Right:
			if m.tableFilterInput.Focused() {
				switch {
				case key.Matches(msg, keys.Enter), key.Matches(msg,
					keys.Escape):
					m.blurSearch()
					return m, nil
				case key.Matches(msg, keys.Switch):
					m.switchFocus(lipgloss.Left)
					return m, nil
				}
				var cmd tea.Cmd
				m.tableFilterInput, cmd = m.tableFilterInput.Update(msg)
				m.applyTableSearch()
				return m, cmd
			} else {
				switch {
				case key.Matches(msg, keys.Add):
					m.newPasswordPrompt()
					return m, m.prompt.Init()
				case key.Matches(msg, keys.Edit):
					m.editPasswordPrompt()
					return m, m.prompt.Init()
				case key.Matches(msg, keys.Delete):
				case key.Matches(msg, keys.Search):
					m.focusSearch()
					return m, textinput.Blink
				case key.Matches(msg, keys.Clear):
					m.tableFilterInput.Reset()
					m.applyTableSearch()
					return m, nil
				case key.Matches(msg, keys.Switch):
					m.switchFocus(lipgloss.Left)
					return m, nil
				}
				var cmd tea.Cmd
				m.table, cmd = m.table.Update(msg)
				return m, cmd
			}
		}
	}

	return m, nil
}

func (m *library) View() string {
	if !m.loaded {
		return m.spinner.View() + " " + "Opening vault... Please wait"
	} else if m.prompt != nil {
		return m.prompt.View()
	}

	height := style.CalculateAppHeight() - 2
	width := cache.Get[int](cache.TerminalWidth)
	libraryViewWidth := width - 6
	categoryPaneWidth := libraryViewWidth * 20 / 100
	passwordPaneWidth := libraryViewWidth * 60 / 100

	libraryContainer := lipgloss.NewStyle().
		Height(height).
		Align(lipgloss.Center, lipgloss.Top).
		Render
	searchBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorGray).
		Align(lipgloss.Left, lipgloss.Center).
		PaddingLeft(1).
		PaddingRight(1)
	paneBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center, lipgloss.Top).
		Height(height).
		MaxHeight(height + 2).
		Padding(1).
		Render

	if m.err != nil {
		errTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(style.ColorRed).
			Render("Oops! Something went wrong.")
		errDesc := lipgloss.NewStyle().
			Italic(true).
			Foreground(style.ColorRedPale).
			Render(m.err.Error())
		return libraryContainer(
			lipgloss.JoinVertical(
				lipgloss.Center,
				errTitle,
				errDesc,
			),
		)
	}

	// Category Pane
	categoryTitleStyle := unfocusedTitleStyle
	if m.focusedPane == lipgloss.Left {
		categoryTitleStyle = titleStyle
	}
	categoryTitle := categoryTitleStyle.Render("Category")
	listHeight := 2 * len(m.list.Items())
	listHeight = min(listHeight, height-18)
	m.list.SetSize(categoryPaneWidth, listHeight)
	listView := m.list.View()
	listWidth := lipgloss.Width(listView)
	searchBox = searchBox.Width(listWidth)
	if m.listFilterInput.Focused() {
		searchBox = searchBox.BorderForeground(style.ColorPurple)
	}
	categoryView := paneBorder(lipgloss.JoinVertical(
		lipgloss.Left,
		categoryTitle,
		searchBox.Render(m.listFilterInput.View()),
		listView,
	))

	// Password Pane
	passwordTitleStyle := unfocusedTitleStyle
	if m.focusedPane == lipgloss.Right {
		passwordTitleStyle = titleStyle
	}
	passwordTitle := passwordTitleStyle.Render("Password")
	tableHeight := 1*len(m.table.Rows()) + tableColumnHeight
	tableHeight = min(height-8, tableHeight)
	m.table.SetHeight(tableHeight)
	m.table.SetWidth(passwordPaneWidth)
	m.table.SetColumns(newTableColumns(passwordPaneWidth))
	tableView := m.table.View()
	if len(m.table.Rows()) == 0 {
		tableView = lipgloss.JoinVertical(
			lipgloss.Center,
			tableView,
			"No passwords found.",
		)
	}
	searchBox = searchBox.Width(m.table.Width() - 4)
	if m.tableFilterInput.Focused() {
		searchBox = searchBox.BorderForeground(style.ColorPurple)
	}
	passwordView := paneBorder(
		lipgloss.JoinVertical(
			lipgloss.Left,
			passwordTitle,
			searchBox.Render(m.tableFilterInput.View()),
			tableView,
		),
	)

	// Help
	var helpView string
	if m.focusedPane == lipgloss.Left {
		helpView = m.listHelp.View(keys)
	} else {
		helpView = m.listHelp.View(keys)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		libraryContainer(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				categoryView,
				passwordView,
			),
		),
		style.HelpContainer(helpView),
	)
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

// List

func (m *library) setItems() {
	m.list.SetItems(
		lo.Map(m.categories,
			func(item entity.Category, index int) list.Item {
				return item
			},
		),
	)
}

// Focus

func (m *library) setKeys() {
	switch m.focusedPane {
	case lipgloss.Left:
		keys.Switch = focusRight
	case lipgloss.Right:
		keys.Switch = focusLeft
	}
}

func (m *library) switchFocus(pos lipgloss.Position) {
	m.blurSearch()
	m.focusedPane = pos
	switch pos {
	case lipgloss.Left:
		m.table.Blur()
		m.delegate.SetFocus(true)
		m.table.SetStyles(newTableStyle(false))
	case lipgloss.Right:
		m.table.Focus()
		m.delegate.SetFocus(false)
		m.table.SetStyles(newTableStyle(true))
	}
}

// Search

func (m *library) focusSearch() {
	switch m.focusedPane {
	case lipgloss.Left:
		m.listFilterInput.Focus()
		m.listFilterInput.Cursor.SetMode(cursor.CursorBlink)
	case lipgloss.Right:
		m.tableFilterInput.Focus()
		m.tableFilterInput.Cursor.SetMode(cursor.CursorBlink)
	}
}

func (m *library) blurSearch() {
	switch m.focusedPane {
	case lipgloss.Left:
		m.listFilterInput.Blur()
		m.listFilterInput.Cursor.SetMode(cursor.CursorStatic)
	case lipgloss.Right:
		m.tableFilterInput.Blur()
		m.tableFilterInput.Cursor.SetMode(cursor.CursorStatic)
	}
}

func (m *library) applyTableSearch() {
	m.setRows() // TODO: Optimize this, perhaps store the current rows in *library...
	if m.tableFilterInput.Value() == "" {
		return
	}

	ranks := fuzzy.Find(m.tableFilterInput.Value(),
		lo.Map(m.table.Rows(), func(row table.Row, _ int) string {
			return row[2] // This is the name...
		}))
	sort.Stable(ranks)

	filteredIndexes := lo.Map(ranks, func(match fuzzy.Match, index int) int {
		return match.Index
	})
	filteredRows := lo.Filter(m.table.Rows(),
		func(item table.Row, index int) bool {
			return lo.Contains(filteredIndexes, index)
		})
	m.table.SetRows(filteredRows)
	m.table.SetCursor(0) // Select the highest ranked row
}

func (m *library) applyListSearch() {
	m.setItems()
	if m.listFilterInput.Value() == "" {
		return
	}

	ranks := fuzzy.Find(m.listFilterInput.Value(),
		lo.Map(m.list.Items(), func(item list.Item, _ int) string {
			return item.(entity.Category).Name
		}))
	sort.Stable(ranks)

	filteredIndexes := lo.Map(ranks, func(match fuzzy.Match, index int) int {
		return match.Index
	})
	filteredItems := lo.Filter(m.list.Items(),
		func(item list.Item, index int) bool {
			return lo.Contains(filteredIndexes, index)
		})
	m.list.SetItems(filteredItems)
	m.list.Select(0) // Select the highest ranked item
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

func (m *library) appendPassword(pw entity.Password) {
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

func (m *library) appendCategory(cat entity.Category) {
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
