package context

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// PickResultsSource holds metadata about the query that populated pick results,
// enabling re-query for filtering.
type PickResultsSource struct {
	Query    string                                               // for display/seed in filter dialog
	Requery  func(filterText string) ([]models.PickResult, error) // combines filter with original query
	Triggers func() []types.CompletionTrigger                     // completion triggers for filter dialog
}

// PickResultsState holds state specific to the pick-results preview mode.
type PickResultsState struct {
	PreviewNavState
	PreviewDisplayState
	Results         []models.PickResult
	SelectedCardIdx int
	Query           string // dialog mode: for title display
	ScopeTitle      string // dialog mode: scoped context name

	FilterText      string
	Source          PickResultsSource
	UnfilteredCount int
}

func (s *PickResultsState) FilterActive() bool {
	return s.FilterText != ""
}

func (s *PickResultsState) ClearFilter() {
	s.FilterText = ""
	s.UnfilteredCount = 0
}

// PickResultsContext owns the pick-results preview mode (inline tag pick,
// pick dialog results).
type PickResultsContext struct {
	BaseContext
	*PickResultsState
	navHistory *SharedNavHistory
}

// NewPickResultsContext creates a PickResultsContext with a shared nav history.
func NewPickResultsContext(navHistory *SharedNavHistory) *PickResultsContext {
	return &PickResultsContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.MAIN_CONTEXT,
			Key:       "pickResults",
			ViewName:  "preview",
			Focusable: true,
			Title:     "Pick Results",
		}),
		PickResultsState: &PickResultsState{
			PreviewNavState:     PreviewNavState{HighlightedLink: -1},
			PreviewDisplayState: PreviewDisplayState{RenderMarkdown: true, DimDone: true},
		},
		navHistory: navHistory,
	}
}

// IPreviewContext implementation.

func (self *PickResultsContext) NavState() *PreviewNavState         { return &self.PreviewNavState }
func (self *PickResultsContext) DisplayState() *PreviewDisplayState { return &self.PreviewDisplayState }
func (self *PickResultsContext) SelectedCardIndex() int             { return self.SelectedCardIdx }
func (self *PickResultsContext) SetSelectedCardIndex(idx int)       { self.SelectedCardIdx = idx }
func (self *PickResultsContext) CardCount() int                     { return len(self.Results) }
func (self *PickResultsContext) NavHistory() *SharedNavHistory      { return self.navHistory }

var pickDialogNavHistory = &SharedNavHistory{Index: -1}

// NewPickDialogContext creates a PickResultsContext configured as a dialog overlay.
func NewPickDialogContext() *PickResultsContext {
	return &PickResultsContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.TEMPORARY_POPUP,
			Key:       "pickDialog",
			ViewName:  "pickDialog",
			Focusable: true,
			Title:     "Pick Dialog",
		}),
		PickResultsState: &PickResultsState{},
		navHistory:       pickDialogNavHistory,
	}
}

var _ types.Context = &PickResultsContext{}
var _ IPreviewContext = &PickResultsContext{}
