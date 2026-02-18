package context

import "kvnd/lazyruin/pkg/gui/types"

// InputPopupContext owns the generic input popup panel.
// During the hybrid migration period, state remains in GuiState.
type InputPopupContext struct {
	BaseContext
}

// NewInputPopupContext creates an InputPopupContext.
func NewInputPopupContext() *InputPopupContext {
	return &InputPopupContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.TEMPORARY_POPUP,
			Key:       "inputPopup",
			ViewName:  "inputPopup",
			Focusable: true,
			Title:     "Input",
		}),
	}
}

var _ types.Context = &InputPopupContext{}
