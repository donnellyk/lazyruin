package context

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

type DatePreviewSection int

const (
	SectionTagPicks  DatePreviewSection = 0
	SectionTodoPicks DatePreviewSection = 1
	SectionNotes     DatePreviewSection = 2
)

type DatePreviewState struct {
	TargetDate         string
	TagPicks           []models.PickResult
	TodoPicks          []models.PickResult
	Notes              []models.Note
	SectionRanges      [3][2]int
	SectionLineRanges  [3][2]int
	SectionHeaderLines []int
}

type DatePreviewContext struct {
	BaseContext
	PreviewContextTrait
	*DatePreviewState
}

func NewDatePreviewContext(navHistory *SharedNavHistory) *DatePreviewContext {
	return &DatePreviewContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.MAIN_CONTEXT,
			Key:       "datePreview",
			ViewName:  "preview",
			Focusable: true,
			Title:     "Date Preview",
		}),
		PreviewContextTrait: NewPreviewContextTrait(navHistory),
		DatePreviewState:    &DatePreviewState{},
	}
}

// IPreviewContext implementation (CardCount varies per context; the rest are
// provided by the embedded PreviewContextTrait).

func (self *DatePreviewContext) CardCount() int {
	return len(self.TagPicks) + len(self.TodoPicks) + len(self.Notes)
}

func (s *DatePreviewState) SectionForCard(idx int) DatePreviewSection {
	for i, r := range s.SectionRanges {
		if idx >= r[0] && idx < r[1] {
			return DatePreviewSection(i)
		}
	}
	return SectionNotes
}

func (s *DatePreviewState) SectionForLine(line int) DatePreviewSection {
	for i, r := range s.SectionLineRanges {
		if line >= r[0] && line < r[1] {
			return DatePreviewSection(i)
		}
	}
	return SectionNotes
}

func (s *DatePreviewState) LocalCardIdx(globalIdx int) int {
	sec := s.SectionForCard(globalIdx)
	return globalIdx - s.SectionRanges[sec][0]
}

var _ types.Context = &DatePreviewContext{}
var _ IPreviewContext = &DatePreviewContext{}
