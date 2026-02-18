package context

import "kvnd/lazyruin/pkg/gui/types"

// ContextTree provides typed access to all context instances.
// During the hybrid migration, only migrated contexts are present here.
type ContextTree struct {
	Tags *TagsContext
}

// All returns all contexts in the tree for iteration.
// During the hybrid migration, this only includes migrated contexts.
func (self *ContextTree) All() []types.Context {
	var all []types.Context
	if self.Tags != nil {
		all = append(all, self.Tags)
	}
	return all
}
