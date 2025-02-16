package prompt

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	Cycle key.Binding
	Close key.Binding

	// For passwords prompt
	TogglePasswordVisibility key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Cycle}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Cycle},
	}
}

var HelpKeys = keymap{
	Cycle: key.NewBinding(
		key.WithKeys("tab", "shift+tab"),
		key.WithHelp("tab", "cycle fields"),
	),
	Close: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel & close"),
	),
	TogglePasswordVisibility: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "toggle password visibility"),
	),
}
