package context

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/models"
)

func TestCardListContext_SnapshotRoundTrip(t *testing.T) {
	ctx := NewCardListContext()
	ctx.SetTitle("Search: foo")
	ctx.Cards = []models.Note{
		{UUID: "u1", Title: "One"},
		{UUID: "u2", Title: "Two"},
	}
	ctx.SelectedCardIdx = 1
	ctx.FilterText = "o"
	ns := ctx.NavState()
	ns.CursorLine = 7
	ns.ScrollOffset = 3
	ds := ctx.DisplayState()
	ds.RenderMarkdown = false
	ds.ShowFrontmatter = true

	snap := ctx.CaptureSnapshot()

	// Reset to different values so restore is meaningful.
	ctx.Cards = nil
	ctx.SelectedCardIdx = 0
	ctx.FilterText = ""
	ns.CursorLine = 0
	ns.ScrollOffset = 0
	ds.RenderMarkdown = true
	ds.ShowFrontmatter = false

	if err := ctx.RestoreSnapshot(snap); err != nil {
		t.Fatalf("RestoreSnapshot: %v", err)
	}

	if ctx.Title() != "Search: foo" {
		t.Errorf("Title = %q, want %q", ctx.Title(), "Search: foo")
	}
	if len(ctx.Cards) != 2 || ctx.Cards[0].UUID != "u1" {
		t.Errorf("Cards not restored: %+v", ctx.Cards)
	}
	if ctx.SelectedCardIdx != 1 {
		t.Errorf("SelectedCardIdx = %d, want 1", ctx.SelectedCardIdx)
	}
	if ctx.FilterText != "o" {
		t.Errorf("FilterText = %q, want o", ctx.FilterText)
	}
	if ns.CursorLine != 7 || ns.ScrollOffset != 3 {
		t.Errorf("view state not restored: cursor=%d scroll=%d", ns.CursorLine, ns.ScrollOffset)
	}
	if ds.RenderMarkdown != false || !ds.ShowFrontmatter {
		t.Errorf("display state not restored: %+v", ds)
	}
}

func TestCardListContext_RestoreViaRequery(t *testing.T) {
	ctx := NewCardListContext()
	ctx.SetTitle("Query")
	fresh := []models.Note{{UUID: "fresh-1", Title: "Fresh"}}
	ctx.Source = CardListSource{
		Query: "q",
		Requery: func(filterText string) ([]models.Note, error) {
			return fresh, nil
		},
	}
	ctx.Cards = []models.Note{{UUID: "stale"}}

	snap := ctx.CaptureSnapshot()
	ctx.Cards = nil
	_ = ctx.RestoreSnapshot(snap)

	if len(ctx.Cards) != 1 || ctx.Cards[0].UUID != "fresh-1" {
		t.Errorf("Restore did not re-run Requery: got %+v", ctx.Cards)
	}
}

func TestPickResultsContext_SnapshotRoundTrip(t *testing.T) {
	ctx := NewPickResultsContext()
	ctx.SetTitle("Pick: foo")
	ctx.Results = []models.PickResult{{UUID: "u1"}}
	ctx.SelectedCardIdx = 0
	ctx.FilterText = "f"
	ctx.NavState().CursorLine = 4

	snap := ctx.CaptureSnapshot()
	ctx.Results = nil
	ctx.FilterText = ""
	ctx.NavState().CursorLine = 0

	if err := ctx.RestoreSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	if len(ctx.Results) != 1 || ctx.Results[0].UUID != "u1" {
		t.Errorf("Results not restored: %+v", ctx.Results)
	}
	if ctx.FilterText != "f" {
		t.Errorf("FilterText = %q, want f", ctx.FilterText)
	}
	if ctx.NavState().CursorLine != 4 {
		t.Errorf("CursorLine = %d, want 4", ctx.NavState().CursorLine)
	}
}

