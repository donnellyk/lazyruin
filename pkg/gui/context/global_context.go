package context

import "kvnd/lazyruin/pkg/gui/types"

// GlobalContext owns global keybindings that fire regardless of focused view.
// Its view name is "" (the gocui global view), so bindings register everywhere.
type GlobalContext struct {
	BaseContext
}

// NewGlobalContext creates a GlobalContext.
func NewGlobalContext() *GlobalContext {
	return &GlobalContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.GLOBAL_CONTEXT,
			Key:       "global",
			ViewNames: []string{""},
			Focusable: false,
			Title:     "Global",
		}),
	}
}

var _ types.Context = &GlobalContext{}
