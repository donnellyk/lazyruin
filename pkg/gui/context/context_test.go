package context

import (
	"testing"

	"kvnd/lazyruin/pkg/models"
)

// --- NotesContext tests ---

func TestNotesContext_Defaults(t *testing.T) {
	ctx := NewNotesContext(func() {}, func() {})

	if ctx.CurrentTab != NotesTabAll {
		t.Errorf("default tab = %q, want %q", ctx.CurrentTab, NotesTabAll)
	}
	if ctx.GetKey() != "notes" {
		t.Errorf("key = %q, want %q", ctx.GetKey(), "notes")
	}
	if ctx.Title() != "Notes" {
		t.Errorf("title = %q, want %q", ctx.Title(), "Notes")
	}
	if !ctx.IsFocusable() {
		t.Error("expected focusable")
	}
	if ctx.Selected() != nil {
		t.Error("expected nil selected on empty items")
	}
	if ctx.GetSelectedItemId() != "" {
		t.Error("expected empty selected item ID on empty items")
	}
}

func TestNotesContext_Selected(t *testing.T) {
	ctx := NewNotesContext(func() {}, func() {})
	ctx.Items = []models.Note{
		{UUID: "aaa", Title: "First"},
		{UUID: "bbb", Title: "Second"},
		{UUID: "ccc", Title: "Third"},
	}

	sel := ctx.Selected()
	if sel == nil || sel.UUID != "aaa" {
		t.Errorf("Selected() = %v, want UUID=aaa", sel)
	}

	ctx.SetSelectedLineIdx(2)
	sel = ctx.Selected()
	if sel == nil || sel.UUID != "ccc" {
		t.Errorf("Selected() after SetSelectedLineIdx(2) = %v, want UUID=ccc", sel)
	}
}

func TestNotesContext_SelectedItemId(t *testing.T) {
	ctx := NewNotesContext(func() {}, func() {})
	ctx.Items = []models.Note{
		{UUID: "aaa", Title: "First"},
		{UUID: "bbb", Title: "Second"},
	}

	if ctx.GetSelectedItemId() != "aaa" {
		t.Errorf("GetSelectedItemId() = %q, want %q", ctx.GetSelectedItemId(), "aaa")
	}
}

func TestNotesContext_TabIndex(t *testing.T) {
	ctx := NewNotesContext(func() {}, func() {})

	if ctx.TabIndex() != 0 {
		t.Errorf("TabIndex for 'all' = %d, want 0", ctx.TabIndex())
	}

	ctx.CurrentTab = NotesTabToday
	if ctx.TabIndex() != 1 {
		t.Errorf("TabIndex for 'today' = %d, want 1", ctx.TabIndex())
	}

	ctx.CurrentTab = NotesTabRecent
	if ctx.TabIndex() != 2 {
		t.Errorf("TabIndex for 'recent' = %d, want 2", ctx.TabIndex())
	}
}

func TestNotesContext_GetList(t *testing.T) {
	ctx := NewNotesContext(func() {}, func() {})
	ctx.Items = []models.Note{
		{UUID: "aaa"},
		{UUID: "bbb"},
	}

	list := ctx.GetList()
	if list.Len() != 2 {
		t.Errorf("list.Len() = %d, want 2", list.Len())
	}
	if list.GetSelectedItemId() != "aaa" {
		t.Errorf("list.GetSelectedItemId() = %q, want %q", list.GetSelectedItemId(), "aaa")
	}
	if list.FindIndexById("bbb") != 1 {
		t.Errorf("FindIndexById('bbb') = %d, want 1", list.FindIndexById("bbb"))
	}
	if list.FindIndexById("missing") != -1 {
		t.Errorf("FindIndexById('missing') = %d, want -1", list.FindIndexById("missing"))
	}
}

func TestNotesContext_Selected_OutOfBoundsClampsToZero(t *testing.T) {
	ctx := NewNotesContext(func() {}, func() {})
	ctx.Items = []models.Note{{UUID: "aaa"}}
	ctx.SetSelectedLineIdx(99)

	sel := ctx.Selected()
	if sel == nil || sel.UUID != "aaa" {
		t.Errorf("Selected() with OOB index should clamp to 0, got %v", sel)
	}
}

// --- TagsContext tests ---

func TestTagsContext_Defaults(t *testing.T) {
	ctx := NewTagsContext(func() {}, func() {})

	if ctx.CurrentTab != TagsTabAll {
		t.Errorf("default tab = %q, want %q", ctx.CurrentTab, TagsTabAll)
	}
	if ctx.GetKey() != "tags" {
		t.Errorf("key = %q, want %q", ctx.GetKey(), "tags")
	}
	if ctx.Selected() != nil {
		t.Error("expected nil selected on empty items")
	}
}

