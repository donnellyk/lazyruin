package gui

import "github.com/jesseduffield/gocui"

// View name constants.
const (
	NotesView        = "notes"
	QueriesView      = "queries"
	TagsView         = "tags"
	PreviewView      = "preview"
	SearchView       = "search"
	SearchFilterView = "searchFilter"
	StatusView       = "status"
)

// Views holds references to all views.
type Views struct {
	Notes        *gocui.View
	Queries      *gocui.View
	Tags         *gocui.View
	Preview      *gocui.View
	Search       *gocui.View
	SearchFilter *gocui.View
	Status       *gocui.View
}
