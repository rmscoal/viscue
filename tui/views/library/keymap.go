package library

import "github.com/charmbracelet/bubbles/key"

type listKeyMap struct {
	// Navigations
	Up         key.Binding
	Down       key.Binding
	NextPage   key.Binding
	PrevPage   key.Binding
	FocusRight key.Binding
	Help       key.Binding

	// Utilities
	Add    key.Binding
	Rename key.Binding
	Delete key.Binding

	// List things
	Filter       key.Binding
	CancelFilter key.Binding
	ClearFilter  key.Binding
}

func (k listKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.FocusRight, k.Help}
}

func (k listKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.NextPage, k.PrevPage, k.FocusRight}, // first column
		{k.Add, k.Rename, k.Delete},                          // second column
		{k.Filter, k.CancelFilter, k.ClearFilter},            // third column
	}
}

var listKeys = listKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next page"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev page"),
	),
	FocusRight: key.NewBinding(
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
	Rename: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit current"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete current"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter"),
	),
	CancelFilter: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit filter"),
	),
	ClearFilter: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear filter"),
	),
}

type tableKeyMap struct {
	// Navigations
	Up        key.Binding
	Down      key.Binding
	NextPage  key.Binding
	PrevPage  key.Binding
	FocusLeft key.Binding
	Help      key.Binding

	// Utilities
	Add    key.Binding
	Edit   key.Binding
	Delete key.Binding

	// List things
	Filter       key.Binding
	CancelFilter key.Binding
	ClearFilter  key.Binding
}

func (k tableKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.FocusLeft, k.Help}
}

func (k tableKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.NextPage, k.PrevPage, k.FocusLeft}, // first column
		{k.Add, k.Edit, k.Delete},                           // second column
		{k.Filter, k.CancelFilter, k.ClearFilter},           // third column
	}
}

var tableKeys = tableKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next page"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev page"),
	),
	FocusLeft: key.NewBinding(
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
		key.WithHelp("e", "edit current"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete current"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter"),
	),
	CancelFilter: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit filter"),
	),
	ClearFilter: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear filter"),
	),
}
