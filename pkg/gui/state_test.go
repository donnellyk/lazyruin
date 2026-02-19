package gui

import (
	"testing"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

func TestNewGuiState_Defaults(t *testing.T) {
	state := NewGuiState()

	if state.currentContext() != "notes" {
		t.Errorf("currentContext() = %v, want %v", state.currentContext(), "notes")
	}

	if state.SearchQuery != "" {
		t.Errorf("SearchQuery = %q, want empty", state.SearchQuery)
	}
}

func TestNewPreviewContext_Initialized(t *testing.T) {
	pc := context.NewPreviewContext()

	if pc.PreviewState == nil {
		t.Error("context.PreviewState should not be nil")
	}
	if pc.NavIndex != -1 {
		t.Errorf("NavIndex = %d, want -1", pc.NavIndex)
	}
}

func TestPreviewContext_StateDefaults(t *testing.T) {
	pc := context.NewPreviewContext()

	if pc.Mode != context.PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want context.PreviewModeCardList", pc.Mode)
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
		ctx      types.ContextKey
		expected string
	}{
		{"notes", "notes"},
		{"queries", "queries"},
		{"tags", "tags"},
		{"preview", "preview"},
		{"search", "search"},
	}

	for _, tc := range tests {
		if string(tc.ctx) != tc.expected {
			t.Errorf("types.ContextKey %v = %q, want %q", tc.ctx, string(tc.ctx), tc.expected)
		}
	}
}

func TestPreviewMode_Values(t *testing.T) {
	if context.PreviewModeCardList != 0 {
		t.Errorf("context.PreviewModeCardList = %d, want 0", context.PreviewModeCardList)
	}
	if context.PreviewModePickResults != 1 {
		t.Errorf("context.PreviewModePickResults = %d, want 1", context.PreviewModePickResults)
	}
}

func TestGuiState_ContextTracking(t *testing.T) {
	state := NewGuiState()

	// Simulate context switch via stack push
	state.ContextStack = append(state.ContextStack, "tags")

	if state.previousContext() != "notes" {
		t.Errorf("previousContext() = %v, want notes", state.previousContext())
	}
	if state.currentContext() != "tags" {
		t.Errorf("currentContext() = %v, want tags", state.currentContext())
	}

	// Switch again
	state.ContextStack = append(state.ContextStack, "preview")

	if state.previousContext() != "tags" {
		t.Errorf("previousContext() = %v, want tags", state.previousContext())
	}
	if state.currentContext() != "preview" {
		t.Errorf("currentContext() = %v, want preview", state.currentContext())
	}
}

func TestPreviewState_ModeSwitch(t *testing.T) {
	pc := context.NewPreviewContext()

	// Default is card list mode
	if pc.Mode != context.PreviewModeCardList {
		t.Errorf("Initial mode = %v, want context.PreviewModeCardList", pc.Mode)
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
	pc.Mode = context.PreviewModePickResults
	pc.ScrollOffset = 5

	if pc.Mode != context.PreviewModePickResults {
		t.Errorf("Mode = %v, want context.PreviewModePickResults", pc.Mode)
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

// popupActive is tested through handlers_test.go where a full *Gui is available.
