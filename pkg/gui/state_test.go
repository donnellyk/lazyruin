package gui

import (
	"testing"

	"kvnd/lazyruin/pkg/models"
)

func TestNewGuiState_Defaults(t *testing.T) {
	state := NewGuiState()

	if state.CurrentContext != NotesContext {
		t.Errorf("CurrentContext = %v, want %v", state.CurrentContext, NotesContext)
	}

	if state.PreviousContext != "" {
		t.Errorf("PreviousContext = %v, want empty", state.PreviousContext)
	}

	if state.SearchMode != false {
		t.Error("SearchMode should default to false")
	}

	if state.SearchQuery != "" {
		t.Errorf("SearchQuery = %q, want empty", state.SearchQuery)
	}
}

func TestNewGuiState_SubStatesInitialized(t *testing.T) {
	state := NewGuiState()

	if state.Notes == nil {
		t.Error("Notes state should not be nil")
	}
	if state.Queries == nil {
		t.Error("Queries state should not be nil")
	}
	if state.Tags == nil {
		t.Error("Tags state should not be nil")
	}
	if state.Preview == nil {
		t.Error("Preview state should not be nil")
	}
}

func TestNewGuiState_NotesStateDefaults(t *testing.T) {
	state := NewGuiState()

	if state.Notes.SelectedIndex != 0 {
		t.Errorf("Notes.SelectedIndex = %d, want 0", state.Notes.SelectedIndex)
	}
	if len(state.Notes.Items) != 0 {
		t.Errorf("Notes.Items length = %d, want 0", len(state.Notes.Items))
	}
}

func TestNewGuiState_PreviewStateDefaults(t *testing.T) {
	state := NewGuiState()

	if state.Preview.Mode != PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want PreviewModeCardList", state.Preview.Mode)
	}
	if state.Preview.SelectedCardIndex != 0 {
		t.Errorf("Preview.SelectedCardIndex = %d, want 0", state.Preview.SelectedCardIndex)
	}
	if state.Preview.ScrollOffset != 0 {
		t.Errorf("Preview.ScrollOffset = %d, want 0", state.Preview.ScrollOffset)
	}
	if state.Preview.ShowFrontmatter != false {
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

	// Simulate context switch (as setContext would do)
	state.PreviousContext = state.CurrentContext
	state.CurrentContext = TagsContext

	if state.PreviousContext != NotesContext {
		t.Errorf("PreviousContext = %v, want NotesContext", state.PreviousContext)
	}
	if state.CurrentContext != TagsContext {
		t.Errorf("CurrentContext = %v, want TagsContext", state.CurrentContext)
	}

	// Switch again
	state.PreviousContext = state.CurrentContext
	state.CurrentContext = PreviewContext

	if state.PreviousContext != TagsContext {
		t.Errorf("PreviousContext = %v, want TagsContext", state.PreviousContext)
	}
	if state.CurrentContext != PreviewContext {
		t.Errorf("CurrentContext = %v, want PreviewContext", state.CurrentContext)
	}
}

func TestNotesState_Selection(t *testing.T) {
	state := NewGuiState()

	// Add some notes
	state.Notes.Items = []models.Note{
		{UUID: "1", Title: "Note 1"},
		{UUID: "2", Title: "Note 2"},
		{UUID: "3", Title: "Note 3"},
	}

	// Test selection bounds
	state.Notes.SelectedIndex = 0
	if state.Notes.SelectedIndex < 0 {
		t.Error("SelectedIndex should not be negative")
	}

	state.Notes.SelectedIndex = 2
	if state.Notes.SelectedIndex >= len(state.Notes.Items) {
		t.Error("SelectedIndex should be within bounds")
	}

	// Get selected note
	selected := state.Notes.Items[state.Notes.SelectedIndex]
	if selected.UUID != "3" {
		t.Errorf("Selected note UUID = %q, want %q", selected.UUID, "3")
	}
}

func TestPreviewState_ModeSwitch(t *testing.T) {
	state := NewGuiState()

	// Default is card list mode
	if state.Preview.Mode != PreviewModeCardList {
		t.Errorf("Initial mode = %v, want PreviewModeCardList", state.Preview.Mode)
	}

	// Set up card list
	state.Preview.Cards = []models.Note{
		{UUID: "1", Title: "Card 1"},
		{UUID: "2", Title: "Card 2"},
	}
	state.Preview.SelectedCardIndex = 0

	if len(state.Preview.Cards) != 2 {
		t.Errorf("Cards length = %d, want 2", len(state.Preview.Cards))
	}

	// Switch to pick results mode
	state.Preview.Mode = PreviewModePickResults
	state.Preview.ScrollOffset = 5

	if state.Preview.Mode != PreviewModePickResults {
		t.Errorf("Mode = %v, want PreviewModePickResults", state.Preview.Mode)
	}
	if state.Preview.ScrollOffset != 5 {
		t.Errorf("ScrollOffset = %d, want 5", state.Preview.ScrollOffset)
	}
}

func TestPreviewState_FrontmatterToggle(t *testing.T) {
	state := NewGuiState()

	if state.Preview.ShowFrontmatter {
		t.Error("ShowFrontmatter should default to false")
	}

	state.Preview.ShowFrontmatter = true
	if !state.Preview.ShowFrontmatter {
		t.Error("ShowFrontmatter should be true after toggle")
	}

	state.Preview.ShowFrontmatter = false
	if state.Preview.ShowFrontmatter {
		t.Error("ShowFrontmatter should be false after toggle")
	}
}

func TestSearchMode_Toggle(t *testing.T) {
	state := NewGuiState()

	if state.SearchMode {
		t.Error("SearchMode should default to false")
	}

	// Enter search mode
	state.SearchMode = true
	state.PreviousContext = state.CurrentContext
	state.CurrentContext = SearchContext

	if !state.SearchMode {
		t.Error("SearchMode should be true")
	}
	if state.CurrentContext != SearchContext {
		t.Errorf("CurrentContext = %v, want SearchContext", state.CurrentContext)
	}

	// Exit search mode
	state.SearchMode = false
	state.CurrentContext = state.PreviousContext

	if state.SearchMode {
		t.Error("SearchMode should be false after exit")
	}
	if state.CurrentContext != NotesContext {
		t.Errorf("CurrentContext = %v, want NotesContext", state.CurrentContext)
	}
}
