package gui

import "github.com/jesseduffield/gocui"

// View name constants.
const (
	NotesView              = "notes"
	QueriesView            = "queries"
	TagsView               = "tags"
	PreviewView            = "preview"
	SearchView             = "search"
	SearchFilterView       = "searchFilter"
	SearchSuggestView      = "searchSuggest"
	CaptureView            = "capture"
	CaptureSuggestView     = "captureSuggest"
	PickView               = "pick"
	PickSuggestView        = "pickSuggest"
	StatusView             = "status"
	MenuView               = "menu"
	PaletteView            = "palette"
	PaletteListView        = "paletteList"
	ParentInputView        = "parentInput"
	ParentInputSuggestView = "parentInputSuggest"
	TagInputView           = "tagInput"
	TagInputSuggestView    = "tagInputSuggest"
)

// Views holds references to all views.
type Views struct {
	Notes        *gocui.View
	Queries      *gocui.View
	Tags         *gocui.View
	Preview      *gocui.View
	Search       *gocui.View
	SearchFilter *gocui.View
	Capture      *gocui.View
	Pick         *gocui.View
	Status       *gocui.View
	Palette      *gocui.View
	PaletteList  *gocui.View
}
