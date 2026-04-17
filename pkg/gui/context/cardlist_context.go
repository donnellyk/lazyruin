package context

import (
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
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

	ComposedNote      *models.Note
	ComposedSourceMap []models.SourceMapEntry
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

// NewCardListContext creates a CardListContext.
func NewCardListContext() *CardListContext {
	return &CardListContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.MAIN_CONTEXT,
			Key:       "cardList",
			ViewName:  "preview",
			Focusable: true,
			Title:     "Preview",
		}),
		PreviewContextTrait: NewPreviewContextTrait(),
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

// cardListSnapshot is the CardListContext-specific snapshot. Carries enough
// view params to re-run the query (Source with a Requery closure) and enough
// view state to reconstruct the exact visual position on restore.
type cardListSnapshot struct {
	Title             string
	Source            CardListSource
	FilterText        string
	FrozenCards       []models.Note
	SelectedCardIdx   int
	CursorLine        int
	ScrollOffset      int
	Display           PreviewDisplayState
	ComposedNote      *models.Note
	ComposedSourceMap []models.SourceMapEntry
}

// CaptureSnapshot captures the CardList's current state.
func (self *CardListContext) CaptureSnapshot() types.Snapshot {
	ns := self.NavState()
	return &cardListSnapshot{
		Title:             self.Title(),
		Source:            self.Source,
		FilterText:        self.FilterText,
		FrozenCards:       append([]models.Note(nil), self.Cards...),
		SelectedCardIdx:   self.SelectedCardIdx,
		CursorLine:        ns.CursorLine,
		ScrollOffset:      ns.ScrollOffset,
		Display:           *self.DisplayState(),
		ComposedNote:      self.ComposedNote,
		ComposedSourceMap: self.ComposedSourceMap,
	}
}

// RestoreSnapshot restores CardList state from a previously captured snapshot,
// re-running the Source query (when available) to pick up any data changes.
func (self *CardListContext) RestoreSnapshot(s types.Snapshot) error {
	snap, ok := s.(*cardListSnapshot)
	if !ok || snap == nil {
		return nil
	}
	self.SetTitle(snap.Title)
	self.Source = snap.Source
	self.FilterText = snap.FilterText
	*self.DisplayState() = snap.Display
	self.ComposedNote = snap.ComposedNote
	self.ComposedSourceMap = snap.ComposedSourceMap

	if snap.Source.Requery != nil {
		notes, err := snap.Source.Requery(snap.FilterText)
		if err == nil {
			self.Cards = notes
		} else {
			self.Cards = append([]models.Note(nil), snap.FrozenCards...)
		}
	} else {
		self.Cards = append([]models.Note(nil), snap.FrozenCards...)
	}

	if snap.SelectedCardIdx < len(self.Cards) {
		self.SelectedCardIdx = snap.SelectedCardIdx
	} else if len(self.Cards) > 0 {
		self.SelectedCardIdx = len(self.Cards) - 1
	} else {
		self.SelectedCardIdx = 0
	}
	ns := self.NavState()
	ns.CursorLine = snap.CursorLine
	ns.ScrollOffset = snap.ScrollOffset
	return nil
}

var _ types.Context = &CardListContext{}
var _ IPreviewContext = &CardListContext{}
var _ Filterable = &CardListContext{}
var _ types.Snapshotter = &CardListContext{}
