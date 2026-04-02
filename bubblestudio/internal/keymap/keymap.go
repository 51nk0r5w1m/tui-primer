// Package keymap defines keyboard bindings for the bubblestudio shell.
package keymap

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds the key bindings used by the app model.
type KeyMap struct {
	Quit key.Binding
	Help key.Binding
	Tab  key.Binding // switch to form demo view
	Back key.Binding // return to list view
}

// ShortHelp implements help.KeyMap so help.Model can render a compact hint.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Help}
}

// FullHelp implements help.KeyMap for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.Help},
		{k.Tab, k.Back},
	}
}

// Default returns a KeyMap with sensible defaults.
func Default() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "form demo"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}
