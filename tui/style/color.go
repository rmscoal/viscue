package style

import "github.com/charmbracelet/lipgloss"

var (
	ColorNormal = lipgloss.AdaptiveColor{
		Light: "#000000", // Black
		Dark:  "#FFFFFF", // White
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
		Light: "#969696", Dark: "#c2c2c2",
	}
)
