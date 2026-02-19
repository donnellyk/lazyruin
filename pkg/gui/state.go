package gui

import (
	"kvnd/lazyruin/pkg/gui/types"
)

type GuiState struct {
	Dialog           *DialogState
	SearchQuery      string
	SearchCompletion *types.CompletionState
	Initialized      bool
	lastWidth        int
	lastHeight       int
}

func NewGuiState() *GuiState {
	return &GuiState{
		SearchCompletion: types.NewCompletionState(),
	}
}
