package context

import "kvnd/lazyruin/pkg/gui/types"

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
