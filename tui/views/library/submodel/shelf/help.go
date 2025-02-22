package shelf

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up, Down, Switch, Help,
	Add, Edit, Delete, Copy,
	Search, ClearSearch key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Switch, k.Copy, k.Search, k.Help}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Switch, k.Help},  // first column
		{k.Add, k.Edit, k.Delete, k.Copy}, // second column
		{k.Search, k.ClearSearch},         // third column
	}
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Switch: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "focus left"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e", "enter"),
		key.WithHelp("e/enter", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Copy: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "copy password"),
	),
	Search: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "search"),
	),
	ClearSearch: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear search"),
	),
}
