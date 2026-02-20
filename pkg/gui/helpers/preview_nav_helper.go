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

func (self *PreviewNavHelper) activeCtx() context.IPreviewContext {
	return self.c.GuiCommon().Contexts().ActivePreview()
}

func (self *PreviewNavHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// --- nav history ---

// PushNavHistory captures the current preview state onto the nav history stack.
func (self *PreviewNavHelper) PushNavHistory() {
	ctx := self.activeCtx()
	if ctx.CardCount() == 0 {
		return
	}

	entry := self.captureCurrentNavEntry()
	nh := ctx.NavHistory()

	// Truncate any forward entries
	if nh.Index >= 0 && nh.Index < len(nh.Entries)-1 {
		nh.Entries = nh.Entries[:nh.Index+1]
	}

	nh.Entries = append(nh.Entries, entry)
	nh.Index = len(nh.Entries) - 1

	// Cap at 50 entries
	if len(nh.Entries) > 50 {
		nh.Entries = nh.Entries[len(nh.Entries)-50:]
		nh.Index = len(nh.Entries) - 1
	}
}

func (self *PreviewNavHelper) captureCurrentNavEntry() context.NavEntry {
	contexts := self.c.GuiCommon().Contexts()
	title := ""
	if v := self.view(); v != nil {
		title = v.Title
	}
	ctx := self.activeCtx()
	ns := ctx.NavState()

	entry := context.NavEntry{
		SelectedCardIndex: ctx.SelectedCardIndex(),
		CursorLine:        ns.CursorLine,
		ScrollOffset:      ns.ScrollOffset,
		Title:             title,
		ContextKey:        contexts.ActivePreviewKey,
	}

	switch contexts.ActivePreviewKey {
	case "pickResults":
		pr := contexts.PickResults
		entry.PickResults = append([]models.PickResult(nil), pr.Results...)
	case "compose":
		comp := contexts.Compose
		entry.Cards = []models.Note{comp.Note}
	default:
		cl := contexts.CardList
		entry.Cards = append([]models.Note(nil), cl.Cards...)
	}

	return entry
}

func (self *PreviewNavHelper) restoreNavEntry(entry context.NavEntry) {
	contexts := self.c.GuiCommon().Contexts()
	gui := self.c.GuiCommon()

	targetKey := entry.ContextKey
	if targetKey == "" {
		targetKey = "cardList"
	}

	switch targetKey {
	case "pickResults":
		pr := contexts.PickResults
		pr.Results = append([]models.PickResult(nil), entry.PickResults...)
		pr.SelectedCardIdx = entry.SelectedCardIndex
		ns := pr.NavState()
		ns.CursorLine = entry.CursorLine
		ns.ScrollOffset = entry.ScrollOffset
	case "compose":
		comp := contexts.Compose
		if len(entry.Cards) > 0 {
			comp.Note = entry.Cards[0]
		}
		comp.SelectedCardIdx = entry.SelectedCardIndex
		ns := comp.NavState()
		ns.CursorLine = entry.CursorLine
		ns.ScrollOffset = entry.ScrollOffset
	default:
		cl := contexts.CardList
		cl.Cards = append([]models.Note(nil), entry.Cards...)
		cl.SelectedCardIdx = entry.SelectedCardIndex
		ns := cl.NavState()
		ns.CursorLine = entry.CursorLine
		ns.ScrollOffset = entry.ScrollOffset
	}

	contexts.ActivePreviewKey = targetKey

	if v := self.view(); v != nil {
		v.Title = entry.Title
		v.SetOrigin(0, entry.ScrollOffset)
	}

	// Switch to the correct preview context for the restored entry
	if context.IsPreviewContextKey(targetKey) {
		if gui.CurrentContextKey() != targetKey {
			gui.ReplaceContextByKey(targetKey)
		}
	}
	gui.RenderPreview()
}

// NavBack navigates backward in history.
func (self *PreviewNavHelper) NavBack() error {
	nh := self.activeCtx().NavHistory()
	if nh.Index < 0 || len(nh.Entries) == 0 {
		return nil
	}

	nh.Entries[nh.Index] = self.captureCurrentNavEntry()

	if nh.Index == len(nh.Entries)-1 {
		nh.Entries = append(nh.Entries, self.captureCurrentNavEntry())
	}

	if nh.Index <= 0 {
		return nil
	}

	nh.Index--
	self.restoreNavEntry(nh.Entries[nh.Index])
	return nil
}

// NavForward navigates forward in history.
func (self *PreviewNavHelper) NavForward() error {
	nh := self.activeCtx().NavHistory()
	if nh.Index >= len(nh.Entries)-1 {
		return nil
	}

	nh.Entries[nh.Index] = self.captureCurrentNavEntry()

	nh.Index++
	self.restoreNavEntry(nh.Entries[nh.Index])
	return nil
}

// ShowNavHistory shows the navigation history stack in a menu dialog.
func (self *PreviewNavHelper) ShowNavHistory() error {
	nh := self.activeCtx().NavHistory()
	if len(nh.Entries) == 0 {
		return nil
	}

	var items []types.MenuItem
	for i := len(nh.Entries) - 1; i >= 0; i-- {
		entry := nh.Entries[i]
		label := strings.TrimSpace(entry.Title)
		if label == "" {
			label = "(untitled)"
		}
		if i == nh.Index {
			label = "> " + label
		}
		idx := i
		items = append(items, types.MenuItem{
			Label: label,
			OnRun: func() error {
				nh.Entries[nh.Index] = self.captureCurrentNavEntry()
				nh.Index = idx
				self.restoreNavEntry(nh.Entries[idx])
				return nil
			},
		})
	}

	self.c.GuiCommon().ShowMenuDialog("Navigation History", items)
	return nil
}

// PreviewEnter dispatches Enter based on the active preview context.
func (self *PreviewNavHelper) PreviewEnter() error {
	switch self.c.GuiCommon().Contexts().ActivePreviewKey {
	case "pickResults":
		return self.OpenPickResult()
	case "cardList":
		return self.FocusNote()
	default:
		return nil
	}
}

// OpenNoteByUUID loads a note by UUID and displays it in the preview.
func (self *PreviewNavHelper) OpenNoteByUUID(uuid string) error {
	opts := self.c.Helpers().Preview().BuildSearchOptions()
	note, err := self.c.RuinCmd().Search.Get(uuid, opts)
	if err != nil || note == nil {
		return nil
	}
	self.PushNavHistory()
	self.c.Helpers().Preview().ShowCardList(" "+note.Title+" ", []models.Note{*note})
	self.c.GuiCommon().PushContextByKey("cardList")
	return nil
}

// OpenPickResult opens the currently selected pick result as a full note in
// card-list view, with the cursor pre-positioned on the matched line.
func (self *PreviewNavHelper) OpenPickResult() error {
	gui := self.c.GuiCommon()
	pr := gui.Contexts().PickResults
	if len(pr.Results) == 0 {
		return nil
	}
	idx := pr.SelectedCardIdx
	if idx >= len(pr.Results) {
		return nil
	}
	result := pr.Results[idx]

	// Resolve the pick target line (nil if cursor is on a separator)
	lineTarget := self.c.Helpers().PreviewLineOps().ResolvePickTarget()

	opts := self.c.Helpers().Preview().BuildSearchOptions()
	note, err := self.c.RuinCmd().Search.Get(result.UUID, opts)
	if err != nil || note == nil {
		return nil
	}

	self.PushNavHistory()
	self.c.Helpers().Preview().ShowCardList(" "+note.Title+" ", []models.Note{*note})
	gui.PushContextByKey("cardList")

	if lineTarget != nil {
		self.positionCursorAtContentLine(note, lineTarget.LineNum)
	}
	return nil
}

// positionCursorAtContentLine repositions the card-list cursor to the visual
// line corresponding to the given 1-indexed content line number.
func (self *PreviewNavHelper) positionCursorAtContentLine(note *models.Note, contentLineNum int) {
	gui := self.c.GuiCommon()
	v := self.view()
	if v == nil {
		return
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)

	cardLines := gui.BuildCardContent(*note, contentWidth)
	srcLine, _, _ := readSourceLine(note.Path, contentLineNum)
	srcLine = strings.TrimSpace(srcLine)
	if srcLine == "" {
		return
	}

	for i, line := range cardLines {
		if strings.TrimSpace(stripAnsi(line)) == srcLine {
			ns := gui.Contexts().CardList.NavState()
			ranges := ns.CardLineRanges
			if len(ranges) > 0 {
				ns.CursorLine = ranges[0][0] + 1 + i
				gui.RenderPreview()
			}
			return
		}
	}
}

// --- cursor movement ---

func (self *PreviewNavHelper) isContentLine(lineNum int) bool {
	ns := self.activeCtx().NavState()
	for _, r := range ns.CardLineRanges {
		if lineNum > r[0] && lineNum < r[1]-1 {
			return true
		}
	}
	return false
}

// SyncCardIndexFromCursor updates SelectedCardIndex based on CursorLine.
func (self *PreviewNavHelper) SyncCardIndexFromCursor() {
	ctx := self.activeCtx()
	ns := ctx.NavState()
	ranges := ns.CardLineRanges
	cursor := ns.CursorLine
	for i, r := range ranges {
		if cursor >= r[0] && cursor < r[1] {
			ctx.SetSelectedCardIndex(i)
			return
		}
	}
	for i := 0; i < len(ranges)-1; i++ {
		if cursor >= ranges[i][1] && cursor < ranges[i+1][0] {
			ctx.SetSelectedCardIndex(i + 1)
			return
		}
	}
}

// MoveDown moves the cursor to the next content line.
func (self *PreviewNavHelper) MoveDown() error {
	ctx := self.activeCtx()
	ns := ctx.NavState()
	ranges := ns.CardLineRanges
	if len(ranges) > 0 {
		maxLine := ranges[len(ranges)-1][1] - 1
		cursor := ns.CursorLine
		for cursor < maxLine {
			cursor++
			if self.isContentLine(cursor) {
				break
			}
		}
		if self.isContentLine(cursor) && cursor != ns.CursorLine {
			ns.CursorLine = cursor
			self.SyncCardIndexFromCursor()
			self.c.GuiCommon().RenderPreview()
		}
	}
	return nil
}

// MoveUp moves the cursor to the previous content line.
func (self *PreviewNavHelper) MoveUp() error {
	ns := self.activeCtx().NavState()
	cursor := ns.CursorLine
	for cursor > 0 {
		cursor--
		if self.isContentLine(cursor) {
			break
		}
	}
	if self.isContentLine(cursor) && cursor != ns.CursorLine {
		ns.CursorLine = cursor
		self.SyncCardIndexFromCursor()
		self.c.GuiCommon().RenderPreview()
	}
	return nil
}

// CardDown jumps to the next card.
func (self *PreviewNavHelper) CardDown() error {
	ctx := self.activeCtx()
	ns := ctx.NavState()
	idx := ctx.SelectedCardIndex()
	count := ctx.CardCount()
	next := idx + 1
	if next >= count {
		return nil
	}
	ctx.SetSelectedCardIndex(next)
	ranges := ns.CardLineRanges
	if next < len(ranges) {
		ns.CursorLine = ranges[next][0] + 1
	}
	self.c.GuiCommon().RenderPreview()
	return nil
}

// CardUp jumps to the previous card.
func (self *PreviewNavHelper) CardUp() error {
	ctx := self.activeCtx()
	ns := ctx.NavState()
	idx := ctx.SelectedCardIndex()
	prev := idx - 1
	if prev < 0 {
		return nil
	}
	ctx.SetSelectedCardIndex(prev)
	ranges := ns.CardLineRanges
	if prev < len(ranges) {
		ns.CursorLine = ranges[prev][0] + 1
	}
	self.c.GuiCommon().RenderPreview()
	return nil
}

// NextHeader jumps to the next markdown header.
func (self *PreviewNavHelper) NextHeader() error {
	ns := self.activeCtx().NavState()
	cursor := ns.CursorLine
	for _, h := range ns.HeaderLines {
		if h > cursor {
			ns.CursorLine = h
			self.SyncCardIndexFromCursor()
			self.c.GuiCommon().RenderPreview()
			return nil
		}
	}
	return nil
}

// PrevHeader jumps to the previous markdown header.
func (self *PreviewNavHelper) PrevHeader() error {
	ns := self.activeCtx().NavState()
	cursor := ns.CursorLine
	for i := len(ns.HeaderLines) - 1; i >= 0; i-- {
		if ns.HeaderLines[i] < cursor {
			ns.CursorLine = ns.HeaderLines[i]
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
	ns := self.activeCtx().NavState()
	ns.ScrollOffset += 3
	v.SetOrigin(0, ns.ScrollOffset)
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
	ns := self.activeCtx().NavState()
	ns.ScrollOffset -= 3
	if ns.ScrollOffset < 0 {
		ns.ScrollOffset = 0
	}
	v.SetOrigin(0, ns.ScrollOffset)
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
	ns := self.activeCtx().NavState()
	for _, link := range ns.Links {
		if link.Line == absY && absX >= link.Col && absX < link.Col+link.Len {
			return links.FollowLink(link)
		}
	}

	clickLine := absY
	for i, lr := range ns.CardLineRanges {
		if absY >= lr[0] && absY < lr[1] {
			self.activeCtx().SetSelectedCardIndex(i)
			if !self.isContentLine(clickLine) {
				clickLine = lr[0] + 1
			}
			break
		}
	}
	ns.CursorLine = clickLine

	// Push the appropriate preview context based on ActivePreviewKey
	gui := self.c.GuiCommon()
	gui.PushContextByKey(gui.Contexts().ActivePreviewKey)
	gui.RenderPreview()
	return nil
}

// Back pops the preview context.
func (self *PreviewNavHelper) Back() error {
	self.c.GuiCommon().PopContext()
	return nil
}

// FocusNote focuses the notes panel on the currently previewed card.
func (self *PreviewNavHelper) FocusNote() error {
	cl := self.c.GuiCommon().Contexts().CardList
	if len(cl.Cards) == 0 {
		return nil
	}
	card := cl.Cards[cl.SelectedCardIdx]
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
