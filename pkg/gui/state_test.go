package gui

import (
	"testing"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

func TestNewGuiState_Defaults(t *testing.T) {
	_ = NewGuiState()
}

func TestNewSearchContext_Defaults(t *testing.T) {
	sc := context.NewSearchContext()

	if sc.Query != "" {
		t.Errorf("Query = %q, want empty", sc.Query)
	}
	if sc.Completion == nil {
		t.Error("Completion should not be nil")
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

func TestContextKey_Values(t *testing.T) {
	tests := []struct {
		ctx      types.ContextKey
		expected string
	}{
		{"notes", "notes"},
		{"queries", "queries"},
		{"tags", "tags"},
		{"cardList", "cardList"},
		{"search", "search"},
	}

	for _, tc := range tests {
		if string(tc.ctx) != tc.expected {
			t.Errorf("types.ContextKey %v = %q, want %q", tc.ctx, string(tc.ctx), tc.expected)
		}
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
	mgr.Push("cardList")

	if mgr.Previous() != "tags" {
		t.Errorf("Previous() = %v, want tags", mgr.Previous())
	}
	if mgr.Current() != "cardList" {
		t.Errorf("Current() = %v, want cardList", mgr.Current())
	}
}

func TestContextMgr_Pop(t *testing.T) {
	mgr := NewContextMgr()
	mgr.Push("tags")
	mgr.Push("cardList")

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

func TestCardListContext_Defaults(t *testing.T) {
	navHistory := context.NewSharedNavHistory()
	cl := context.NewCardListContext(navHistory)

	if cl.SelectedCardIdx != 0 {
		t.Errorf("SelectedCardIdx = %d, want 0", cl.SelectedCardIdx)
	}

	ns := cl.NavState()
	if ns.ScrollOffset != 0 {
		t.Errorf("ScrollOffset = %d, want 0", ns.ScrollOffset)
	}

	ds := cl.DisplayState()
	if ds.ShowFrontmatter {
		t.Error("ShowFrontmatter should default to false")
	}
	if !ds.RenderMarkdown {
		t.Error("RenderMarkdown should default to true")
	}
}

func TestDisplayState_FrontmatterToggle(t *testing.T) {
	navHistory := context.NewSharedNavHistory()
	cl := context.NewCardListContext(navHistory)
	ds := cl.DisplayState()

	if ds.ShowFrontmatter {
		t.Error("ShowFrontmatter should default to false")
	}

	ds.ShowFrontmatter = true
	if !ds.ShowFrontmatter {
		t.Error("ShowFrontmatter should be true after toggle")
	}

	ds.ShowFrontmatter = false
	if ds.ShowFrontmatter {
		t.Error("ShowFrontmatter should be false after toggle")
	}
}

// popupActive is tested through handlers_test.go where a full *Gui is available.
