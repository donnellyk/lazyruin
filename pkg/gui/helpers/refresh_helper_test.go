package helpers

import (
	"testing"

	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// mockList implements types.IList for testing.
type mockList struct {
	ids         []string
	selectedIdx int
}

func (m *mockList) GetSelectedLineIdx() int { return m.selectedIdx }
func (m *mockList) SetSelectedLineIdx(i int) {
	if i >= 0 && i < len(m.ids) {
		m.selectedIdx = i
	}
}
func (m *mockList) MoveSelectedLine(d int) { m.selectedIdx += d }
func (m *mockList) ClampSelection() {
	if len(m.ids) == 0 {
		m.selectedIdx = 0
		return
	}
	if m.selectedIdx >= len(m.ids) {
		m.selectedIdx = len(m.ids) - 1
	}
	if m.selectedIdx < 0 {
		m.selectedIdx = 0
	}
}
func (m *mockList) Len() int { return len(m.ids) }
func (m *mockList) GetSelectedItemId() string {
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.ids) {
		return m.ids[m.selectedIdx]
	}
	return ""
}
func (m *mockList) FindIndexById(id string) int {
	for i, v := range m.ids {
		if v == id {
			return i
		}
	}
	return -1
}

// mockListContext implements types.IListContext with a mockList.
// All methods not under test are stubs.
type mockListContext struct {
	id   string    // return value for GetSelectedItemId
	list *mockList // backing list
}

func (m *mockListContext) GetSelectedItemId() string { return m.id }
func (m *mockListContext) GetList() types.IList      { return m.list }

// types.IBaseContext stubs
func (m *mockListContext) GetKind() types.ContextKind                            { return 0 }
func (m *mockListContext) GetKey() types.ContextKey                              { return "mock" }
func (m *mockListContext) IsFocusable() bool                                     { return false }
func (m *mockListContext) Title() string                                         { return "" }
func (m *mockListContext) GetViewNames() []string                                { return nil }
func (m *mockListContext) GetPrimaryViewName() string                            { return "" }
func (m *mockListContext) GetKeybindings(types.KeybindingsOpts) []*types.Binding { return nil }
func (m *mockListContext) GetMouseKeybindings(types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return nil
}
func (m *mockListContext) GetOnClick() func() error                       { return nil }
func (m *mockListContext) GetTabClickBindingFn() func(int) error          { return nil }
func (m *mockListContext) AddKeybindingsFn(types.KeybindingsFn)           {}
func (m *mockListContext) AddMouseKeybindingsFn(types.MouseKeybindingsFn) {}
func (m *mockListContext) AddOnFocusFn(func(types.OnFocusOpts))           {}
func (m *mockListContext) AddOnFocusLostFn(func(types.OnFocusLostOpts))   {}
func (m *mockListContext) AddOnRenderToMainFn(func())                     {}

// types.Context stubs
func (m *mockListContext) HandleFocus(types.OnFocusOpts)         {}
func (m *mockListContext) HandleFocusLost(types.OnFocusLostOpts) {}
func (m *mockListContext) HandleRender()                         {}

func newRefreshHelper() *RefreshHelper {
	return NewRefreshHelper(nil) // HelperCommon not used by PreserveSelection
}

func TestPreserveSelection_ItemStillPresent(t *testing.T) {
	// List: ["a", "b", "c"]. Selected "b" at idx 1.
	// After refresh, "b" is still at idx 2 (items shifted). Should restore to 2.
	list := &mockList{ids: []string{"a", "x", "b", "c"}, selectedIdx: 0}
	ctx := &mockListContext{id: "b", list: list}

	h := newRefreshHelper()
	h.PreserveSelection(ctx)

	if list.selectedIdx != 2 {
		t.Errorf("selectedIdx = %d, want 2 (found b at index 2)", list.selectedIdx)
	}
}

func TestPreserveSelection_ItemGone_FallsBackToClamp(t *testing.T) {
	// List: ["a", "b"]. Previously selected "z" (now gone).
	// Cursor was at idx 3 (stale), should clamp to 1 (last).
	list := &mockList{ids: []string{"a", "b"}, selectedIdx: 3}
	ctx := &mockListContext{id: "z", list: list}

	h := newRefreshHelper()
	h.PreserveSelection(ctx)

	if list.selectedIdx != 1 {
		t.Errorf("selectedIdx = %d, want 1 (clamped to last)", list.selectedIdx)
	}
}

func TestPreserveSelection_EmptyID_JustClamps(t *testing.T) {
	// No previous ID stored; cursor is OOB. Should just clamp.
	list := &mockList{ids: []string{"a", "b"}, selectedIdx: 5}
	ctx := &mockListContext{id: "", list: list}

	h := newRefreshHelper()
	h.PreserveSelection(ctx)

	if list.selectedIdx != 1 {
		t.Errorf("selectedIdx = %d, want 1 (clamped to last)", list.selectedIdx)
	}
}

func TestPreserveSelection_EmptyList(t *testing.T) {
	list := &mockList{ids: []string{}, selectedIdx: 0}
	ctx := &mockListContext{id: "x", list: list}

	h := newRefreshHelper()
	h.PreserveSelection(ctx)

	if list.selectedIdx != 0 {
		t.Errorf("selectedIdx = %d, want 0 (empty list)", list.selectedIdx)
	}
}
