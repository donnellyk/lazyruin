package gui

import (
	"testing"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

func TestNewGuiState_Defaults(t *testing.T) {
	state := NewGuiState()

	if state.SearchQuery != "" {
		t.Errorf("SearchQuery = %q, want empty", state.SearchQuery)
	}
}

func TestNewContextMgr_Defaults(t *testing.T) {
	mgr := NewContextMgr()

	if mgr.Current() != "notes" {
		t.Errorf("Current() = %v, want notes", mgr.Current())
	}
	if mgr.Previous() != "notes" {
		t.Errorf("Previous() = %v, want notes", mgr.Previous())
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

func TestContextMgr_ContextTracking(t *testing.T) {
	mgr := NewContextMgr()

	// Simulate context switch via stack push
	mgr.Push("tags")

	if mgr.Previous() != "notes" {
		t.Errorf("Previous() = %v, want notes", mgr.Previous())
	}
	if mgr.Current() != "tags" {
		t.Errorf("Current() = %v, want tags", mgr.Current())
	}

	// Switch again
	mgr.Push("preview")

	if mgr.Previous() != "tags" {
		t.Errorf("Previous() = %v, want tags", mgr.Previous())
	}
	if mgr.Current() != "preview" {
		t.Errorf("Current() = %v, want preview", mgr.Current())
	}
}

func TestContextMgr_Pop(t *testing.T) {
	mgr := NewContextMgr()
	mgr.Push("tags")
	mgr.Push("preview")

	mgr.Pop()
	if mgr.Current() != "tags" {
		t.Errorf("after Pop: Current() = %v, want tags", mgr.Current())
	}

	mgr.Pop()
	if mgr.Current() != "notes" {
		t.Errorf("after second Pop: Current() = %v, want notes", mgr.Current())
	}

	// Pop on single-element stack is a no-op
	mgr.Pop()
	if mgr.Current() != "notes" {
		t.Errorf("after Pop on single: Current() = %v, want notes", mgr.Current())
	}
}

func TestContextMgr_Replace(t *testing.T) {
	mgr := NewContextMgr()
	mgr.Push("tags")

	mgr.Replace("queries")
	if mgr.Current() != "queries" {
		t.Errorf("after Replace: Current() = %v, want queries", mgr.Current())
	}
	if mgr.Previous() != "notes" {
		t.Errorf("after Replace: Previous() = %v, want notes", mgr.Previous())
	}
}

func TestContextMgr_SetStack(t *testing.T) {
	mgr := NewContextMgr()
	mgr.SetStack([]types.ContextKey{"notes", "capture"})

	if mgr.Current() != "capture" {
		t.Errorf("Current() = %v, want capture", mgr.Current())
	}
	if mgr.Previous() != "notes" {
		t.Errorf("Previous() = %v, want notes", mgr.Previous())
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
