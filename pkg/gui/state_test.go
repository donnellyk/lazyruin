package gui

import (
	"testing"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"
)

func TestNewGuiState_Defaults(t *testing.T) {
	state := NewGuiState()

	if state.currentContext() != NotesContext {
		t.Errorf("currentContext() = %v, want %v", state.currentContext(), NotesContext)
	}

	if state.popupActive() {
		t.Error("popupActive() should be false by default")
	}

	if state.SearchQuery != "" {
		t.Errorf("SearchQuery = %q, want empty", state.SearchQuery)
	}
}

func TestNewPreviewContext_Initialized(t *testing.T) {
	pc := context.NewPreviewContext()

	if pc.PreviewState == nil {
		t.Error("PreviewState should not be nil")
	}
	if pc.NavIndex != -1 {
		t.Errorf("NavIndex = %d, want -1", pc.NavIndex)
	}
}

func TestPreviewContext_StateDefaults(t *testing.T) {
	pc := context.NewPreviewContext()

	if pc.Mode != PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want PreviewModeCardList", pc.Mode)
	}
	if pc.SelectedCardIndex != 0 {
		t.Errorf("Preview.SelectedCardIndex = %d, want 0", pc.SelectedCardIndex)
	}
	if pc.ScrollOffset != 0 {
		t.Errorf("Preview.ScrollOffset = %d, want 0", pc.ScrollOffset)
	}
	if pc.ShowFrontmatter != false {
		t.Error("Preview.ShowFrontmatter should default to false")
	}
}

func TestContextKey_Values(t *testing.T) {
	tests := []struct {
		ctx      ContextKey
		expected string
	}{
		{NotesContext, "notes"},
		{QueriesContext, "queries"},
		{TagsContext, "tags"},
		{PreviewContext, "preview"},
		{SearchContext, "search"},
	}

	for _, tc := range tests {
		if string(tc.ctx) != tc.expected {
			t.Errorf("ContextKey %v = %q, want %q", tc.ctx, string(tc.ctx), tc.expected)
		}
	}
}

func TestPreviewMode_Values(t *testing.T) {
	if PreviewModeCardList != 0 {
		t.Errorf("PreviewModeCardList = %d, want 0", PreviewModeCardList)
	}
	if PreviewModePickResults != 1 {
		t.Errorf("PreviewModePickResults = %d, want 1", PreviewModePickResults)
	}
}

func TestGuiState_ContextTracking(t *testing.T) {
	state := NewGuiState()

	// Simulate context switch via stack push
	state.ContextStack = append(state.ContextStack, TagsContext)

	if state.previousContext() != NotesContext {
		t.Errorf("previousContext() = %v, want NotesContext", state.previousContext())
	}
	if state.currentContext() != TagsContext {
		t.Errorf("currentContext() = %v, want TagsContext", state.currentContext())
	}

	// Switch again
	state.ContextStack = append(state.ContextStack, PreviewContext)

	if state.previousContext() != TagsContext {
		t.Errorf("previousContext() = %v, want TagsContext", state.previousContext())
	}
	if state.currentContext() != PreviewContext {
		t.Errorf("currentContext() = %v, want PreviewContext", state.currentContext())
	}
}

func TestPreviewState_ModeSwitch(t *testing.T) {
	pc := context.NewPreviewContext()

	// Default is card list mode
	if pc.Mode != PreviewModeCardList {
		t.Errorf("Initial mode = %v, want PreviewModeCardList", pc.Mode)
	}

	// Set up card list
	pc.Cards = []models.Note{
		{UUID: "1", Title: "Card 1"},
		{UUID: "2", Title: "Card 2"},
	}
	pc.SelectedCardIndex = 0

	if len(pc.Cards) != 2 {
		t.Errorf("Cards length = %d, want 2", len(pc.Cards))
	}

	// Switch to pick results mode
	pc.Mode = PreviewModePickResults
	pc.ScrollOffset = 5

	if pc.Mode != PreviewModePickResults {
		t.Errorf("Mode = %v, want PreviewModePickResults", pc.Mode)
	}
	if pc.ScrollOffset != 5 {
		t.Errorf("ScrollOffset = %d, want 5", pc.ScrollOffset)
	}
}

func TestPreviewState_FrontmatterToggle(t *testing.T) {
	pc := context.NewPreviewContext()

	if pc.ShowFrontmatter {
		t.Error("ShowFrontmatter should default to false")
	}

	pc.ShowFrontmatter = true
	if !pc.ShowFrontmatter {
		t.Error("ShowFrontmatter should be true after toggle")
	}

	pc.ShowFrontmatter = false
	if pc.ShowFrontmatter {
		t.Error("ShowFrontmatter should be false after toggle")
	}
}

func TestSearchPopupActive_Toggle(t *testing.T) {
	state := NewGuiState()

	if state.popupActive() {
		t.Error("popupActive() should be false by default")
	}

	// Enter search via context stack
	state.ContextStack = append(state.ContextStack, SearchContext)

	if !state.popupActive() {
		t.Error("popupActive() should be true when SearchContext is active")
	}
	if state.currentContext() != SearchContext {
		t.Errorf("currentContext() = %v, want SearchContext", state.currentContext())
	}

	// Exit search
	state.ContextStack = state.ContextStack[:len(state.ContextStack)-1]

	if state.popupActive() {
		t.Error("popupActive() should be false after exiting search")
	}
	if state.currentContext() != NotesContext {
		t.Errorf("currentContext() = %v, want NotesContext", state.currentContext())
	}
}
