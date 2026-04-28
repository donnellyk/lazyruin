package helpers

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/config"
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/donnellyk/lazyruin/pkg/testutil"
)

func newTestNotesHomeHelper(mock *testutil.MockExecutor, sections []config.NotesPaneSection) (*NotesHomeHelper, *mockGuiCommon) {
	gui := &mockGuiCommon{
		contexts: &context.ContextTree{
			NotesHome: context.NewNotesHomeContext(),
		},
	}
	ruinCmd := commands.NewRuinCommandWithExecutor(mock, "/mock")
	common := NewHelperCommon(ruinCmd, nil, gui)
	return NewNotesHomeHelper(common, func() []config.NotesPaneSection { return sections }), gui
}

func TestBuildRows_HardcodedOnly(t *testing.T) {
	mock := testutil.NewMockExecutor() // empty parents & queries
	helper, _ := newTestNotesHomeHelper(mock, nil)

	rows := helper.BuildRows()

	// Group 1: Inbox.
	// Blank.
	// Group 2: Today, Next 7 Days.
	// (Pinned section omitted because no parents/queries.)
	if len(rows) != 4 {
		t.Fatalf("expected 4 rows (Inbox, blank, Today, Next7), got %d: %+v", len(rows), rows)
	}
	expect := []struct {
		isHeader bool
		blank    bool
		title    string
	}{
		{false, false, "Inbox"},
		{false, true, ""},
		{false, false, "Today"},
		{false, false, "Next 7 Days"},
	}
	for i, w := range expect {
		got := rows[i]
		if got.IsHeader != w.isHeader || got.Blank != w.blank || got.Title != w.title {
			t.Errorf("row[%d] = %+v, want header=%v blank=%v title=%q", i, got, w.isHeader, w.blank, w.title)
		}
	}
}

func TestBuildRows_PinnedSection(t *testing.T) {
	mock := testutil.NewMockExecutor().
		WithParents(
			models.ParentBookmark{Name: "p1", UUID: "uuid-1", Title: "Parent One"},
			models.ParentBookmark{Name: "p2", UUID: "uuid-2", Title: "Parent Two"},
		).
		WithQueries(
			models.Query{Name: "q1", Query: "#x"},
			models.Query{Name: "q2", Query: "#y"},
		)
	helper, _ := newTestNotesHomeHelper(mock, nil)

	rows := helper.BuildRows()

	// Find Pinned header.
	pinnedIdx := -1
	for i, r := range rows {
		if r.IsHeader && r.Title == "Pinned" {
			pinnedIdx = i
			break
		}
	}
	if pinnedIdx == -1 {
		t.Fatal("expected Pinned header in rows")
	}

	// Expect: Parent One, Parent Two, blank, q1, q2 after the header.
	expectedAfter := []struct {
		blank  bool
		isItem bool
		title  string
		itemID string
	}{
		{false, true, "Parent One", "parent:uuid-1"},
		{false, true, "Parent Two", "parent:uuid-2"},
		{true, false, "", ""},
		{false, true, "q1", "query:q1"},
		{false, true, "q2", "query:q2"},
	}
	for i, w := range expectedAfter {
		idx := pinnedIdx + 1 + i
		if idx >= len(rows) {
			t.Fatalf("missing row at idx %d", idx)
		}
		got := rows[idx]
		if got.Blank != w.blank {
			t.Errorf("row[%d].Blank = %v, want %v", idx, got.Blank, w.blank)
		}
		if !w.blank {
			if got.Title != w.title {
				t.Errorf("row[%d].Title = %q, want %q", idx, got.Title, w.title)
			}
			if got.ItemID != w.itemID {
				t.Errorf("row[%d].ItemID = %q, want %q", idx, got.ItemID, w.itemID)
			}
		}
	}
}

func TestBuildRows_PinnedOmitsEmptySubgroups(t *testing.T) {
	// Only queries, no parents → no leading blank between absent parents
	// and the queries group.
	mock := testutil.NewMockExecutor().
		WithQueries(models.Query{Name: "only", Query: "#x"})
	helper, _ := newTestNotesHomeHelper(mock, nil)

	rows := helper.BuildRows()

	// Find Pinned header.
	pinnedIdx := -1
	for i, r := range rows {
		if r.IsHeader && r.Title == "Pinned" {
			pinnedIdx = i
			break
		}
	}
	if pinnedIdx == -1 {
		t.Fatal("expected Pinned header")
	}
	// Next row should be the query item, no preceding blank.
	if pinnedIdx+1 >= len(rows) || rows[pinnedIdx+1].Blank {
		t.Errorf("row immediately after Pinned should be the query item, got %+v", rows[pinnedIdx+1:])
	}
	if rows[pinnedIdx+1].Title != "only" {
		t.Errorf("row[%d].Title = %q, want only", pinnedIdx+1, rows[pinnedIdx+1].Title)
	}
}

