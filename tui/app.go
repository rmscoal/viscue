package tui

import (
	"os"
	"time"

	"viscue/tui/event"
	"viscue/tui/style"
	"viscue/tui/tool/cache"
	"viscue/tui/tool/database"
	"viscue/tui/views/library"
	"viscue/tui/views/login"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/term"
	"github.com/jmoiron/sqlx"
)

type app struct {
	db *sqlx.DB

	currentView tea.Model
}

func NewApp(db *sqlx.DB) tea.Model {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		log.Fatal("NewApp: failed getting terminal size", "err", err)
	}

	cache.Set(cache.TerminalWidth, width)
	cache.Set(cache.TerminalHeight, height)

	return &app{
		db:          db,
		currentView: login.New(db),
	}
}

func (m *app) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cache.Set(cache.TerminalHeight, msg.Height)
		cache.Set(cache.TerminalWidth, msg.Width)
		goto PassToCurrentView
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case event.AppMessage:
		switch msg {
		case event.UserLoggedIn:
			libraryView := library.New(m.db)
			m.currentView = libraryView
			return m, m.currentView.Init()
		}
	}

PassToCurrentView:
	var cmd tea.Cmd
	m.currentView, cmd = m.currentView.Update(msg)
	return m, cmd
}

func (m *app) View() string {
	width := cache.Get[int](cache.TerminalWidth)
	canvas := lipgloss.NewStyle().Width(width).Render
	header := style.TitleContainer.Width(width).Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			style.Title,
			style.SubTitle,
		),
	)

	return canvas(
		lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			m.currentView.View(),
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

	filename := "error.log"
	log.SetLevel(log.ErrorLevel)

	_, ok := os.LookupEnv("DEBUG")
	if ok {
		log.SetLevel(log.DebugLevel)
		filename = "debug.log"
	}

	file, err := tea.LogToFileWith(filename, "", log.Default())
	if err != nil {
		log.Error("failed to create log file", "err", err)
		return 1
	}
	defer file.Close()
	log.Info("viscue application started",
		"time", time.Now().Format(time.RFC3339))

	_, err = tea.NewProgram(NewApp(db)).Run()
	if err != nil {
		log.Error("unable to start application", "err", err)
		return 1
	}

	return 0
}
