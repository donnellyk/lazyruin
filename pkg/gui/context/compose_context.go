package context

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// ComposeState holds state specific to the compose preview mode.
type ComposeState struct {
	PreviewNavState
	PreviewDisplayState
	Note            models.Note             // single composed note
	SelectedCardIdx int                     // always 0
	SourceMap       []models.SourceMapEntry // maps composed line ranges to source children
	ParentUUID      string                  // for reload after mutations
	ParentTitle     string                  // for reload after mutations
}

// ComposeContext owns the compose preview mode (parent composition via `ruin compose`).
type ComposeContext struct {
	BaseContext
	*ComposeState
	navHistory *SharedNavHistory
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
		ComposeState: &ComposeState{
			PreviewNavState:     PreviewNavState{HighlightedLink: -1},
			PreviewDisplayState: PreviewDisplayState{RenderMarkdown: true},
		},
		navHistory: navHistory,
	}
}

// IPreviewContext implementation.

func (self *ComposeContext) NavState() *PreviewNavState         { return &self.PreviewNavState }
func (self *ComposeContext) DisplayState() *PreviewDisplayState { return &self.PreviewDisplayState }
func (self *ComposeContext) SelectedCardIndex() int             { return self.SelectedCardIdx }
func (self *ComposeContext) SetSelectedCardIndex(idx int)       { self.SelectedCardIdx = idx }
func (self *ComposeContext) CardCount() int                     { return 1 }
func (self *ComposeContext) NavHistory() *SharedNavHistory      { return self.navHistory }

var _ types.Context = &ComposeContext{}
var _ IPreviewContext = &ComposeContext{}
