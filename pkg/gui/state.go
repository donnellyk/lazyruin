package gui

import (
	"kvnd/lazyruin/pkg/gui/types"
)

type GuiState struct {
	Dialog           *DialogState
	ContextStack     []types.ContextKey
	SearchQuery      string
	SearchCompletion *types.CompletionState
	Initialized      bool
	lastWidth        int
	lastHeight       int
}

func NewGuiState() *GuiState {
	return &GuiState{
		SearchCompletion: types.NewCompletionState(),
		ContextStack:     []types.ContextKey{"notes"},
	}
}

// currentContext returns the top of the context stack.
func (s *GuiState) currentContext() types.ContextKey {
	if len(s.ContextStack) == 0 {
		return "notes"
	}
	return s.ContextStack[len(s.ContextStack)-1]
}

// previousContext returns the second-from-top of the context stack.
func (s *GuiState) previousContext() types.ContextKey {
	if len(s.ContextStack) < 2 {
		return "notes"
	}
	return s.ContextStack[len(s.ContextStack)-2]
}
