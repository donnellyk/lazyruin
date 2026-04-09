package context

import "github.com/donnellyk/lazyruin/pkg/gui/types"

// PreviewNavState holds navigation state shared across all preview contexts.
type PreviewNavState struct {
	ScrollOffset    int
	CursorLine      int                // highlighted line in multi-card modes (-1 = no cursor)
	CardLineRanges  [][2]int           // [startLine, endLine) for each card
	HeaderLines     []int              // absolute line numbers containing markdown headers
	Lines           []types.SourceLine // indexed by visual line number; populated during rendering
	Links           []PreviewLink
	HighlightedLink int // index into Links; -1 = none, auto-cleared each render
	RenderedLink    int // snapshot of HighlightedLink used during current render
}

// PreviewDisplayState holds display toggle state shared across all preview contexts.
type PreviewDisplayState struct {
	ShowFrontmatter bool
	ShowTitle       bool
	ShowGlobalTags  bool
	RenderMarkdown  bool
	DimDone         bool
}

// SharedNavHistory holds the navigation history stack shared across all three preview contexts.
type SharedNavHistory struct {
	Entries []NavEntry
	Index   int // -1 = no history
}

// NewSharedNavHistory creates a SharedNavHistory.
func NewSharedNavHistory() *SharedNavHistory {
	return &SharedNavHistory{Index: -1}
}

// PreviewState bundles navigation and display state shared by all preview contexts.
type PreviewState struct {
	PreviewNavState
	PreviewDisplayState
	SelectedCardIdx int
}

// PreviewContextTrait provides the IPreviewContext delegation methods that are
// identical across CardList, PickResults, Compose, and DatePreview.  Each
// concrete preview context embeds this trait instead of reimplementing the
// five common accessors.
type PreviewContextTrait struct {
	PreviewState
	navHistory *SharedNavHistory
}

// NewPreviewContextTrait creates a PreviewContextTrait with sensible defaults.
func NewPreviewContextTrait(navHistory *SharedNavHistory) PreviewContextTrait {
	return PreviewContextTrait{
		PreviewState: PreviewState{
			PreviewNavState:     PreviewNavState{HighlightedLink: -1},
			PreviewDisplayState: PreviewDisplayState{RenderMarkdown: true, DimDone: true},
		},
		navHistory: navHistory,
	}
}

func (t *PreviewContextTrait) NavState() *PreviewNavState         { return &t.PreviewNavState }
func (t *PreviewContextTrait) DisplayState() *PreviewDisplayState { return &t.PreviewDisplayState }
func (t *PreviewContextTrait) SelectedCardIndex() int             { return t.SelectedCardIdx }
func (t *PreviewContextTrait) SetSelectedCardIndex(idx int)       { t.SelectedCardIdx = idx }
func (t *PreviewContextTrait) NavHistory() *SharedNavHistory      { return t.navHistory }

// IPreviewContext is the interface that all preview contexts implement,
// allowing helpers to work generically across CardList, PickResults, Compose, and DatePreview.
type IPreviewContext interface {
	types.Context

	NavState() *PreviewNavState
	DisplayState() *PreviewDisplayState
	SelectedCardIndex() int
	SetSelectedCardIndex(int)
	CardCount() int
	NavHistory() *SharedNavHistory
	SetTitle(string)
}

// Filterable abstracts the filter state shared by CardListContext and
// PickResultsContext, allowing a single set of open/apply/clear helpers
// to work with both context types.
type Filterable interface {
	Title() string
	GetFilterText() string
	SetFilterText(string)
	FilterActive() bool
	ClearFilter()
	ItemCount() int
	GetUnfilteredCount() int
	SetUnfilteredCount(int)
	ResetSelectedCard()
	HasRequery() bool
	RequeryAndApply(filterText string) error
	FilterTriggers() func() []types.CompletionTrigger
}