func TestBuildRows_CustomSections(t *testing.T) {
	mock := testutil.NewMockExecutor() // no parents / queries
	sections := []config.NotesPaneSection{
		{
			Title: "Reading Queue",
			Items: []config.NotesPaneSectionItem{
				{Title: "Articles", Embed: "![[search: #article]]"},
			},
		},
		{
			// untitled section
			Items: []config.NotesPaneSectionItem{
				{Title: "Loose item", Embed: "![[search: #loose]]"},
			},
		},
	}
	helper, _ := newTestNotesHomeHelper(mock, sections)

	rows := helper.BuildRows()

	// Find Reading Queue header.
	rqIdx := -1
	for i, r := range rows {
		if r.IsHeader && r.Title == "Reading Queue" {
			rqIdx = i
			break
		}
	}
	if rqIdx == -1 {
		t.Fatal("expected Reading Queue header in rows")
	}
	// Next row: Articles item.
	if rqIdx+1 >= len(rows) || rows[rqIdx+1].Title != "Articles" {
		t.Errorf("expected Articles item after Reading Queue header, got %+v", rows[rqIdx+1:])
	}
	if rows[rqIdx+1].Action.Kind != context.NotesHomeActionEmbed {
		t.Errorf("Articles item action Kind = %v, want Embed", rows[rqIdx+1].Action.Kind)
	}

	// Untitled section: should follow with a blank row but no header. Find
	// the loose item.
	looseIdx := -1
	for i, r := range rows {
		if r.Title == "Loose item" {
			looseIdx = i
			break
		}
	}
	if looseIdx == -1 {
		t.Fatal("expected Loose item in rows")
	}
	// Walk back and confirm we don't pass through a header before hitting
	// the previous group's last item.
	sawHeader := false
	for i := looseIdx - 1; i >= 0; i-- {
		if rows[i].IsHeader {
			sawHeader = true
			if rows[i].Title != "Reading Queue" {
				// Should never see a header for the untitled section.
				t.Errorf("untitled section unexpectedly produced header %q", rows[i].Title)
			}
			break
		}
	}
	if !sawHeader {
		t.Error("expected to encounter Reading Queue header before Loose item")
	}
}

func TestBuildRows_SkipsMalformedCustomItems(t *testing.T) {
	mock := testutil.NewMockExecutor()
	sections := []config.NotesPaneSection{
		{
			Title: "Mixed",
			Items: []config.NotesPaneSectionItem{
				{Title: "", Embed: "![[search: #x]]"}, // missing title — drop
				{Title: "Good", Embed: "![[search: #y]]"},
				{Title: "Bad", Embed: ""}, // missing embed — drop
			},
		},
	}
	helper, _ := newTestNotesHomeHelper(mock, sections)

	rows := helper.BuildRows()

	// Find the Mixed header.
	hdrIdx := -1
	for i, r := range rows {
		if r.IsHeader && r.Title == "Mixed" {
			hdrIdx = i
			break
		}
	}
	if hdrIdx == -1 {
		t.Fatal("expected Mixed header")
	}
	// Only one valid item should follow.
	if hdrIdx+1 >= len(rows) || rows[hdrIdx+1].Title != "Good" {
		t.Errorf("expected single Good item after Mixed header, got %+v", rows[hdrIdx+1:])
	}
	// No second item afterwards (or, if present, it's a different section's blank).
	if hdrIdx+2 < len(rows) && !rows[hdrIdx+2].Blank && !rows[hdrIdx+2].IsHeader {
		// Allow trailing rows from later sections; just confirm no Bad/empty title leaks.
		if rows[hdrIdx+2].Title == "Bad" || rows[hdrIdx+2].Title == "" {
			t.Errorf("malformed item leaked into rows: %+v", rows[hdrIdx+2])
		}
	}
}

func TestBuildRows_EmptyCustomSectionDropped(t *testing.T) {
	mock := testutil.NewMockExecutor()
	sections := []config.NotesPaneSection{
		{
			Title: "EmptyOne",
			Items: []config.NotesPaneSectionItem{}, // no items
		},
		{
			Title: "Real",
			Items: []config.NotesPaneSectionItem{
				{Title: "Item", Embed: "![[search: #x]]"},
			},
		},
	}
	helper, _ := newTestNotesHomeHelper(mock, sections)

	rows := helper.BuildRows()
	for _, r := range rows {
		if r.IsHeader && r.Title == "EmptyOne" {
			t.Error("EmptyOne header rendered despite having no valid items")
		}
	}
}
