package gui

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

const (
	NotesContext        types.ContextKey = "notes"
	QueriesContext      types.ContextKey = "queries"
	TagsContext         types.ContextKey = "tags"
	PreviewContext      types.ContextKey = "preview"
	SearchContext       types.ContextKey = "search"
	SearchFilterContext types.ContextKey = "searchFilter"
	CaptureContext      types.ContextKey = "capture"
	PickContext         types.ContextKey = "pick"
	PaletteContext      types.ContextKey = "palette"
	InputPopupCtx       types.ContextKey = "inputPopup"
	SnippetEditorCtx    types.ContextKey = "snippetName"
	CalendarCtx         types.ContextKey = "calendarGrid"
	ContribCtx          types.ContextKey = "contribGrid"
)

// mainPanelContexts is the set of non-popup panel contexts.
// A context NOT in this set is treated as a popup by popupActive().
var mainPanelContexts = map[types.ContextKey]bool{
	NotesContext:        true,
	QueriesContext:      true,
	TagsContext:         true,
	PreviewContext:      true,
	SearchFilterContext: true,
}

// CaptureParentInfo tracks the parent selected via > completion in the capture dialog.
type CaptureParentInfo struct {
	UUID  string
	Title string // display title for footer (e.g. "Parent / Child")
}

type GuiState struct {
	Dialog                  *DialogState
	ContextStack            []types.ContextKey
	SearchQuery             string
	CaptureParent           *CaptureParentInfo
	SearchCompletion        *types.CompletionState
	CaptureCompletion       *types.CompletionState
	PickCompletion          *types.CompletionState
	PickQuery               string
	PickAnyMode             bool
	PickSeedHash            bool
	PaletteSeedDone         bool
	Palette                 *PaletteState
	InputPopupCompletion    *types.CompletionState
	InputPopupSeedDone      bool
	InputPopupConfig        *types.InputPopupConfig
	SnippetEditorFocus      int // 0 = name, 1 = expansion
	SnippetEditorCompletion *types.CompletionState
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
		CaptureCompletion:       types.NewCompletionState(),
		PickCompletion:          types.NewCompletionState(),
		InputPopupCompletion:    types.NewCompletionState(),
		SnippetEditorCompletion: types.NewCompletionState(),
		ContextStack:            []types.ContextKey{NotesContext},
	}
}

// popupActive returns true when the current context is a popup (not a main panel).
func (s *GuiState) popupActive() bool {
	return !mainPanelContexts[s.currentContext()]
}

// currentContext returns the top of the context stack.
func (s *GuiState) currentContext() types.ContextKey {
	if len(s.ContextStack) == 0 {
		return NotesContext
	}
	return s.ContextStack[len(s.ContextStack)-1]
}

// previousContext returns the second-from-top of the context stack.
func (s *GuiState) previousContext() types.ContextKey {
	if len(s.ContextStack) < 2 {
		return NotesContext
	}
	return s.ContextStack[len(s.ContextStack)-2]
}
