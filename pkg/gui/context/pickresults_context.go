package context

import (
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
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
	Results    []models.PickResult
	Query      string // dialog mode: for title display
	ScopeTitle string // dialog mode: scoped context name

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
	PreviewContextTrait
	*PickResultsState
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
		PreviewContextTrait: NewPreviewContextTrait(navHistory),
		PickResultsState:    &PickResultsState{},
	}
}

// IPreviewContext implementation (CardCount varies per context; the rest are
// provided by the embedded PreviewContextTrait).

func (self *PickResultsContext) CardCount() int { return len(self.Results) }

// Filterable implementation.

func (self *PickResultsContext) GetFilterText() string    { return self.FilterText }
func (self *PickResultsContext) SetFilterText(s string)   { self.FilterText = s }
func (self *PickResultsContext) ItemCount() int           { return len(self.Results) }
func (self *PickResultsContext) GetUnfilteredCount() int  { return self.UnfilteredCount }
func (self *PickResultsContext) SetUnfilteredCount(n int) { self.UnfilteredCount = n }
func (self *PickResultsContext) ResetSelectedCard()       { self.SelectedCardIdx = 0 }
func (self *PickResultsContext) HasRequery() bool         { return self.Source.Requery != nil }
func (self *PickResultsContext) FilterTriggers() func() []types.CompletionTrigger {
	return self.Source.Triggers
}

func (self *PickResultsContext) RequeryAndApply(filterText string) error {
	results, err := self.Source.Requery(filterText)
	if err != nil {
		return err
	}
	self.Results = results
	return nil
}

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
		PreviewContextTrait: NewPreviewContextTrait(pickDialogNavHistory),
		PickResultsState:    &PickResultsState{},
	}
}

var _ types.Context = &PickResultsContext{}
var _ IPreviewContext = &PickResultsContext{}
var _ Filterable = &PickResultsContext{}
