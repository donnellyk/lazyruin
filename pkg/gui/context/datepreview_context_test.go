package context

import "testing"

func TestSectionForCard(t *testing.T) {
	state := &DatePreviewState{
		SectionRanges: [3][2]int{
			{0, 3}, // TagPicks: cards 0-2
			{3, 5}, // TodoPicks: cards 3-4
			{5, 8}, // Notes: cards 5-7
		},
	}

	tests := []struct {
		idx  int
		want DatePreviewSection
	}{
		{0, SectionTagPicks},
		{2, SectionTagPicks},
		{3, SectionTodoPicks},
		{4, SectionTodoPicks},
		{5, SectionNotes},
		{7, SectionNotes},
		{8, SectionNotes}, // out of range, fallback
	}
	for _, tt := range tests {
		if got := state.SectionForCard(tt.idx); got != tt.want {
			t.Errorf("SectionForCard(%d) = %d, want %d", tt.idx, got, tt.want)
		}
	}
}

func TestSectionForCard_EmptySection(t *testing.T) {
	state := &DatePreviewState{
		SectionRanges: [3][2]int{
			{0, 0}, // TagPicks: empty
			{0, 2}, // TodoPicks: cards 0-1
			{2, 5}, // Notes: cards 2-4
		},
	}
	if got := state.SectionForCard(0); got != SectionTodoPicks {
		t.Errorf("SectionForCard(0) = %d, want SectionTodoPicks", got)
	}
	if got := state.SectionForCard(3); got != SectionNotes {
		t.Errorf("SectionForCard(3) = %d, want SectionNotes", got)
	}
}

func TestSectionForLine(t *testing.T) {
	state := &DatePreviewState{
		SectionLineRanges: [3][2]int{
			{0, 20},
			{20, 40},
			{40, 60},
		},
	}

	tests := []struct {
		line int
		want DatePreviewSection
	}{
		{0, SectionTagPicks},
		{19, SectionTagPicks},
		{20, SectionTodoPicks},
		{39, SectionTodoPicks},
		{40, SectionNotes},
		{59, SectionNotes},
		{60, SectionNotes}, // out of range, fallback
	}
	for _, tt := range tests {
		if got := state.SectionForLine(tt.line); got != tt.want {
			t.Errorf("SectionForLine(%d) = %d, want %d", tt.line, got, tt.want)
		}
	}
}

func TestLocalCardIdx(t *testing.T) {
	state := &DatePreviewState{
		SectionRanges: [3][2]int{
			{0, 3},
			{3, 5},
			{5, 8},
		},
	}

	tests := []struct {
		global int
		want   int
	}{
		{0, 0},
		{2, 2},
		{3, 0},
		{4, 1},
		{5, 0},
		{7, 2},
	}
	for _, tt := range tests {
		if got := state.LocalCardIdx(tt.global); got != tt.want {
			t.Errorf("LocalCardIdx(%d) = %d, want %d", tt.global, got, tt.want)
		}
	}
}

func TestDatePreviewContext_CardCount(t *testing.T) {
	ctx := NewDatePreviewContext(NewSharedNavHistory())
	ctx.DatePreviewState.TargetDate = "2026-02-23"

	if ctx.CardCount() != 0 {
		t.Errorf("CardCount() = %d, want 0 for empty state", ctx.CardCount())
	}
}
