package context

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// ContribState holds the runtime state of the contribution chart dialog.
type ContribState struct {
	DayCounts    map[string]int // "YYYY-MM-DD" -> count
	SelectedDate string         // "YYYY-MM-DD"
	Focus        int            // 0 = grid, 1 = note list
	Notes        []models.Note
	NoteIndex    int
	WeekCount    int // number of weeks displayed
}

// ContribContext owns the contribution chart dialog popup.
// The popup has two views: grid (heatmap) and notes (note list).
type ContribContext struct {
	BaseContext
	State *ContribState
}

// NewContribContext creates a ContribContext.
func NewContribContext() *ContribContext {
	return &ContribContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:            types.TEMPORARY_POPUP,
			Key:             "contribGrid",
			ViewNames:       []string{"contribGrid", "contribNotes"},
			PrimaryViewName: "contribGrid",
			Focusable:       true,
			Title:           "Contributions",
		}),
	}
}

var _ types.Context = &ContribContext{}
