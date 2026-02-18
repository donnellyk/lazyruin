package context

import "kvnd/lazyruin/pkg/gui/types"

// PickContext owns the pick popup panel.
// During the hybrid migration period, state remains in GuiState.
type PickContext struct {
	BaseContext
}

// NewPickContext creates a PickContext.
func NewPickContext() *PickContext {
	return &PickContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.TEMPORARY_POPUP,
			Key:       "pick",
			ViewName:  "pick",
			Focusable: true,
			Title:     "Pick",
		}),
	}
}

var _ types.Context = &PickContext{}
