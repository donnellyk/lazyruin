package context

import "kvnd/lazyruin/pkg/gui/types"

// SearchContext owns the search popup panel.
// During the hybrid migration period, state remains in GuiState.
type SearchContext struct {
	BaseContext
}

// NewSearchContext creates a SearchContext.
func NewSearchContext() *SearchContext {
	return &SearchContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.PERSISTENT_POPUP,
			Key:       "search",
			ViewName:  "search",
			Focusable: true,
			Title:     "Search",
		}),
	}
}

var _ types.Context = &SearchContext{}
