package types

// MenuItem represents a single item in a menu dialog.
type MenuItem struct {
	Label       string
	Hint        string // dim right-aligned hint for header items
	Key         string // shortcut key hint (e.g. "j", "k"), closes dialog on press
	KeepOpenKey string // hidden shortcut, runs OnRun on selected item without closing
	OnRun       func() error
	IsHeader    bool
}
