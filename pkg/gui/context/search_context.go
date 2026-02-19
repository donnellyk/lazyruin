package context

import "kvnd/lazyruin/pkg/gui/types"

// SearchContext owns the search popup panel and its associated state.
type SearchContext struct {
	BaseContext
	Query      string
	Completion *types.CompletionState
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
		Completion: types.NewCompletionState(),
	}
}

var _ types.Context = &SearchContext{}
