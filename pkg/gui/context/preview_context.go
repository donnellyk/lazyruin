package context

import "kvnd/lazyruin/pkg/gui/types"

// PreviewContext owns the preview panel in the context/controller architecture.
// Unlike Notes/Tags/Queries, preview uses content-line navigation rather than
// item-list navigation, so it does NOT embed ListContextTrait.
type PreviewContext struct {
	BaseContext
	*PreviewState
	NavHistory []NavEntry
	NavIndex   int // -1 = no history
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
		PreviewState: &PreviewState{RenderMarkdown: true, HighlightedLink: -1},
		NavIndex:     -1,
	}
}

// Verify interface compliance at compile time.
var _ types.Context = &PreviewContext{}
