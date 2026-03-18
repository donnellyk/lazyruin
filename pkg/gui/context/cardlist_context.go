package context

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// CardListSource holds metadata about the query that populated a card list,
// enabling re-query for filtering.
type CardListSource struct {
	Query    string                                         // for display/seed in filter dialog
	Requery  func(filterText string) ([]models.Note, error) // combines filter with original query
	Triggers func() []types.CompletionTrigger               // completion triggers for filter dialog
}

// CardListState holds state specific to the card-list preview mode.
type CardListState struct {
	PreviewNavState
	PreviewDisplayState
	Cards            []models.Note
	SelectedCardIdx  int
	TemporarilyMoved map[int]bool

	FilterText      string
	Source          CardListSource
	UnfilteredCount int
}

func (s *CardListState) FilterActive() bool {
	return s.FilterText != ""
}

func (s *CardListState) ClearFilter() {
	s.FilterText = ""
	s.UnfilteredCount = 0
}

// CardListContext owns the card-list preview mode (search results, tag/query
// results, calendar/contrib dates, single-note view).
type CardListContext struct {
	BaseContext
	*CardListState
	navHistory *SharedNavHistory
}

// NewCardListContext creates a CardListContext with a shared nav history.
func NewCardListContext(navHistory *SharedNavHistory) *CardListContext {
	return &CardListContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.MAIN_CONTEXT,
			Key:       "cardList",
			ViewName:  "preview",
			Focusable: true,
			Title:     "Preview",
		}),
		CardListState: &CardListState{
			PreviewNavState:     PreviewNavState{HighlightedLink: -1},
			PreviewDisplayState: PreviewDisplayState{RenderMarkdown: true, DimDone: true},
		},
		navHistory: navHistory,
	}
}

// IPreviewContext implementation.

func (self *CardListContext) NavState() *PreviewNavState         { return &self.PreviewNavState }
func (self *CardListContext) DisplayState() *PreviewDisplayState { return &self.PreviewDisplayState }
func (self *CardListContext) SelectedCardIndex() int             { return self.SelectedCardIdx }
func (self *CardListContext) SetSelectedCardIndex(idx int)       { self.SelectedCardIdx = idx }
func (self *CardListContext) CardCount() int                     { return len(self.Cards) }
func (self *CardListContext) NavHistory() *SharedNavHistory      { return self.navHistory }

var _ types.Context = &CardListContext{}
var _ IPreviewContext = &CardListContext{}
