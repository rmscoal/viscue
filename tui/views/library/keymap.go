package library

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up, Down, Next, Prev, Switch, Help,
	Add, Edit, Delete,
	Search, Clear, Escape, Enter key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Switch, k.Search, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Next, k.Prev, k.Switch}, // first column
		{k.Add, k.Edit, k.Delete},                // second column
		{k.Search, k.Clear, k.Escape},            // third column
	}
}

var focusRight = key.NewBinding(
	key.WithKeys("ctrl+l"),
	key.WithHelp("ctrl+l", "focus right"),
)

var focusLeft = key.NewBinding(
	key.WithKeys("ctrl+h"),
	key.WithHelp("ctrl+h", "focus left"),
)

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Next: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next page"),
	),
	Prev: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev page"),
	),
	Switch: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "focus right"),
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
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Search: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "search"),
	),
	Clear: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear search"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
}
