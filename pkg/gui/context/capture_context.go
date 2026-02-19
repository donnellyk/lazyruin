package context

import "kvnd/lazyruin/pkg/gui/types"

// CaptureParentInfo tracks the parent selected via > completion in the capture dialog.
type CaptureParentInfo struct {
	UUID  string
	Title string // display title for footer (e.g. "Parent / Child")
}

// CaptureContext owns the capture popup panel.
type CaptureContext struct {
	BaseContext
	Parent     *CaptureParentInfo
	Completion *types.CompletionState
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
		Completion: types.NewCompletionState(),
	}
}

var _ types.Context = &CaptureContext{}
