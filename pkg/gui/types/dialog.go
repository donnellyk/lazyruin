package types

// MenuItem represents a single item in a menu dialog.
type MenuItem struct {
	Label    string
	Key      string // shortcut key hint (e.g. "j", "k")
	OnRun    func() error
	IsHeader bool
}
