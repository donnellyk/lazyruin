package context

import (
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
)

type DatePreviewSection int

const (
	SectionTagPicks  DatePreviewSection = 0
	SectionTodoPicks DatePreviewSection = 1
	SectionNotes     DatePreviewSection = 2
)

// DatePreviewRequery re-runs the three queries that populate the date preview
// (tag picks, todo picks, notes).
type DatePreviewRequery func() (tagPicks, todoPicks []models.PickResult, notes []models.Note, err error)

type DatePreviewState struct {
	TargetDate         string
	TagPicks           []models.PickResult
	TodoPicks          []models.PickResult
	Notes              []models.Note
	SectionRanges      [3][2]int
	SectionLineRanges  [3][2]int
	SectionHeaderLines []int
	Requery            DatePreviewRequery // closure to re-fetch all three sections for TargetDate
}

type DatePreviewContext struct {
	BaseContext
	PreviewContextTrait
	*DatePreviewState
}

func NewDatePreviewContext() *DatePreviewContext {
	return &DatePreviewContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.MAIN_CONTEXT,
			Key:       "datePreview",
			ViewName:  "preview",
			Focusable: true,
			Title:     "Date Preview",
		}),
		PreviewContextTrait: NewPreviewContextTrait(),
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

// datePreviewSnapshot carries view params (TargetDate + Requery closure) and
// view state for DatePreview restoration.
type datePreviewSnapshot struct {
	Title           string
	TargetDate      string
	Requery         DatePreviewRequery
	FrozenTagPicks  []models.PickResult
	FrozenTodoPicks []models.PickResult
	FrozenNotes     []models.Note
	SelectedCardIdx int
	CursorLine      int
	ScrollOffset    int
	Display         PreviewDisplayState
}

func (self *DatePreviewContext) CaptureSnapshot() types.Snapshot {
	ns := self.NavState()
	return &datePreviewSnapshot{
		Title:           self.Title(),
		TargetDate:      self.TargetDate,
		Requery:         self.Requery,
		FrozenTagPicks:  append([]models.PickResult(nil), self.TagPicks...),
		FrozenTodoPicks: append([]models.PickResult(nil), self.TodoPicks...),
		FrozenNotes:     append([]models.Note(nil), self.Notes...),
		SelectedCardIdx: self.SelectedCardIdx,
		CursorLine:      ns.CursorLine,
		ScrollOffset:    ns.ScrollOffset,
		Display:         *self.DisplayState(),
	}
}

func (self *DatePreviewContext) RestoreSnapshot(s types.Snapshot) error {
	snap, ok := s.(*datePreviewSnapshot)
	if !ok || snap == nil {
		return nil
	}
	self.SetTitle(snap.Title)
	self.TargetDate = snap.TargetDate
	self.Requery = snap.Requery
	*self.DisplayState() = snap.Display

	if snap.Requery != nil {
		tag, todo, notes, err := snap.Requery()
		if err == nil {
			self.TagPicks = tag
			self.TodoPicks = todo
			self.Notes = notes
		} else {
			self.TagPicks = append([]models.PickResult(nil), snap.FrozenTagPicks...)
			self.TodoPicks = append([]models.PickResult(nil), snap.FrozenTodoPicks...)
			self.Notes = append([]models.Note(nil), snap.FrozenNotes...)
		}
	} else {
		self.TagPicks = append([]models.PickResult(nil), snap.FrozenTagPicks...)
		self.TodoPicks = append([]models.PickResult(nil), snap.FrozenTodoPicks...)
		self.Notes = append([]models.Note(nil), snap.FrozenNotes...)
	}

	self.SelectedCardIdx = snap.SelectedCardIdx
	ns := self.NavState()
	ns.CursorLine = snap.CursorLine
	ns.ScrollOffset = snap.ScrollOffset
	return nil
}

var _ types.Context = &DatePreviewContext{}
var _ IPreviewContext = &DatePreviewContext{}
var _ types.Snapshotter = &DatePreviewContext{}
