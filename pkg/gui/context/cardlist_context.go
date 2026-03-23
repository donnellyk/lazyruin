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
	Cards            []models.Note
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
	PreviewContextTrait
	*CardListState
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
		PreviewContextTrait: NewPreviewContextTrait(navHistory),
		CardListState:       &CardListState{},
	}
}

// IPreviewContext implementation (CardCount varies per context; the rest are
// provided by the embedded PreviewContextTrait).

func (self *CardListContext) CardCount() int { return len(self.Cards) }

// Filterable implementation.

func (self *CardListContext) GetFilterText() string    { return self.FilterText }
func (self *CardListContext) SetFilterText(s string)   { self.FilterText = s }
func (self *CardListContext) ItemCount() int           { return len(self.Cards) }
func (self *CardListContext) GetUnfilteredCount() int  { return self.UnfilteredCount }
func (self *CardListContext) SetUnfilteredCount(n int) { self.UnfilteredCount = n }
func (self *CardListContext) ResetSelectedCard()       { self.SelectedCardIdx = 0 }
func (self *CardListContext) HasRequery() bool         { return self.Source.Requery != nil }
func (self *CardListContext) FilterTriggers() func() []types.CompletionTrigger {
	return self.Source.Triggers
}

func (self *CardListContext) RequeryAndApply(filterText string) error {
	notes, err := self.Source.Requery(filterText)
	if err != nil {
		return err
	}
	self.Cards = notes
	return nil
}

var _ types.Context = &CardListContext{}
var _ IPreviewContext = &CardListContext{}
var _ Filterable = &CardListContext{}
