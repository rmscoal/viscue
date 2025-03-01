package style

import (
	"viscue/tui/tool/cache"

	"github.com/charmbracelet/lipgloss"
)

var (
	ErrorText = lipgloss.NewStyle().MarginTop(2).Foreground(ColorRed).Render

	LogoContainer = lipgloss.NewStyle().Align(lipgloss.Center,
		lipgloss.Center).
		MarginBottom(2)
	Logo = lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(ColorPurple).
		SetString(` ▌ ▐·▪  .▄▄ ·  ▄▄· ▄• ▄▌▄▄▄ .
▪█·█▌██ ▐█ ▀. ▐█ ▌▪█▪██▌▀▄.▀·
▐█▐█•▐█·▄▀▀▀█▄██ ▄▄█▌▐█▌▐▀▀▪▄
 ███ ▐█▌▐█▄▪▐█▐███▌▐█▄█▌▐█▄▄▌
. ▀  ▀▀▀ ▀▀▀▀ ·▀▀▀  ▀▀▀  ▀▀▀
`).Render()
	SubLogo = lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(ColorPurplePale).
		SetString(`Your personal terminal password manager.`).Render()
	HeaderHeight = lipgloss.Height(Logo) + lipgloss.Height(SubLogo) + 2

	HelpContainer = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Height(HelpViewHeight).
			MaxHeight(HelpViewHeight).
			Render

	ModelTitleStyle = lipgloss.NewStyle().Background(ColorGray).
			Foreground(ColorNormal).
			MarginBottom(1).
			Padding(0, 1)
	ModelTitleFocusedStyle = ModelTitleStyle.Background(ColorPurple)
)

const (
	HelpViewHeight = 4 // Maximum height of the help view
)

func CalculateAppHeight() int {
	terminalHeight := cache.Get[int](cache.TerminalHeight)
	return terminalHeight - HeaderHeight - HelpViewHeight
}
