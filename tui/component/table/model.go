package table

import (
	"viscue/tui/style"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	_defaultHeaderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(style.ColorNormal).
				PaddingLeft(1).
				Bold(true)
	_defaultCellStyle         = lipgloss.NewStyle().PaddingLeft(1).Foreground(style.ColorNormal)
	_defaultBlurredCellStyle  = _defaultCellStyle.Foreground(style.ColorGray)
	_defaultSelectedCellStyle = lipgloss.NewStyle().
					Foreground(style.ColorNormal).
					Background(style.ColorPurple)
	_defaultBlurredSelectedCellStyle = _defaultSelectedCellStyle.Background(style.ColorGray)
)

type Row []string

type Column struct {
	Title string
	Width int
}

type Model struct {
	vp      viewport.Model
	rows    []Row
	columns []Column
	currIdx int
	focused bool
}

func New(opts ...Option) Model {
	model := Model{
		vp: viewport.New(0, 0),
	}
	for _, opt := range opts {
		opt(&model)
	}
	return model
}

// Options

type Option func(*Model)

func WithRows(rows []Row) Option {
	return func(m *Model) {
		m.SetRows(rows)
	}
}

func WithColumns(columns []Column) Option {
	return func(m *Model) {
		m.SetColumns(columns)
	}
}

func WithHeight(height int) Option {
	return func(m *Model) {
		m.SetHeight(height)
	}
}

func WithWidth(width int) Option {
	return func(m *Model) {
		m.SetWidth(width)
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
	return m.vp.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.Down()
			return m, nil
		case "k", "up":
			m.Up()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

// View should display header and viewport that joined vertically as a table.
func (m Model) View() string {
	m.vp.SetContent(m.renderRows())
	return lipgloss.JoinVertical(lipgloss.Top, m.renderHeader(), m.vp.View())
}

// renderHeader renders the header of the table from columns
func (m Model) renderHeader() string {
	var headers []string
	for _, col := range m.columns {
		if col.Width <= 0 {
			continue
		}
		str := col.Title
		width := lipgloss.Width(str)
		if width > col.Width && col.Width > 0 {
			str = str[:col.Width-2] + "…"
		}
		headers = append(headers,
			_defaultHeaderStyle.Width(col.Width).MaxWidth(col.Width).Render(str))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, headers...)
}

func (m Model) cellStyle(index, width int) lipgloss.Style {
	cellStyle := _defaultBlurredCellStyle
	selectedStyle := _defaultBlurredSelectedCellStyle

	if m.focused {
		cellStyle = _defaultCellStyle
		selectedStyle = _defaultSelectedCellStyle
	}

	st := cellStyle.Width(width).MaxWidth(width)
	if m.currIdx == index {
		// TODO: Perhaps make a util function to enhance `Inherit`
		if _defaultSelectedCellStyle.GetForeground() == nil {
			cellStyle = cellStyle.UnsetForeground()
		}
		st = st.UnsetForeground().Inherit(selectedStyle)
	}

	return st
}

// renderRows renders all the rows that will be set as viewport content
func (m Model) renderRows() string {
	var rows []string
	for rowIndex, row := range m.rows {
		var cells []string
		for columnIndex := 0; columnIndex < len(row) &&
			columnIndex < len(m.columns); columnIndex++ {
			cell := row[columnIndex]
			columnWidth := m.columns[columnIndex].Width
			if columnWidth <= 0 {
				continue
			} else if len(cell)-1 >= columnWidth {
				cell = cell[:columnWidth-3] + "…"
			}
			cellStyle := m.cellStyle(rowIndex, columnWidth)
			cells = append(cells, cellStyle.Render(cell))
		}
		str := lipgloss.JoinHorizontal(lipgloss.Left, cells...)
		rows = append(rows, str)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// SetRows replace the row field and reset the cursor to 0
func (m *Model) SetRows(rows []Row) {
	m.rows = rows
	m.vp.SetContent(m.renderRows())
	m.currIdx = 0
	m.vp.SetYOffset(0)
}

func (m Model) Rows() []Row {
	return m.rows
}

// SelectedRow returns the row where the cursor is at
func (m Model) SelectedRow() Row {
	if m.currIdx > len(m.rows)-1 {
		return nil
	}
	return m.rows[m.currIdx]
}

// SetColumns replace the columns field
func (m *Model) SetColumns(columns []Column) {
	m.columns = columns
}

func (m Model) Columns() []Column {
	return m.columns
}

// SetColumnsWidth sets each of the column width based on given widths sequentially
func (m *Model) SetColumnsWidth(widths ...int) {
	length := min(len(widths), len(m.columns))
	for i := 0; i < length; i++ {
		m.columns[i].Width = widths[i]
	}
}

func (m *Model) SetHeight(height int) {
	headerHeight := lipgloss.Height(m.renderHeader())
	m.vp.Height = max(0, height-headerHeight)
}

func (m Model) Height() int {
	headerHeight := lipgloss.Height(m.renderHeader())
	return m.vp.Height + headerHeight
}

func (m *Model) SetWidth(width int) {
	m.vp.Width = width
}

func (m Model) Width() int {
	return m.vp.Width
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

func (m *Model) Up() {
	if m.currIdx <= 0 {
		return
	}

	m.currIdx--
	// The height of the current row is currIdx + 1,
	// assuming all rows have height of 1
	currRowYPos := m.currIdx + 1
	for currRowYPos <= m.vp.YOffset && !m.vp.AtTop() {
		m.vp.LineUp(1)
	}
}

func (m *Model) Down() {
	length := len(m.rows)
	if length == 0 || m.currIdx >= length-1 {
		return
	}

	m.currIdx++
	// The height of the current row is currIdx + 1,
	// assuming all rows have height of 1
	currRowYPos := m.currIdx + 1
	for currRowYPos > m.vp.Height+m.vp.YOffset && !m.vp.AtBottom() {
		m.vp.LineDown(1)
	}
}
