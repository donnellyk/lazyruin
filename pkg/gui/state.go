package gui

type GuiState struct {
	Dialog      *DialogState
	Initialized bool
	lastWidth   int
	lastHeight  int
	// StartupWarning is a persistent warning shown in the status bar from
	// app startup (e.g., the ruin CLI version is below the minimum). Empty
	// when no warning. Cleared on the first dismissible keypress via
	// DismissStartupWarning and not shown again until the next launch.
	StartupWarning string
	// NeedsInit is set by app startup when the configured vault path has
	// not yet been initialized as a ruin vault. The first-run layout hook
	// shows an init prompt before any onboarding flow.
	NeedsInit bool
	// ExitError is set when the TUI exits via a flow that should surface
	// an error to the terminal (e.g., the user picks "Quit" on the init
	// prompt). Read by app.Run after Gui.Run returns.
	ExitError error
	// NotesOuterTab tracks which outer tab is active in the Notes pane
	// when sections_mode is enabled. Values: "home" or "notes". Empty
	// string defaults to "home" on first render. Used by RenderNotes to
	// pick which content to draw, independent of which context is
	// currently focused (the user can navigate away to Preview without
	// switching the outer tab).
	NotesOuterTab string
}

func NewGuiState() *GuiState {
	return &GuiState{}
}
