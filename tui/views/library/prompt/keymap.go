package prompt

import "github.com/charmbracelet/bubbles/key"

type baseKeyMap struct {
	Tab    key.Binding
	Enter  key.Binding
	Escape key.Binding
}

func (k baseKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Enter, k.Escape}
}

func (k baseKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab},
		{k.Enter},
		{k.Escape},
	}
}

var baseKeys = baseKeyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab", "shift+tab"),
		key.WithHelp("tab/shift+tab", "cycle focus"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

type passwordKeyMap struct {
	baseKeyMap
	ShowPassword key.Binding
}

func (k passwordKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.ShowPassword, k.Enter, k.Escape}
}

func (k passwordKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ShowPassword, k.Tab},
		{k.Enter, k.Escape},
	}
}

var passwordKeys = passwordKeyMap{
	baseKeyMap: baseKeys,
	ShowPassword: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "show/hide password"),
	),
}
