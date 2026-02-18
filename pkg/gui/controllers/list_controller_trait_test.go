package controllers

import (
	"testing"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// simpleList implements context.ICursorList.
type simpleList struct{ n int }

func (s *simpleList) Len() int { return s.n }

// newTraitFor builds a ListControllerTrait[string] with the given items and starting index.
// It returns the trait and a pointer to a counter that increments each time HandleLineChange
// is called (tracked via the renderFn callback, which HandleLineChange invokes).
func newTraitFor(items []string, startIdx int) (*ListControllerTrait[string], *int) {
	renderCalls := 0
	list := &simpleList{len(items)}
	cursor := context.NewListCursor(list)
	cursor.SetSelectedLineIdx(startIdx)
	inner := context.NewListContextTrait(cursor, func() { renderCalls++ }, nil)

	lct := NewListControllerTrait(
		nil, // ControllerCommon not used by tested methods
		nil, // getContext not used by tested methods
		func() []string { return items },
		func() *context.ListContextTrait { return inner },
	)
	return lct, &renderCalls
}

// --- withItem ---

func TestWithItem_CallsHandlerWithSelectedItem(t *testing.T) {
	items := []string{"a", "b", "c"}
	lct, _ := newTraitFor(items, 1)

	got := ""
	err := lct.withItem(func(s string) error {
		got = s
		return nil
	})()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "b" {
		t.Errorf("got %q, want %q", got, "b")
	}
}

func TestWithItem_EmptyList_NoOp(t *testing.T) {
	lct, _ := newTraitFor([]string{}, 0)
	called := false
	err := lct.withItem(func(s string) error {
		called = true
		return nil
	})()
	if err != nil || called {
		t.Error("withItem should be a no-op on empty list")
	}
}

func TestWithItem_OOBCursor_NoOp(t *testing.T) {
	// Start with 1 item; move cursor past it via direct manipulation.
	list := &simpleList{1}
	cursor := context.NewListCursor(list)
	cursor.SetSelectedLineIdx(5) // deliberately OOB
	inner := context.NewListContextTrait(cursor, nil, nil)
	lct := NewListControllerTrait(
		nil, nil,
		func() []string { return []string{"a"} },
		func() *context.ListContextTrait { return inner },
	)

	called := false
	err := lct.withItem(func(s string) error {
		called = true
		return nil
	})()
	if err != nil || called {
		t.Error("withItem should be a no-op when cursor is out of range")
	}
}

// --- singleItemSelected ---

func TestSingleItemSelected_EmptyList(t *testing.T) {
	lct, _ := newTraitFor([]string{}, 0)
	reason := lct.singleItemSelected()()
	if reason == nil {
		t.Fatal("expected a DisabledReason for empty list")
	}
	if reason.Text != "No items" {
		t.Errorf("reason.Text = %q, want %q", reason.Text, "No items")
	}
}

func TestSingleItemSelected_OOBCursor(t *testing.T) {
	list := &simpleList{1}
	cursor := context.NewListCursor(list)
	cursor.SetSelectedLineIdx(5)
	inner := context.NewListContextTrait(cursor, nil, nil)
	lct := NewListControllerTrait(
		nil, nil,
		func() []string { return []string{"a"} },
		func() *context.ListContextTrait { return inner },
	)

	reason := lct.singleItemSelected()()
	if reason == nil {
		t.Fatal("expected a DisabledReason for out-of-bounds cursor")
	}
	if reason.Text != "No item selected" {
		t.Errorf("reason.Text = %q, want %q", reason.Text, "No item selected")
	}
}

func TestSingleItemSelected_ValidItem_Nil(t *testing.T) {
	lct, _ := newTraitFor([]string{"a", "b"}, 1)
	reason := lct.singleItemSelected()()
	if reason != nil {
		t.Errorf("expected nil reason for valid selection, got %v", reason)
	}
}

// --- require ---

func TestRequire_AllNil_ReturnsNil(t *testing.T) {
	lct, _ := newTraitFor([]string{"a"}, 0)
	nilFn := func() *types.DisabledReason { return nil }
	if got := lct.require(nilFn, nilFn)(); got != nil {
		t.Errorf("require(nil, nil) = %v, want nil", got)
	}
}

func TestRequire_FirstNonNilWins(t *testing.T) {
	lct, _ := newTraitFor([]string{"a"}, 0)
	first := &types.DisabledReason{Text: "first"}
	second := &types.DisabledReason{Text: "second"}
	secondCalled := false
	got := lct.require(
		func() *types.DisabledReason { return first },
		func() *types.DisabledReason { secondCalled = true; return second },
	)()
	if got != first {
		t.Errorf("require returned wrong reason: got %v, want first", got)
	}
	if secondCalled {
		t.Error("second producer should not be called when first returns non-nil")
	}
}

func TestRequire_SkipsNil_ReturnsSecond(t *testing.T) {
	lct, _ := newTraitFor([]string{"a"}, 0)
	second := &types.DisabledReason{Text: "second"}
	got := lct.require(
		func() *types.DisabledReason { return nil },
		func() *types.DisabledReason { return second },
	)()
	if got != second {
		t.Errorf("require = %v, want second", got)
	}
}

// --- navigation ---

// traitWithCursor builds a ListControllerTrait with an externally accessible cursor.
// Returns the trait, cursor (for inspecting final index), and renderCalls counter.
func traitWithCursor(items []string, startIdx int) (*ListControllerTrait[string], *context.ListCursor, *int) {
	renderCalls := 0
	list := &simpleList{len(items)}
	cursor := context.NewListCursor(list)
	cursor.SetSelectedLineIdx(startIdx)
	inner := context.NewListContextTrait(cursor, func() { renderCalls++ }, nil)
	lct := NewListControllerTrait(
		nil, nil,
		func() []string { return items },
		func() *context.ListContextTrait { return inner },
	)
	return lct, cursor, &renderCalls
}

func TestNextItem_MidList_MovesAndCallsHandleLineChange(t *testing.T) {
	lct, cursor, renders := traitWithCursor([]string{"a", "b", "c"}, 1)
	_ = lct.nextItem()
	if cursor.GetSelectedLineIdx() != 2 {
		t.Errorf("idx = %d, want 2", cursor.GetSelectedLineIdx())
	}
	if *renders != 1 {
		t.Errorf("HandleLineChange (via renderFn) called %d times, want 1", *renders)
	}
}

func TestNextItem_AtEnd_NoMove(t *testing.T) {
	lct, cursor, renders := traitWithCursor([]string{"a", "b", "c"}, 2)
	_ = lct.nextItem()
	if cursor.GetSelectedLineIdx() != 2 {
		t.Errorf("idx = %d, want 2 (unchanged)", cursor.GetSelectedLineIdx())
	}
	if *renders != 0 {
		t.Errorf("HandleLineChange should not be called at list end, got %d", *renders)
	}
}

func TestPrevItem_MidList_MovesAndCallsHandleLineChange(t *testing.T) {
	lct, cursor, renders := traitWithCursor([]string{"a", "b", "c"}, 1)
	_ = lct.prevItem()
	if cursor.GetSelectedLineIdx() != 0 {
		t.Errorf("idx = %d, want 0", cursor.GetSelectedLineIdx())
	}
	if *renders != 1 {
		t.Errorf("HandleLineChange called %d times, want 1", *renders)
	}
}

func TestPrevItem_AtStart_NoMove(t *testing.T) {
	lct, cursor, renders := traitWithCursor([]string{"a", "b", "c"}, 0)
	_ = lct.prevItem()
	if cursor.GetSelectedLineIdx() != 0 {
		t.Errorf("idx = %d, want 0 (unchanged)", cursor.GetSelectedLineIdx())
	}
	if *renders != 0 {
		t.Errorf("HandleLineChange should not be called at list start, got %d", *renders)
	}
}

func TestGoTop_GoesToZeroAndCallsHandleLineChange(t *testing.T) {
	lct, cursor, renders := traitWithCursor([]string{"a", "b", "c"}, 2)
	_ = lct.goTop()
	if cursor.GetSelectedLineIdx() != 0 {
		t.Errorf("idx = %d, want 0", cursor.GetSelectedLineIdx())
	}
	if *renders != 1 {
		t.Errorf("HandleLineChange called %d times, want 1", *renders)
	}
}

func TestGoTop_AlreadyAtTop_StillCallsHandleLineChange(t *testing.T) {
	lct, _, renders := traitWithCursor([]string{"a", "b"}, 0)
	_ = lct.goTop()
	if *renders != 1 {
		t.Errorf("HandleLineChange called %d times, want 1", *renders)
	}
}

func TestGoBottom_GoesToLastAndCallsHandleLineChange(t *testing.T) {
	lct, cursor, renders := traitWithCursor([]string{"a", "b", "c"}, 0)
	_ = lct.goBottom()
	if cursor.GetSelectedLineIdx() != 2 {
		t.Errorf("idx = %d, want 2", cursor.GetSelectedLineIdx())
	}
	if *renders != 1 {
		t.Errorf("HandleLineChange called %d times, want 1", *renders)
	}
}

func TestGoBottom_EmptyList_NoMove(t *testing.T) {
	lct, _, renders := traitWithCursor([]string{}, 0)
	_ = lct.goBottom()
	if *renders != 0 {
		t.Errorf("HandleLineChange should not be called on empty list, got %d", *renders)
	}
}
