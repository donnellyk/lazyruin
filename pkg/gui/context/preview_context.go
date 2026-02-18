package context

import "kvnd/lazyruin/pkg/gui/types"

// PreviewContext owns the preview panel in the context/controller architecture.
// Unlike Notes/Tags/Queries, preview uses content-line navigation rather than
// item-list navigation, so it does NOT embed ListContextTrait.
// During the hybrid migration period, state remains in GuiState.Preview.
type PreviewContext struct {
	BaseContext
}

// NewPreviewContext creates a PreviewContext.
func NewPreviewContext() *PreviewContext {
	return &PreviewContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.MAIN_CONTEXT,
			Key:       "preview",
			ViewName:  "preview",
			Focusable: true,
			Title:     "Preview",
		}),
	}
}

// Verify interface compliance at compile time.
var _ types.Context = &PreviewContext{}
