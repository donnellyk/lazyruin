package context

import "kvnd/lazyruin/pkg/gui/types"

// ContextTree provides typed access to all context instances.
type ContextTree struct {
	Global           *GlobalContext
	Notes            *NotesContext
	Tags             *TagsContext
	Queries          *QueriesContext
	CardList         *CardListContext
	PickResults      *PickResultsContext
	Compose          *ComposeContext
	Search           *SearchContext
	Capture          *CaptureContext
	Pick             *PickContext
	InputPopup       *InputPopupContext
	Palette          *PaletteContext
	SnippetEditor    *SnippetEditorContext
	Calendar         *CalendarContext
	Contrib          *ContribContext
	PickDialog       *PickResultsContext
	ActivePreviewKey types.ContextKey // "cardList", "pickResults", or "compose"
}

// ViewNameForKey returns the primary view name for a context key,
// derived from the context's own metadata. For lightweight contexts not
// in the tree (e.g. "searchFilter"), the key itself is the view name.
func (self *ContextTree) ViewNameForKey(key types.ContextKey) string {
	for _, ctx := range self.All() {
		if ctx.GetKey() == key {
			return ctx.GetPrimaryViewName()
		}
	}
	return string(key)
}

// All returns all contexts in the tree for iteration.
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
	if self.CardList != nil {
		all = append(all, self.CardList)
	}
	if self.PickResults != nil {
		all = append(all, self.PickResults)
	}
	if self.Compose != nil {
		all = append(all, self.Compose)
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
	if self.PickDialog != nil {
		all = append(all, self.PickDialog)
	}
	return all
}

// ActivePreview returns the IPreviewContext for the current ActivePreviewKey.
// Defaults to CardList if ActivePreviewKey is unset.
func (self *ContextTree) ActivePreview() IPreviewContext {
	switch self.ActivePreviewKey {
	case "pickResults":
		return self.PickResults
	case "compose":
		return self.Compose
	default:
		return self.CardList
	}
}

// IsPreviewContextKey returns true if the key belongs to one of the three
// preview contexts.
func IsPreviewContextKey(key types.ContextKey) bool {
	switch key {
	case "cardList", "pickResults", "compose":
		return true
	default:
		return false
	}
}
