package context

import "kvnd/lazyruin/pkg/gui/types"

// SnippetEditorContext owns the snippet editor popup panel and its state.
// The popup has two stacked views: name (top) and expansion (bottom).
type SnippetEditorContext struct {
	BaseContext
	Focus      int // 0 = name, 1 = expansion
	Completion *types.CompletionState
}

// NewSnippetEditorContext creates a SnippetEditorContext.
func NewSnippetEditorContext() *SnippetEditorContext {
	return &SnippetEditorContext{
		Completion: types.NewCompletionState(),
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
