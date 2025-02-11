package library

import (
	"database/sql"
	"sort"

	"viscue/tui/component/list"
	"viscue/tui/component/table"
	"viscue/tui/entity"
	"viscue/tui/style"
	"viscue/tui/tool/cache"
	"viscue/tui/views/library/prompt"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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
	list        list.Model
	listHelp    help.Model
	listFilter  textinput.Model
	table       table.Model
	tableHelp   help.Model
	tableFilter textinput.Model
	spinner     spinner.Model

	// Dynamic styles
	container       lipgloss.Style
	listPaneBorder  lipgloss.Style
	tablePaneBorder lipgloss.Style

	// Data
	categories    []entity.Category
	passwords     []entity.Password
	height, width int
	focusedPane   lipgloss.Position
	err           error
	loaded        bool
}

func New(db *sqlx.DB) tea.Model {
	height := style.CalculateAppHeight() - 2
	width := cache.Get[int](cache.TerminalWidth) - 6
	container := lipgloss.NewStyle().
		Height(height).
		Align(lipgloss.Center, lipgloss.Top)
	paneBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Top, lipgloss.Top).
		Height(height).
		MaxHeight(height + 2).
		Padding(1)

	model := &library{
		db: db,
		// Components
		list:       list.New(),
		listHelp:   help.New(),
		listFilter: newSearch("Search category..."),
		table: table.New(
			table.WithColumns(
				[]table.Column{
					{Title: "Id", Width: 0},
					{Title: "CategoryId", Width: 0},
					{Title: "Name", Width: 24},
					{Title: "Email", Width: 24},
					{Title: "Username", Width: 24},
					{Title: "Password", Width: 0},
				}),
			table.WithFocused(true),
		),
		tableHelp:   help.New(),
		tableFilter: newSearch("Search password..."),
		spinner:     style.NewSpinner(),
		// Dynamic styles
		container:       container,
		listPaneBorder:  paneBorder,
		tablePaneBorder: paneBorder,
		// Data
		height:      height,
		width:       width,
		focusedPane: lipgloss.Right,
	}

	model.calculateDimension(width, height)

	return model
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
		case m.listFilter.Focused():
			m.listFilter, cmd = m.listFilter.Update(msg)
		case m.tableFilter.Focused():
			m.tableFilter, cmd = m.tableFilter.Update(msg)
		}
		return m, cmd
	case tea.WindowSizeMsg:
		// NOTE: The main app has saved this size in cache
		// before passing to its submodels.
		m.calculateDimension(msg.Width-6, style.CalculateAppHeight()-2)
		return m, nil
	case tea.KeyMsg:
		if m.prompt != nil {
			var cmd tea.Cmd
			m.prompt, cmd = m.prompt.Update(msg)
			return m, cmd
		}

		switch m.focusedPane {
		case lipgloss.Left:
			if m.listFilter.Focused() {
				switch {
				case key.Matches(msg, keys.Escape),
					key.Matches(msg, keys.Enter):
					m.blurSearch()
					return m, nil
				}
				var cmd tea.Cmd
				m.listFilter, cmd = m.listFilter.Update(msg)
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
					m.listFilter.Reset()
					m.setItems()
					m.setRows()
					return m, nil
				case key.Matches(msg, keys.Switch),
					key.Matches(msg, keys.Enter):
					m.switchFocus(lipgloss.Right)
					return m, nil
				}
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				m.setRows()
				return m, cmd
			}
		case lipgloss.Right:
			if m.tableFilter.Focused() {
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
				m.tableFilter, cmd = m.tableFilter.Update(msg)
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
					m.tableFilter.Reset()
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

	if m.err != nil {
		errTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(style.ColorRed).
			Render("Oops! Something went wrong.")
		errDesc := lipgloss.NewStyle().
			Italic(true).
			Foreground(style.ColorRedPale).
			Render(m.err.Error())
		return m.container.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				errTitle,
				errDesc,
			),
		)
	}

	// List Pane -- Category List
	listTitleStyle := unfocusedTitleStyle
	if m.focusedPane == lipgloss.Left {
		listTitleStyle = titleStyle
	}
	categoryTitle := listTitleStyle.Render("Category")
	searchBox = searchBox.Width(m.list.Width())
	if m.listFilter.Focused() {
		searchBox = searchBox.BorderForeground(style.ColorPurple)
	} else {
		searchBox = searchBox.BorderForeground(style.ColorGray)
	}
	categoryView := m.listPaneBorder.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		categoryTitle,
		searchBox.Render(m.listFilter.View()),
		m.list.View(),
	))

	// Table Pane -- Password Table
	tableTitleStyle := unfocusedTitleStyle
	if m.focusedPane == lipgloss.Right {
		tableTitleStyle = titleStyle
	}
	tableTitle := tableTitleStyle.Render("Password")
	searchBox = searchBox.Width(m.table.Width() - 4)
	if m.tableFilter.Focused() {
		searchBox = searchBox.BorderForeground(style.ColorPurple)
	} else {
		searchBox = searchBox.BorderForeground(style.ColorGray)
	}
	passwordView := m.tablePaneBorder.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			tableTitle,
			searchBox.Render(m.tableFilter.View()),
			m.table.View(),
		),
	)

	// Help View
	var helpView string
	if m.focusedPane == lipgloss.Left {
		helpView = m.listHelp.View(keys)
	} else {
		helpView = m.tableHelp.View(keys)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.container.Render(
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

// calculateDimension calculates the width and height for
// both list and table components based on the given.
// It also sets values for dynamic styles.
func (m *library) calculateDimension(width int, height int) {
	m.height = height
	m.width = width

	listWidth := width * 20 / 100
	listPaneWidth := listWidth + 4
	m.list.SetHeight(height - 8)
	m.list.SetWidth(listWidth)
	m.listFilter.Width = listWidth - 11
	m.listPaneBorder = m.listPaneBorder.Width(listPaneWidth).
		Height(height).
		MaxHeight(height + 2)

	tableWidth := width * 60 / 100
	tablePaneWidth := tableWidth + 4
	columnWidth := (tableWidth - 8) / 3
	m.table.SetHeight(height - 8)
	m.table.SetWidth(tableWidth)
	m.table.SetColumnsWidth(0, 0, columnWidth, columnWidth, columnWidth, 0)
	m.tableFilter.Width = tableWidth - 11
	m.tablePaneBorder = m.tablePaneBorder.Width(tablePaneWidth).
		Height(height).
		MaxHeight(height + 2)

	m.container = m.container.Height(height)
}

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

func (m *library) setItems() {
	m.list.SetItems(
		lo.Map(m.categories,
			func(item entity.Category, index int) list.Item {
				return item
			},
		),
	)
}

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
		m.list.Focus()
	case lipgloss.Right:
		m.table.Focus()
		m.list.Blur()
	}
}

func (m *library) focusSearch() {
	switch m.focusedPane {
	case lipgloss.Left:
		m.listFilter.Focus()
		m.listFilter.Cursor.SetMode(cursor.CursorBlink)
	case lipgloss.Right:
		m.tableFilter.Focus()
		m.tableFilter.Cursor.SetMode(cursor.CursorBlink)
	}
}

func (m *library) blurSearch() {
	switch m.focusedPane {
	case lipgloss.Left:
		m.listFilter.Blur()
		m.listFilter.Cursor.SetMode(cursor.CursorStatic)
	case lipgloss.Right:
		m.tableFilter.Blur()
		m.tableFilter.Cursor.SetMode(cursor.CursorStatic)
	}
}

func (m *library) applyTableSearch() {
	m.setRows() // TODO: Optimize this, perhaps store the current rows in *library...
	if m.tableFilter.Value() == "" {
		return
	}

	ranks := fuzzy.Find(m.tableFilter.Value(),
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
}

func (m *library) applyListSearch() {
	m.setItems()
	if m.listFilter.Value() == "" {
		return
	}

	ranks := fuzzy.Find(m.listFilter.Value(),
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
}

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
