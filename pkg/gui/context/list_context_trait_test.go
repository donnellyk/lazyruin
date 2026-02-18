package context

import "testing"

func newTrait(n, idx int) (*ListContextTrait, *int, *int) {
	list := &simpleList{n}
	cursor := NewListCursor(list)
	cursor.selectedLineIdx = idx
	renderCount := 0
	previewCount := 0
	trait := NewListContextTrait(cursor, func() { renderCount++ }, func() { previewCount++ })
	return trait, &renderCount, &previewCount
}

func TestListContextTrait_DelegatesToCursor(t *testing.T) {
	trait, _, _ := newTrait(5, 2)

	if trait.GetSelectedLineIdx() != 2 {
		t.Errorf("GetSelectedLineIdx = %d, want 2", trait.GetSelectedLineIdx())
	}

	trait.SetSelectedLineIdx(4)
	if trait.GetSelectedLineIdx() != 4 {
		t.Errorf("SetSelectedLineIdx: got %d, want 4", trait.GetSelectedLineIdx())
	}

	trait.MoveSelectedLine(-1)
	if trait.GetSelectedLineIdx() != 3 {
		t.Errorf("MoveSelectedLine: got %d, want 3", trait.GetSelectedLineIdx())
	}

	trait.SetSelectedLineIdx(10)
	trait.ClampSelection()
	if trait.GetSelectedLineIdx() != 4 {
		t.Errorf("ClampSelection: got %d, want 4 (clamped to last)", trait.GetSelectedLineIdx())
	}
}

func TestListContextTrait_HandleLineChange_CallsBoth(t *testing.T) {
	trait, renderCount, previewCount := newTrait(3, 0)
	trait.HandleLineChange()
	if *renderCount != 1 {
		t.Errorf("renderFn called %d times, want 1", *renderCount)
	}
	if *previewCount != 1 {
		t.Errorf("previewFn called %d times, want 1", *previewCount)
	}
}

func TestListContextTrait_HandleLineChange_NilRenderFn(t *testing.T) {
	list := &simpleList{3}
	cursor := NewListCursor(list)
	trait := NewListContextTrait(cursor, nil, func() {})
	// Must not panic
	trait.HandleLineChange()
}

func TestListContextTrait_HandleLineChange_NilPreviewFn(t *testing.T) {
	list := &simpleList{3}
	cursor := NewListCursor(list)
	trait := NewListContextTrait(cursor, func() {}, nil)
	// Must not panic
	trait.HandleLineChange()
}

func TestListContextTrait_GetCursor(t *testing.T) {
	trait, _, _ := newTrait(3, 1)
	if trait.GetCursor() == nil {
		t.Error("GetCursor returned nil")
	}
}
