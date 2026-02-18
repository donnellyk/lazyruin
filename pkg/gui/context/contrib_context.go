package context

import "kvnd/lazyruin/pkg/gui/types"

// ContribContext owns the contribution chart dialog popup.
// The popup has two views: grid (heatmap) and notes (note list).
type ContribContext struct {
	BaseContext
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