func TestTagsContext_FilteredItems(t *testing.T) {
	ctx := NewTagsContext(func() {}, func() {})
	ctx.Items = []models.Tag{
		{Name: "meeting", Scope: []string{"global"}},
		{Name: "todo", Scope: []string{"inline"}},
		{Name: "work", Scope: []string{"global", "inline"}},
	}

	// All tab
	ctx.CurrentTab = TagsTabAll
	if len(ctx.FilteredItems()) != 3 {
		t.Errorf("All tab: %d items, want 3", len(ctx.FilteredItems()))
	}

	// Global tab
	ctx.CurrentTab = TagsTabGlobal
	filtered := ctx.FilteredItems()
	if len(filtered) != 2 {
		t.Fatalf("Global tab: %d items, want 2", len(filtered))
	}
	if filtered[0].Name != "meeting" || filtered[1].Name != "work" {
		t.Errorf("Global tab: unexpected items %v", filtered)
	}

	// Inline tab
	ctx.CurrentTab = TagsTabInline
	filtered = ctx.FilteredItems()
	if len(filtered) != 2 {
		t.Fatalf("Inline tab: %d items, want 2", len(filtered))
	}
	if filtered[0].Name != "todo" || filtered[1].Name != "work" {
		t.Errorf("Inline tab: unexpected items %v", filtered)
	}
}

func TestTagsContext_Selected(t *testing.T) {
	ctx := NewTagsContext(func() {}, func() {})
	ctx.Items = []models.Tag{
		{Name: "meeting", Scope: []string{"global"}},
		{Name: "todo", Scope: []string{"inline"}},
	}

	sel := ctx.Selected()
	if sel == nil || sel.Name != "meeting" {
		t.Errorf("Selected() = %v, want meeting", sel)
	}
}

func TestTagsContext_TabIndex(t *testing.T) {
	ctx := NewTagsContext(func() {}, func() {})

	tests := []struct {
		tab  TagsTab
		want int
	}{
		{TagsTabAll, 0},
		{TagsTabGlobal, 1},
		{TagsTabInline, 2},
	}

	for _, tc := range tests {
		ctx.CurrentTab = tc.tab
		if ctx.TabIndex() != tc.want {
			t.Errorf("TabIndex for %q = %d, want %d", tc.tab, ctx.TabIndex(), tc.want)
		}
	}
}

// --- QueriesContext tests ---

func TestQueriesContext_Defaults(t *testing.T) {
	ctx := NewQueriesContext(func() {}, func() {}, func() {}, func() {})

	if ctx.CurrentTab != QueriesTabQueries {
		t.Errorf("default tab = %q, want %q", ctx.CurrentTab, QueriesTabQueries)
	}
	if ctx.GetKey() != "queries" {
		t.Errorf("key = %q, want %q", ctx.GetKey(), "queries")
	}
	if ctx.SelectedQuery() != nil {
		t.Error("expected nil SelectedQuery on empty items")
	}
	if ctx.SelectedParent() != nil {
		t.Error("expected nil SelectedParent on empty items")
	}
}

func TestQueriesContext_SelectedQuery(t *testing.T) {
	ctx := NewQueriesContext(func() {}, func() {}, func() {}, func() {})
	ctx.Queries = []models.Query{
		{Name: "recent", Query: "created:this-week"},
		{Name: "todos", Query: "#todo"},
	}

	sel := ctx.SelectedQuery()
	if sel == nil || sel.Name != "recent" {
		t.Errorf("SelectedQuery() = %v, want recent", sel)
	}

	ctx.QueriesTrait().SetSelectedLineIdx(1)
	sel = ctx.SelectedQuery()
	if sel == nil || sel.Name != "todos" {
		t.Errorf("SelectedQuery() = %v, want todos", sel)
	}
}

func TestQueriesContext_SelectedParent(t *testing.T) {
	ctx := NewQueriesContext(func() {}, func() {}, func() {}, func() {})
	ctx.Parents = []models.ParentBookmark{
		{Name: "project", UUID: "abc", Title: "Project Notes"},
		{Name: "journal", UUID: "def", Title: "Daily Journal"},
	}

	ctx.CurrentTab = QueriesTabParents
	sel := ctx.SelectedParent()
	if sel == nil || sel.UUID != "abc" {
		t.Errorf("SelectedParent() = %v, want abc", sel)
	}
}

