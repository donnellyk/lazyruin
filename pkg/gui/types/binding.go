package types

import "github.com/jesseduffield/gocui"

// Binding represents a single keybinding with metadata for palette, help, and conflict detection.
type Binding struct {
	ID                string             // stable identity for auditing (e.g. "tags.rename")
	Key               any                // gocui key (rune or gocui.Key)
	Mod               gocui.Modifier     // default: gocui.ModNone
	Handler           func() error
	Description       string             // shown in palette & help; empty = nav-only
	Tooltip           string
	Category          string             // palette grouping
	GetDisabledReason func() *DisabledReason
	DisplayOnScreen   bool               // show in status bar hints
}

// DisabledReason explains why a binding is currently disabled.
type DisabledReason struct {
	Text string
}

// KeybindingsOpts provides configuration for keybinding generation.
type KeybindingsOpts struct {
	GetKey func(string) any // config lookup (future: user-configurable keys)
}

// KeybindingsFn is a producer function that returns keybindings for a context.
type KeybindingsFn func(opts KeybindingsOpts) []*Binding

// MouseKeybindingsFn is a producer function that returns mouse bindings for a context.
type MouseKeybindingsFn func(opts KeybindingsOpts) []*gocui.ViewMouseBinding
