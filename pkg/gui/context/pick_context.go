package context

import "kvnd/lazyruin/pkg/gui/types"

// PickContext owns the pick popup panel and its state.
type PickContext struct {
	BaseContext
	Query       string
	AnyMode     bool
	TodoMode    bool
	AllTagsMode bool
	SeedHash    bool
	DialogMode  bool   // when true, ExecutePick shows results in a dialog overlay
	ScopeTitle  string // contextual title set when DialogMode is true
	Completion  *types.CompletionState
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
		Completion: types.NewCompletionState(),
	}
}

var _ types.Context = &PickContext{}
