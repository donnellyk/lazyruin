package context

import "kvnd/lazyruin/pkg/gui/types"

// SearchContext owns the search popup panel and its associated state.
type SearchContext struct {
	BaseContext
	Query      string
	Completion *types.CompletionState

	// Filter mode: when OnFilterSubmit is set, the search dialog acts as a
	// filter input instead of a normal search.
	FilterTitle    string
	FilterSeed     string
	FilterSeedDone bool
	FilterTriggers func() []types.CompletionTrigger
	OnFilterSubmit func(string) error
}

func (s *SearchContext) InFilterMode() bool {
	return s.OnFilterSubmit != nil
}

func (s *SearchContext) ClearFilterMode() {
	s.FilterTitle = ""
	s.FilterSeed = ""
	s.FilterSeedDone = false
	s.FilterTriggers = nil
	s.OnFilterSubmit = nil
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
