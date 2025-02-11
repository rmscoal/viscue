package style

import (
	"viscue/tui/tool/cache"

	"github.com/charmbracelet/lipgloss"
)

var (
	ErrorText = lipgloss.NewStyle().MarginTop(2).Foreground(ColorRed).Render

	TitleContainer = lipgloss.NewStyle().Align(lipgloss.Center,
		lipgloss.Center).
		MarginBottom(2)
	Title = lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(ColorPurple).
		SetString(` ▌ ▐·▪  .▄▄ ·  ▄▄· ▄• ▄▌▄▄▄ .
▪█·█▌██ ▐█ ▀. ▐█ ▌▪█▪██▌▀▄.▀·
▐█▐█•▐█·▄▀▀▀█▄██ ▄▄█▌▐█▌▐▀▀▪▄
 ███ ▐█▌▐█▄▪▐█▐███▌▐█▄█▌▐█▄▄▌
. ▀  ▀▀▀ ▀▀▀▀ ·▀▀▀  ▀▀▀  ▀▀▀
`).Render()
	SubTitle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(ColorPurplePale).
			SetString(`Your personal terminal password manager.`).Render()
	HeaderHeight = lipgloss.Height(Title) + lipgloss.Height(SubTitle) + 2

	HelpContainer = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Height(HelpViewHeight).
			MaxHeight(HelpViewHeight).
			Render
)

const (
	HelpViewHeight = 4 // Maximum height of the help view
)

func CalculateAppHeight() int {
	terminalHeight := cache.Get[int](cache.TerminalHeight)
	return terminalHeight - HeaderHeight - HelpViewHeight
}
