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
}

func NewGuiState() *GuiState {
	return &GuiState{}
}
