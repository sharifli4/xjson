package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Home       key.Binding
	End        key.Binding
	Next       key.Binding
	Prev       key.Binding
	Refresh    key.Binding
	Search     key.Binding
	Profile    key.Binding
	Timeline   key.Binding
	Expand     key.Binding
	Collapse   key.Binding
	Help       key.Binding
	Quit       key.Binding
	Enter      key.Binding
	Escape     key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/↓", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("PgUp", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("PgDn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G", "bottom"),
		),
		Next: key.NewBinding(
			key.WithKeys("n", "tab"),
			key.WithHelp("n", "next item"),
		),
		Prev: key.NewBinding(
			key.WithKeys("p", "shift+tab"),
			key.WithHelp("p", "prev item"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "ctrl+r"),
			key.WithHelp("r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Profile: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "user profile"),
		),
		Timeline: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "timeline"),
		),
		Expand: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l", "expand"),
		),
		Collapse: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h", "collapse"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↵", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// ShortHelp returns keybindings for the short help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Next, k.Prev, k.Search, k.Refresh, k.Quit}
}

// FullHelp returns keybindings for the full help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Next, k.Prev, k.Home, k.End},
		{k.Search, k.Profile, k.Timeline, k.Refresh},
		{k.Expand, k.Collapse, k.Help, k.Quit},
	}
}
