package prompt

import "github.com/charmbracelet/bubbles/key"

type baseKeyMap struct {
	Tab    key.Binding
	Submit key.Binding
	Cancel key.Binding
}

func (k baseKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Submit, k.Cancel}
}

func (k baseKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab},
		{k.Submit},
		{k.Cancel},
	}
}

var baseKeys = baseKeyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab", "shift+tab"),
		key.WithHelp("tab/shift+tab", "cycle focus"),
	),
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

type passwordKeyMap struct {
	baseKeyMap
	ShowPassword key.Binding
}

func (k passwordKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.ShowPassword, k.Submit, k.Cancel}
}

func (k passwordKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ShowPassword, k.Tab},
		{k.Submit, k.Cancel},
	}
}

var passwordKeys = passwordKeyMap{
	baseKeyMap: baseKeys,
	ShowPassword: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "show/hide password"),
	),
}
