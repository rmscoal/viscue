package prompt

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Cycle key.Binding
	Close key.Binding

	// For passwords prompt
	TogglePasswordVisibility key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Cycle}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Cycle},
	}
}

var Keys = KeyMap{
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
