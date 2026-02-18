package context

import "kvnd/lazyruin/pkg/gui/types"

// CalendarContext owns the calendar dialog popup.
// The popup has three views: grid (navigation), input (date entry), notes (note list).
type CalendarContext struct {
	BaseContext
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
