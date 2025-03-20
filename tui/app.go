package tui

import (
	"os"

	"viscue/tui/style"
	"viscue/tui/tool/cache"
	"viscue/tui/tool/database"
	"viscue/tui/tool/debugger"
	"viscue/tui/views/library"
	"viscue/tui/views/login"
	"viscue/tui/views/warning"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/term"
	"github.com/jmoiron/sqlx"
)

type app struct {
	db *sqlx.DB

	warningView tea.Model
	appView     tea.Model
}

func NewApp(db *sqlx.DB) tea.Model {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		log.Fatal("NewApp: failed getting terminal size", "err", err)
	}

	cache.Set(cache.TerminalWidth, width)
	cache.Set(cache.TerminalHeight, height)

	return &app{
		db:      db,
		appView: login.New(db),
	}
}

func (m *app) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Height < 24 || msg.Width < 90 {
			m.warningView = warning.NewScreenSize(msg.Width, msg.Height)
		} else {
			m.warningView = nil
			cache.Set(cache.TerminalHeight, msg.Height)
			cache.Set(cache.TerminalWidth, msg.Width)
			goto PassToCurrentView
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case login.Successful:
		m.appView = library.New(m.db)
		return m, m.appView.Init()
	}

PassToCurrentView:
	var cmd tea.Cmd
	m.appView, cmd = m.appView.Update(msg)
	return m, cmd
}

func (m *app) View() string {
	width := cache.Get[int](cache.TerminalWidth)
	canvas := lipgloss.NewStyle().Width(width).Render
	header := style.LogoContainer.Width(width).Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			style.Logo,
			style.SubLogo,
		),
	)

	var view string
	if m.warningView != nil {
		view = m.warningView.View()
	} else {
		view = m.appView.View()
	}

	return canvas(
		lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			view,
		),
	)
}

func Run() int {
	db, err := database.New()
	if err != nil {
		log.Error("failed connecting to database", "err", err)
		return 1
	}
	defer db.Close()

	file, err := debugger.New()
	if err != nil {
		log.Error("failed initializing debugger", "err", err)
		return 1
	}
	defer file.Close()

	_, err = tea.NewProgram(NewApp(db)).Run()
	if err != nil {
		log.Error("unable to start application", "err", err)
		return 1
	}

	return 0
}
