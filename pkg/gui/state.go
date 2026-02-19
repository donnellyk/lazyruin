package gui

type GuiState struct {
	Dialog      *DialogState
	Initialized bool
	lastWidth   int
	lastHeight  int
}

func NewGuiState() *GuiState {
	return &GuiState{}
}
