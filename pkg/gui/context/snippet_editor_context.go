package context

import "kvnd/lazyruin/pkg/gui/types"

// SnippetEditorContext owns the snippet editor popup panel.
// The popup has two stacked views: name (top) and expansion (bottom).
type SnippetEditorContext struct {
	BaseContext
}

// NewSnippetEditorContext creates a SnippetEditorContext.
func NewSnippetEditorContext() *SnippetEditorContext {
	return &SnippetEditorContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:            types.TEMPORARY_POPUP,
			Key:             "snippetName",
			ViewNames:       []string{"snippetName", "snippetExpansion"},
			PrimaryViewName: "snippetName",
			Focusable:       true,
			Title:           "New Snippet",
		}),
	}
}

var _ types.Context = &SnippetEditorContext{}
