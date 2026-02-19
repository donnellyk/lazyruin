package context

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// CalendarState holds the runtime state of the calendar dialog.
type CalendarState struct {
	Year        int
	Month       int // 1-12
	SelectedDay int // 1-31
	Focus       int // 0 = grid, 1 = notes, 2 = input
	Notes       []models.Note
	NoteIndex   int
}

// CalendarContext owns the calendar dialog popup and its state.
// The popup has three views: grid (navigation), input (date entry), notes (note list).
type CalendarContext struct {
	BaseContext
	State *CalendarState
}

// NewCalendarContext creates a CalendarContext.
func NewCalendarContext() *CalendarContext {
	return &CalendarContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:            types.TEMPORARY_POPUP,
			Key:             "calendarGrid",
			ViewNames:       []string{"calendarGrid", "calendarInput", "calendarNotes"},
			PrimaryViewName: "calendarGrid",
			Focusable:       true,
			Title:           "Calendar",
		}),
	}
}

var _ types.Context = &CalendarContext{}
