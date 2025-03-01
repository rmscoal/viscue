package prompt

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Cycle  key.Binding
	Close  key.Binding
	Submit key.Binding

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
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	TogglePasswordVisibility: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "toggle password visibility"),
	),
}

type DropdownActiveKeyMap struct {
	Up, Down, Select, Cancel key.Binding
}

func (k DropdownActiveKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Cancel}
}

func (k DropdownActiveKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Down, k.Select},
	}
}

var DropdownKeys = DropdownActiveKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}
