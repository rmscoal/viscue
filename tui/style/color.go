package style

import "github.com/charmbracelet/lipgloss"

var (
	ColorWhite         = lipgloss.Color("15")
	ColorBlack         = lipgloss.Color("0")
	ColorTerminalGreen = lipgloss.AdaptiveColor{
		Light: "#33FF33", Dark: "#33FF33",
	}
	ColorPurple    = lipgloss.AdaptiveColor{Light: "#A020F0", Dark: "#A020F0"}
	ColorDarkgreen = lipgloss.AdaptiveColor{
		Light: "#0E4E0B", Dark: "#0E4E0B",
	}
	ColorRed = lipgloss.AdaptiveColor{
		Light: "#FF3333", Dark: "#FF3333",
	}
	ColorGray = lipgloss.AdaptiveColor{
		Light: "#D9DCCF", Dark: "#626262",
	}
	ColorMilk = lipgloss.AdaptiveColor{
		Light: "#FFF7DB", Dark: "#FFF7DB",
	}
)
