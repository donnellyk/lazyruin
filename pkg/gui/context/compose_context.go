package context

import (
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
)

// ComposeState holds state specific to the compose preview mode.
type ComposeState struct {
	Note      models.Note             // single composed note
	SourceMap []models.SourceMapEntry // maps composed line ranges to source children
	Parent    models.ParentBookmark   // for reload after mutations
}

// ComposeContext owns the compose preview mode (parent composition via `ruin compose`).
type ComposeContext struct {
	BaseContext
	PreviewContextTrait
	*ComposeState
}

// NewComposeContext creates a ComposeContext with a shared nav history.
func NewComposeContext(navHistory *SharedNavHistory) *ComposeContext {
	return &ComposeContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.MAIN_CONTEXT,
			Key:       "compose",
			ViewName:  "preview",
			Focusable: true,
			Title:     "Compose",
		}),
		PreviewContextTrait: NewPreviewContextTrait(navHistory),
		ComposeState:        &ComposeState{},
	}
}

// IPreviewContext implementation (CardCount varies per context; the rest are
// provided by the embedded PreviewContextTrait).

func (self *ComposeContext) CardCount() int { return 1 }

var _ types.Context = &ComposeContext{}
var _ IPreviewContext = &ComposeContext{}
