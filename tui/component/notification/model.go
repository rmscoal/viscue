package notification

import (
	"time"

	"viscue/tui/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Position struct {
	X lipgloss.Position
	Y lipgloss.Position
}

var (
	BottomRight = Position{
		X: lipgloss.Right,
		Y: lipgloss.Bottom,
	}
	BottomLeft = Position{
		X: lipgloss.Left,
		Y: lipgloss.Bottom,
	}
	TopRight = Position{
		X: lipgloss.Right,
		Y: lipgloss.Top,
	}
	TopLeft = Position{
		X: lipgloss.Left,
		Y: lipgloss.Top,
	}
)

type Model struct {
	Style    lipgloss.Style
	duration time.Duration
	position Position
	msg      string
	visible  bool
}

type TickMsg struct{}

type ShowMsg struct {
	Message string
}

func New(opts ...Option) Model {
	m := Model{
		Style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
			BorderForeground(style.ColorPurple).Padding(1, 2),
		position: BottomRight,
		duration: 1 * time.Second,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

type Option func(*Model)

// WithDuration sets the duration for which the notification is displayed
func WithDuration(duration time.Duration) Option {
	return func(m *Model) {
		m.SetDuration(duration)
	}
}

// WithPosition adjusts the position of the notification box
func WithPosition(position Position) Option {
	return func(m *Model) {
		m.SetPosition(position)
	}
}

// WithStyle sets the style for the notification box
func WithStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.Style = style
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		m.Hide()
		return m, nil
	case ShowMsg:
		cmd := m.Show(msg.Message)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	if !m.visible {
		return ""
	}
	return m.Style.Render(m.msg)
}

func (m *Model) Show(msg string) tea.Cmd {
	m.msg = msg
	m.visible = true
	return tea.Tick(m.duration,
		func(t time.Time) tea.Msg {
			return TickMsg{}
		})
}

func (m *Model) Hide() {
	m.visible = false
}

func (m *Model) SetDuration(duration time.Duration) {
	m.duration = duration
}

func (m Model) Duration() time.Duration {
	return m.duration
}

func (m *Model) SetPosition(position Position) {
	m.position = position
}

func (m Model) Position() Position {
	return m.position
}

func (m Model) Visible() bool {
	return m.visible
}
