package context

import "kvnd/lazyruin/pkg/models"

// PreviewMode selects between card-list and pick-results display.
type PreviewMode int

const (
	PreviewModeCardList PreviewMode = iota
	PreviewModePickResults
)

// PreviewLink represents a detected link in the preview content.
type PreviewLink struct {
	Text string // display text (wiki-link target or URL)
	Line int    // absolute line number in the rendered preview
	Col  int    // start column (visible characters, 0-indexed)
	Len  int    // visible length of the link text
}

// NavEntry captures a snapshot of preview state for back/forward navigation.
type NavEntry struct {
	Cards             []models.Note
	SelectedCardIndex int
	CursorLine        int
	ScrollOffset      int
	Mode              PreviewMode
	Title             string
	PickResults       []models.PickResult
}

// PreviewState holds all mutable state for the preview panel.
type PreviewState struct {
	Mode              PreviewMode
	Cards             []models.Note
	SelectedCardIndex int
	ScrollOffset      int
	CursorLine        int // highlighted line in multi-card modes (-1 = no cursor)
	ShowFrontmatter   bool
	ShowTitle         bool
	ShowGlobalTags    bool
	CardLineRanges    [][2]int // [startLine, endLine) for each card
	HeaderLines       []int    // absolute line numbers containing markdown headers
	RenderMarkdown    bool     // true to render markdown with glamour
	PickResults       []models.PickResult
	Links             []PreviewLink // detected links in current render
	HighlightedLink   int           // index into Links; -1 = none, auto-cleared each render
	RenderedLink      int           // snapshot of HighlightedLink used during current render
	TemporarilyMoved  map[int]bool  // card indices temporarily moved
}
