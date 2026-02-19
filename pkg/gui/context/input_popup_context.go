package context

import "kvnd/lazyruin/pkg/gui/types"

// InputPopupContext owns the generic input popup panel and its state.
type InputPopupContext struct {
	BaseContext
	Completion *types.CompletionState
	SeedDone   bool
	Config     *types.InputPopupConfig
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
		Completion: types.NewCompletionState(),
	}
}

var _ types.Context = &InputPopupContext{}
