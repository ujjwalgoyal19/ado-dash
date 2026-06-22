package ui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeyMap holds keybindings active on every screen.
type GlobalKeyMap struct {
	Quit key.Binding
	Help key.Binding
	Zen  key.Binding
}

// GlobalKeys is the default global keymap.
var GlobalKeys = GlobalKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Zen: key.NewBinding(
		key.WithKeys("Z"),
		key.WithHelp("Z", "zen mode"),
	),
}
