package style

import "github.com/charmbracelet/lipgloss"

var (
	ColorLight         = lipgloss.Color("15")
	ColorBlack         = lipgloss.Color("0")
	ColorTerminalGreen = lipgloss.AdaptiveColor{
		Light: "#33FF33", Dark: "#33FF33",
	}
	ColorPurple     = lipgloss.AdaptiveColor{Light: "#A020F0", Dark: "#A020F0"}
	ColorPurplePale = lipgloss.AdaptiveColor{
		Light: "#F0A0F0", Dark: "#F0A0F0",
	}
	ColorRed = lipgloss.AdaptiveColor{
		Light: "#FF3333", Dark: "#FF3333",
	}
	ColorRedPale = lipgloss.AdaptiveColor{
		Light: "#FFA0A0", Dark: "#FFA0A0",
	}
	ColorGray = lipgloss.AdaptiveColor{
		Light: "#D9DCCF", Dark: "#626262",
	}
	ColorMilk = lipgloss.AdaptiveColor{
		Light: "#FFF7DB", Dark: "#FFF7DB",
	}
)
