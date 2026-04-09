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
}

func NewGuiState() *GuiState {
	return &GuiState{}
}
