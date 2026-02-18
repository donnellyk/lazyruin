package context

import "kvnd/lazyruin/pkg/gui/types"

// ContextTree provides typed access to all context instances.
// During the hybrid migration, only migrated contexts are present here.
type ContextTree struct {
	Global        *GlobalContext
	Notes         *NotesContext
	Tags          *TagsContext
	Queries       *QueriesContext
	Preview       *PreviewContext
	Search        *SearchContext
	Capture       *CaptureContext
	Pick          *PickContext
	InputPopup    *InputPopupContext
	Palette       *PaletteContext
	SnippetEditor *SnippetEditorContext
	Calendar      *CalendarContext
	Contrib       *ContribContext
}

// All returns all contexts in the tree for iteration.
// During the hybrid migration, this only includes migrated contexts.
func (self *ContextTree) All() []types.Context {
	var all []types.Context
	if self.Global != nil {
		all = append(all, self.Global)
	}
	if self.Notes != nil {
		all = append(all, self.Notes)
	}
	if self.Tags != nil {
		all = append(all, self.Tags)
	}
	if self.Queries != nil {
		all = append(all, self.Queries)
	}
	if self.Preview != nil {
		all = append(all, self.Preview)
	}
	if self.Search != nil {
		all = append(all, self.Search)
	}
	if self.Capture != nil {
		all = append(all, self.Capture)
	}
	if self.Pick != nil {
		all = append(all, self.Pick)
	}
	if self.InputPopup != nil {
		all = append(all, self.InputPopup)
	}
	if self.Palette != nil {
		all = append(all, self.Palette)
	}
	if self.SnippetEditor != nil {
		all = append(all, self.SnippetEditor)
	}
	if self.Calendar != nil {
		all = append(all, self.Calendar)
	}
	if self.Contrib != nil {
		all = append(all, self.Contrib)
	}
	return all
}
