package types

// PaletteCommand represents a single command in the command palette.
type PaletteCommand struct {
	Name     string
	Category string
	Key      string
	OnRun    func() error
	Contexts []ContextKey // nil = always available
}

// PaletteState holds the runtime state of the command palette.
type PaletteState struct {
	Commands      []PaletteCommand
	Filtered      []PaletteCommand
	SelectedIndex int
	FilterText    string
}
