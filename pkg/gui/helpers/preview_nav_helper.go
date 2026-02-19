package helpers

import (
	"strings"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// PreviewNavHelper handles preview navigation: history stack, cursor movement,
// scrolling, and interaction (click, back, focus, editor).
type PreviewNavHelper struct {
	c *HelperCommon
}

// NewPreviewNavHelper creates a new PreviewNavHelper.
func NewPreviewNavHelper(c *HelperCommon) *PreviewNavHelper {
	return &PreviewNavHelper{c: c}
}

func (self *PreviewNavHelper) ctx() *context.PreviewContext {
	return self.c.GuiCommon().Contexts().Preview
}

func (self *PreviewNavHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// --- nav history ---

// PushNavHistory captures the current preview state onto the nav history stack.
func (self *PreviewNavHelper) PushNavHistory() {
	pc := self.ctx()
	if len(pc.Cards) == 0 && len(pc.PickResults) == 0 {
		return
	}

	entry := self.captureCurrentNavEntry()

	// Truncate any forward entries
	if pc.NavIndex >= 0 && pc.NavIndex < len(pc.NavHistory)-1 {
		pc.NavHistory = pc.NavHistory[:pc.NavIndex+1]
	}

	pc.NavHistory = append(pc.NavHistory, entry)
	pc.NavIndex = len(pc.NavHistory) - 1

	// Cap at 50 entries
	if len(pc.NavHistory) > 50 {
		pc.NavHistory = pc.NavHistory[len(pc.NavHistory)-50:]
		pc.NavIndex = len(pc.NavHistory) - 1
	}
}

func (self *PreviewNavHelper) captureCurrentNavEntry() context.NavEntry {
	pc := self.ctx()
	title := ""
	if v := self.view(); v != nil {
		title = v.Title
	}
	return context.NavEntry{
		Cards:             append([]models.Note(nil), pc.Cards...),
		SelectedCardIndex: pc.SelectedCardIndex,
		CursorLine:        pc.CursorLine,
		ScrollOffset:      pc.ScrollOffset,
		Mode:              pc.Mode,
		Title:             title,
		PickResults:       append([]models.PickResult(nil), pc.PickResults...),
	}
}

func (self *PreviewNavHelper) restoreNavEntry(entry context.NavEntry) {
	pc := self.ctx()
	pc.Mode = entry.Mode
	pc.Cards = append([]models.Note(nil), entry.Cards...)
	pc.PickResults = append([]models.PickResult(nil), entry.PickResults...)
	pc.SelectedCardIndex = entry.SelectedCardIndex
	pc.CursorLine = entry.CursorLine
	pc.ScrollOffset = entry.ScrollOffset
	if v := self.view(); v != nil {
		v.Title = entry.Title
		v.SetOrigin(0, entry.ScrollOffset)
	}
	self.c.GuiCommon().RenderPreview()
}

// NavBack navigates backward in history.
func (self *PreviewNavHelper) NavBack() error {
	pc := self.ctx()
	if pc.NavIndex < 0 || len(pc.NavHistory) == 0 {
		return nil
	}

	pc.NavHistory[pc.NavIndex] = self.captureCurrentNavEntry()

	if pc.NavIndex == len(pc.NavHistory)-1 {
		pc.NavHistory = append(pc.NavHistory, self.captureCurrentNavEntry())
	}

	if pc.NavIndex <= 0 {
		return nil
	}

	pc.NavIndex--
	self.restoreNavEntry(pc.NavHistory[pc.NavIndex])
	return nil
}

// NavForward navigates forward in history.
func (self *PreviewNavHelper) NavForward() error {
	pc := self.ctx()
	if pc.NavIndex >= len(pc.NavHistory)-1 {
		return nil
	}

	pc.NavHistory[pc.NavIndex] = self.captureCurrentNavEntry()

	pc.NavIndex++
	self.restoreNavEntry(pc.NavHistory[pc.NavIndex])
	return nil
}

// ShowNavHistory shows the navigation history stack in a menu dialog.
func (self *PreviewNavHelper) ShowNavHistory() error {
	pc := self.ctx()
	if len(pc.NavHistory) == 0 {
		return nil
	}

	var items []types.MenuItem
	for i := len(pc.NavHistory) - 1; i >= 0; i-- {
		entry := pc.NavHistory[i]
		label := strings.TrimSpace(entry.Title)
		if label == "" {
			label = "(untitled)"
		}
		if i == pc.NavIndex {
			label = "> " + label
		}
		idx := i
		items = append(items, types.MenuItem{
			Label: label,
			OnRun: func() error {
				pc.NavHistory[pc.NavIndex] = self.captureCurrentNavEntry()
				pc.NavIndex = idx
				self.restoreNavEntry(pc.NavHistory[idx])
				return nil
			},
		})
	}

	self.c.GuiCommon().ShowMenuDialog("Navigation History", items)
	return nil
}

// OpenNoteByUUID loads a note by UUID and displays it in the preview.
func (self *PreviewNavHelper) OpenNoteByUUID(uuid string) error {
	opts := self.c.GuiCommon().BuildSearchOptions()
	note, err := self.c.RuinCmd().Search.Get(uuid, opts)
	if err != nil || note == nil {
		return nil
	}
	self.PushNavHistory()
	self.c.Helpers().Preview().ShowCardList(" "+note.Title+" ", []models.Note{*note})
	self.c.GuiCommon().PushContextByKey("preview")
	return nil
}

// --- cursor movement ---

func (self *PreviewNavHelper) multiCardCount() int {
	pc := self.ctx()
	switch pc.Mode {
	case context.PreviewModeCardList:
		return len(pc.Cards)
	case context.PreviewModePickResults:
		return len(pc.PickResults)
	default:
		return 0
	}
}

func (self *PreviewNavHelper) isContentLine(lineNum int) bool {
	for _, r := range self.ctx().CardLineRanges {
		if lineNum > r[0] && lineNum < r[1]-1 {
			return true
		}
	}
	return false
}

// SyncCardIndexFromCursor updates SelectedCardIndex based on CursorLine.
func (self *PreviewNavHelper) SyncCardIndexFromCursor() {
	pc := self.ctx()
	ranges := pc.CardLineRanges
	cursor := pc.CursorLine
	for i, r := range ranges {
		if cursor >= r[0] && cursor < r[1] {
			pc.SelectedCardIndex = i
			return
		}
	}
	for i := 0; i < len(ranges)-1; i++ {
		if cursor >= ranges[i][1] && cursor < ranges[i+1][0] {
			pc.SelectedCardIndex = i + 1
			return
		}
	}
}

// MoveDown moves the cursor to the next content line.
func (self *PreviewNavHelper) MoveDown() error {
	pc := self.ctx()
	ranges := pc.CardLineRanges
	if len(ranges) > 0 {
		maxLine := ranges[len(ranges)-1][1] - 1
		cursor := pc.CursorLine
		for cursor < maxLine {
			cursor++
			if self.isContentLine(cursor) {
				break
			}
		}
		if self.isContentLine(cursor) && cursor != pc.CursorLine {
			pc.CursorLine = cursor
			self.SyncCardIndexFromCursor()
			self.c.GuiCommon().RenderPreview()
		}
	}
	return nil
}

// MoveUp moves the cursor to the previous content line.
func (self *PreviewNavHelper) MoveUp() error {
	pc := self.ctx()
	cursor := pc.CursorLine
	for cursor > 0 {
		cursor--
		if self.isContentLine(cursor) {
			break
		}
	}
	if self.isContentLine(cursor) && cursor != pc.CursorLine {
		pc.CursorLine = cursor
		self.SyncCardIndexFromCursor()
		self.c.GuiCommon().RenderPreview()
	}
	return nil
}

// CardDown jumps to the next card.
func (self *PreviewNavHelper) CardDown() error {
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	count := self.multiCardCount()
	next := idx + 1
	if next >= count {
		return nil
	}
	pc.SelectedCardIndex = next
	ranges := pc.CardLineRanges
	if next < len(ranges) {
		pc.CursorLine = ranges[next][0] + 1
	}
	self.c.GuiCommon().RenderPreview()
	return nil
}

// CardUp jumps to the previous card.
func (self *PreviewNavHelper) CardUp() error {
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	prev := idx - 1
	if prev < 0 {
		return nil
	}
	pc.SelectedCardIndex = prev
	ranges := pc.CardLineRanges
	if prev < len(ranges) {
		pc.CursorLine = ranges[prev][0] + 1
	}
	self.c.GuiCommon().RenderPreview()
	return nil
}

// NextHeader jumps to the next markdown header.
func (self *PreviewNavHelper) NextHeader() error {
	pc := self.ctx()
	cursor := pc.CursorLine
	for _, h := range pc.HeaderLines {
		if h > cursor {
			pc.CursorLine = h
			self.SyncCardIndexFromCursor()
			self.c.GuiCommon().RenderPreview()
			return nil
		}
	}
	return nil
}

// PrevHeader jumps to the previous markdown header.
func (self *PreviewNavHelper) PrevHeader() error {
	pc := self.ctx()
	cursor := pc.CursorLine
	for i := len(pc.HeaderLines) - 1; i >= 0; i-- {
		if pc.HeaderLines[i] < cursor {
			pc.CursorLine = pc.HeaderLines[i]
			self.SyncCardIndexFromCursor()
			self.c.GuiCommon().RenderPreview()
			return nil
		}
	}
	return nil
}

// --- scrolling ---

// ScrollDown scrolls the preview viewport down.
func (self *PreviewNavHelper) ScrollDown() error {
	gui := self.c.GuiCommon()
	if gui.CurrentContextKey() == "palette" {
		if v := gui.GetView("paletteList"); v != nil {
			ScrollViewport(v, 3)
		}
		return nil
	}
	v := self.view()
	if v == nil || v.Name() != "preview" {
		return nil
	}
	pc := self.ctx()
	pc.ScrollOffset += 3
	v.SetOrigin(0, pc.ScrollOffset)
	return nil
}

// ScrollUp scrolls the preview viewport up.
func (self *PreviewNavHelper) ScrollUp() error {
	gui := self.c.GuiCommon()
	if gui.CurrentContextKey() == "palette" {
		if v := gui.GetView("paletteList"); v != nil {
			ScrollViewport(v, -3)
		}
		return nil
	}
	v := self.view()
	if v == nil || v.Name() != "preview" {
		return nil
	}
	pc := self.ctx()
	pc.ScrollOffset -= 3
	if pc.ScrollOffset < 0 {
		pc.ScrollOffset = 0
	}
	v.SetOrigin(0, pc.ScrollOffset)
	return nil
}

// --- interaction ---

// Click handles a mouse click on the preview.
func (self *PreviewNavHelper) Click() error {
	v := self.view()
	if v == nil {
		return nil
	}
	cx, cy := v.Cursor()
	ox, oy := v.Origin()
	absX := cx + ox
	absY := cy + oy

	links := self.c.Helpers().PreviewLinks()
	links.ExtractLinks()
	pc := self.ctx()
	for _, link := range pc.Links {
		if link.Line == absY && absX >= link.Col && absX < link.Col+link.Len {
			return links.FollowLink(link)
		}
	}

	clickLine := absY
	for i, lr := range pc.CardLineRanges {
		if absY >= lr[0] && absY < lr[1] {
			pc.SelectedCardIndex = i
			if !self.isContentLine(clickLine) {
				clickLine = lr[0] + 1
			}
			break
		}
	}
	pc.CursorLine = clickLine

	self.c.GuiCommon().PushContextByKey("preview")
	self.c.GuiCommon().RenderPreview()
	return nil
}

// Back pops the preview context.
func (self *PreviewNavHelper) Back() error {
	self.c.GuiCommon().PopContext()
	return nil
}

// FocusNote focuses the notes panel on the currently previewed card.
func (self *PreviewNavHelper) FocusNote() error {
	pc := self.ctx()
	if len(pc.Cards) == 0 {
		return nil
	}
	card := pc.Cards[pc.SelectedCardIndex]
	gui := self.c.GuiCommon()
	notes := gui.Contexts().Notes
	for i, note := range notes.Items {
		if note.UUID == card.UUID {
			notes.SetSelectedLineIdx(i)
			gui.PushContextByKey("notes")
			gui.RenderNotes()
			return nil
		}
	}
	return nil
}

// OpenInEditor opens the currently selected card in $EDITOR.
func (self *PreviewNavHelper) OpenInEditor() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	return self.c.Helpers().Editor().OpenInEditor(card.Path)
}