func TestQueriesContext_ActiveTrait_SwitchesByTab(t *testing.T) {
	ctx := NewQueriesContext(func() {}, func() {}, func() {}, func() {})

	ctx.CurrentTab = QueriesTabQueries
	if ctx.ActiveTrait() != ctx.QueriesTrait() {
		t.Error("ActiveTrait should return queries trait on queries tab")
	}

	ctx.CurrentTab = QueriesTabParents
	if ctx.ActiveTrait() != ctx.ParentsTrait() {
		t.Error("ActiveTrait should return parents trait on parents tab")
	}
}

func TestQueriesContext_TabIndex(t *testing.T) {
	ctx := NewQueriesContext(func() {}, func() {}, func() {}, func() {})

	ctx.CurrentTab = QueriesTabQueries
	if ctx.TabIndex() != 0 {
		t.Errorf("TabIndex for queries = %d, want 0", ctx.TabIndex())
	}

	ctx.CurrentTab = QueriesTabParents
	if ctx.TabIndex() != 1 {
		t.Errorf("TabIndex for parents = %d, want 1", ctx.TabIndex())
	}
}

func TestQueriesContext_ActiveItemCount(t *testing.T) {
	ctx := NewQueriesContext(func() {}, func() {}, func() {}, func() {})
	ctx.Queries = []models.Query{{Name: "q1"}, {Name: "q2"}}
	ctx.Parents = []models.ParentBookmark{{Name: "p1"}}

	ctx.CurrentTab = QueriesTabQueries
	if ctx.ActiveItemCount() != 2 {
		t.Errorf("ActiveItemCount on queries = %d, want 2", ctx.ActiveItemCount())
	}

	ctx.CurrentTab = QueriesTabParents
	if ctx.ActiveItemCount() != 1 {
		t.Errorf("ActiveItemCount on parents = %d, want 1", ctx.ActiveItemCount())
	}
}

func TestQueriesContext_GetSelectedItemId_ByTab(t *testing.T) {
	ctx := NewQueriesContext(func() {}, func() {}, func() {}, func() {})
	ctx.Queries = []models.Query{{Name: "q1"}}
	ctx.Parents = []models.ParentBookmark{{UUID: "p-uuid"}}

	ctx.CurrentTab = QueriesTabQueries
	if ctx.GetSelectedItemId() != "q1" {
		t.Errorf("GetSelectedItemId on queries = %q, want q1", ctx.GetSelectedItemId())
	}

	ctx.CurrentTab = QueriesTabParents
	if ctx.GetSelectedItemId() != "p-uuid" {
		t.Errorf("GetSelectedItemId on parents = %q, want p-uuid", ctx.GetSelectedItemId())
	}
}

// --- TabIndexOf tests ---

func TestTabIndexOf(t *testing.T) {
	tabs := []string{"a", "b", "c"}

	if TabIndexOf(tabs, "a") != 0 {
		t.Errorf("TabIndexOf('a') = %d, want 0", TabIndexOf(tabs, "a"))
	}
	if TabIndexOf(tabs, "c") != 2 {
		t.Errorf("TabIndexOf('c') = %d, want 2", TabIndexOf(tabs, "c"))
	}
	if TabIndexOf(tabs, "missing") != 0 {
		t.Errorf("TabIndexOf('missing') = %d, want 0 (default)", TabIndexOf(tabs, "missing"))
	}
}

// --- CalendarContext tests ---

func TestCalendarContext_Defaults(t *testing.T) {
	ctx := NewCalendarContext()

	if ctx.GetKey() != "calendarGrid" {
		t.Errorf("key = %q, want calendarGrid", ctx.GetKey())
	}
	if ctx.Title() != "Calendar" {
		t.Errorf("title = %q, want Calendar", ctx.Title())
	}
	if !ctx.IsFocusable() {
		t.Error("expected focusable")
	}
	if ctx.State != nil {
		t.Error("expected nil initial state")
	}

	viewNames := ctx.GetViewNames()
	if len(viewNames) != 3 {
		t.Fatalf("expected 3 view names, got %d", len(viewNames))
	}
	if ctx.GetPrimaryViewName() != "calendarGrid" {
		t.Errorf("primary view = %q, want calendarGrid", ctx.GetPrimaryViewName())
	}
}

// --- CaptureContext tests ---

func TestCaptureContext_Defaults(t *testing.T) {
	ctx := NewCaptureContext()

	if ctx.GetKey() != "capture" {
		t.Errorf("key = %q, want capture", ctx.GetKey())
	}
	if ctx.Title() != "Capture" {
		t.Errorf("title = %q, want Capture", ctx.Title())
	}
	if ctx.Parent != nil {
		t.Error("expected nil parent")
	}
	if ctx.Completion == nil {
		t.Error("expected non-nil completion state")
	}
}
