package gui

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

type GuiState struct {
	Dialog                  *DialogState
	ContextStack            []types.ContextKey
	SearchQuery             string
	SearchCompletion        *types.CompletionState
	PaletteSeedDone         bool
	Palette                 *PaletteState
	Calendar                *CalendarState
	Contrib                 *ContribState
	Initialized             bool
	lastWidth               int
	lastHeight              int
}

// PaletteCommand represents a single command in the command palette.
type PaletteCommand struct {
	Name     string
	Category string
	Key      string
	OnRun    func() error
	Contexts []types.ContextKey // nil = always available
}

// PaletteState holds the runtime state of the command palette.
type PaletteState struct {
	Commands      []PaletteCommand
	Filtered      []PaletteCommand
	SelectedIndex int
	FilterText    string
}

// CalendarState holds the runtime state of the calendar dialog.
type CalendarState struct {
	Year        int
	Month       int // 1-12
	SelectedDay int // 1-31
	Focus       int // 0 = grid, 1 = notes, 2 = input
	Notes       []models.Note
	NoteIndex   int
}

// ContribState holds the runtime state of the contribution chart dialog.
type ContribState struct {
	DayCounts    map[string]int // "YYYY-MM-DD" -> count
	SelectedDate string         // "YYYY-MM-DD"
	Focus        int            // 0 = grid, 1 = note list
	Notes        []models.Note
	NoteIndex    int
	WeekCount    int // number of weeks displayed
}

func NewGuiState() *GuiState {
	return &GuiState{
		SearchCompletion:        types.NewCompletionState(),
		ContextStack:            []types.ContextKey{"notes"},
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
