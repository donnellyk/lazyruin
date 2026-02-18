package context

import "kvnd/lazyruin/pkg/gui/types"

// CaptureContext owns the capture popup panel.
// During the hybrid migration period, state remains in GuiState.
type CaptureContext struct {
	BaseContext
}

// NewCaptureContext creates a CaptureContext.
func NewCaptureContext() *CaptureContext {
	return &CaptureContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.PERSISTENT_POPUP,
			Key:       "capture",
			ViewName:  "capture",
			Focusable: true,
			Title:     "Capture",
		}),
	}
}

var _ types.Context = &CaptureContext{}
