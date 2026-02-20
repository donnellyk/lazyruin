package context

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
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
	Title             string
	PickResults       []models.PickResult
	ContextKey        types.ContextKey        // which context key to restore ("cardList", "pickResults", "compose")
	SourceMap         []models.SourceMapEntry // compose source map
	ParentUUID        string                  // compose parent UUID
	ParentTitle       string                  // compose parent title
}