func TestComposeContext_SnapshotRoundTrip(t *testing.T) {
	ctx := NewComposeContext()
	ctx.SetTitle("Parent: p1")
	ctx.Note = models.Note{UUID: "n-stale", Title: "Stale"}
	ctx.Parent = models.ParentBookmark{Name: "p1"}
	ctx.NavState().CursorLine = 5

	requeryCalled := false
	ctx.Requery = func() (models.Note, []models.SourceMapEntry, error) {
		requeryCalled = true
		return models.Note{UUID: "n-fresh", Title: "Fresh"}, nil, nil
	}

	snap := ctx.CaptureSnapshot()
	ctx.Note = models.Note{}
	ctx.NavState().CursorLine = 0

	if err := ctx.RestoreSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	if !requeryCalled {
		t.Error("Requery was not invoked on restore")
	}
	if ctx.Note.UUID != "n-fresh" {
		t.Errorf("Note not refreshed: %+v", ctx.Note)
	}
	if ctx.NavState().CursorLine != 5 {
		t.Errorf("CursorLine = %d, want 5", ctx.NavState().CursorLine)
	}
}

func TestDatePreviewContext_SnapshotRoundTrip(t *testing.T) {
	ctx := NewDatePreviewContext()
	ctx.SetTitle("Monday")
	ctx.TargetDate = "2026-04-17"
	ctx.Notes = []models.Note{{UUID: "stale"}}

	ctx.Requery = func() ([]models.PickResult, []models.PickResult, []models.Note, error) {
		return nil, nil, []models.Note{{UUID: "fresh"}}, nil
	}

	snap := ctx.CaptureSnapshot()
	ctx.Notes = nil

	if err := ctx.RestoreSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	if len(ctx.Notes) != 1 || ctx.Notes[0].UUID != "fresh" {
		t.Errorf("DatePreview Notes not re-queried: %+v", ctx.Notes)
	}
	if ctx.TargetDate != "2026-04-17" {
		t.Errorf("TargetDate = %q, want 2026-04-17", ctx.TargetDate)
	}
}

func TestCardListContext_RestoreFallsBackToFrozenOnRequeryError(t *testing.T) {
	ctx := NewCardListContext()
	ctx.Cards = []models.Note{{UUID: "frozen"}}

	failErr := &errStub{"requery fail"}
	ctx.Source = CardListSource{
		Requery: func(filterText string) ([]models.Note, error) {
			return nil, failErr
		},
	}

	snap := ctx.CaptureSnapshot()
	ctx.Cards = nil

	if err := ctx.RestoreSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	if len(ctx.Cards) != 1 || ctx.Cards[0].UUID != "frozen" {
		t.Errorf("Did not fall back to frozen cards: %+v", ctx.Cards)
	}
}

func TestCardListContext_RestoreClampsSelectedCardIdx(t *testing.T) {
	ctx := NewCardListContext()
	ctx.Cards = []models.Note{{UUID: "a"}, {UUID: "b"}, {UUID: "c"}}
	ctx.SelectedCardIdx = 2

	// Requery returns only one note — clamp should apply.
	ctx.Source = CardListSource{
		Requery: func(filterText string) ([]models.Note, error) {
			return []models.Note{{UUID: "a"}}, nil
		},
	}

	snap := ctx.CaptureSnapshot()
	_ = ctx.RestoreSnapshot(snap)

	if ctx.SelectedCardIdx != 0 {
		t.Errorf("SelectedCardIdx = %d, want 0 (clamped)", ctx.SelectedCardIdx)
	}
}

func TestPickResultsContext_RestoreClampsSelectedCardIdx(t *testing.T) {
	ctx := NewPickResultsContext()
	ctx.Results = []models.PickResult{{UUID: "a"}, {UUID: "b"}, {UUID: "c"}}
	ctx.SelectedCardIdx = 2

	ctx.Source = PickResultsSource{
		Requery: func(filterText string) ([]models.PickResult, error) {
			return []models.PickResult{{UUID: "a"}}, nil
		},
	}

	snap := ctx.CaptureSnapshot()
	_ = ctx.RestoreSnapshot(snap)

	if ctx.SelectedCardIdx != 0 {
		t.Errorf("SelectedCardIdx = %d, want 0 (clamped)", ctx.SelectedCardIdx)
	}
}

type errStub struct{ msg string }

func (e *errStub) Error() string { return e.msg }
