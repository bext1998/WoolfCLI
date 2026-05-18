package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	Top        key.Binding
	Bottom     key.Binding
	Submit     key.Binding
	Newline    key.Binding
	Tab        key.Binding
	Quit       key.Binding
	Cancel     key.Binding
	Help       key.Binding
}

var keys = keyMap{
	Up:         key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "scroll up")),
	Down:       key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "scroll down")),
	ScrollUp:   key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "scroll up")),
	ScrollDown: key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "scroll down")),
	Top:        key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "jump to top")),
	Bottom:     key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "jump to bottom")),
	Submit:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	Newline:    key.NewBinding(key.WithKeys("shift+enter"), key.WithHelp("shift+enter", "newline")),
	Tab:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch focus")),
	Quit:       key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	Cancel:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}
