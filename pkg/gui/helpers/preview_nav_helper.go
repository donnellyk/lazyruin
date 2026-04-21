package helpers

import (
	"strings"

	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// PreviewNavHelper handles preview cursor movement, scrolling, and
// interaction (click, back, focus, editor). Preview navigation history
// (back/forward between committed views) lives in the Navigator helper.
type PreviewNavHelper struct {
	c *HelperCommon
}

// NewPreviewNavHelper creates a new PreviewNavHelper.
func NewPreviewNavHelper(c *HelperCommon) *PreviewNavHelper {
	return &PreviewNavHelper{c: c}
}

func (self *PreviewNavHelper) activeCtx() context.IPreviewContext {
	if self.c.GuiCommon().CurrentContextKey() == "pickDialog" {
		return self.c.GuiCommon().Contexts().PickDialog
	}
	return self.c.GuiCommon().Contexts().ActivePreview()
}

func (self *PreviewNavHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

func (self *PreviewNavHelper) renderActive() {
	gui := self.c.GuiCommon()
	if gui.CurrentContextKey() == "pickDialog" {
		gui.RenderPickDialog()
	} else {
		gui.RenderPreview()
	}
}

// ShowNavHistory opens a menu dialog listing the preview navigation
// history, newest entry first, with a `>` marker on the current entry.
// Selecting an entry jumps there via Navigator.JumpTo.
func (self *PreviewNavHelper) ShowNavHistory() error {
	nav := self.c.Helpers().Navigator()
	mgr := nav.Manager()
	entries := mgr.Entries()
	if len(entries) == 0 {
		return nil
	}
	current := mgr.Index()

	var items []types.MenuItem
	for i := len(entries) - 1; i >= 0; i-- {
		label := strings.TrimPrefix(strings.TrimSpace(entries[i].Title), "◌ ")
		if label == "" {
			label = "(untitled)"
		}
		if i == current {
			label = "> " + label
		}
		idx := i
		items = append(items, types.MenuItem{
			Label: label,
			OnRun: func() error { return nav.JumpTo(idx) },
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
	case "datePreview":
		return self.OpenDatePreviewResult()
	case "cardList":
		return self.FocusNote()
	case "compose":
		return self.OpenSourceAtCursor()
	default:
		return nil
	}
}

// OpenSourceAtCursor opens the source note attributed to the cursor line
// (via the compose source map) in card-list view, with the cursor
// pre-positioned on the matched source line. No-op if the cursor line has
// no resolvable source identity (separator, blank, or a dynamic-embed line
// the CLI did not attribute to a resolvable source).
func (self *PreviewNavHelper) OpenSourceAtCursor() error {
	target := self.c.Helpers().PreviewLineOps().ResolveTarget()
	if target == nil {
		return nil
	}
	opts := self.c.Helpers().Preview().BuildSearchOptions()
	note, err := self.c.RuinCmd().Search.Get(target.UUID, opts)
	if err != nil || note == nil {
		return nil
	}
	noteCopy := *note
	targetLine := target.LineNum
	title := displayTitleForNote(noteCopy.Title)
	return self.c.Helpers().Navigator().NavigateTo("cardList", title, func() error {
		source := self.c.Helpers().Preview().NewSingleNoteSource(noteCopy.UUID)
		self.c.Helpers().Preview().ShowCardList(title, []models.Note{noteCopy}, source)
		self.positionCursorAtContentLine(&noteCopy, targetLine)
		return nil
	})
}

// OpenDatePreviewResult opens the currently selected item in the date preview.
// For pick sections, it opens the note with cursor on the matched line.
// For note sections, it focuses the note in the notes panel.
func (self *PreviewNavHelper) OpenDatePreviewResult() error {
	gui := self.c.GuiCommon()
	dp := gui.Contexts().DatePreview
	idx := dp.SelectedCardIdx
	section := dp.SectionForCard(idx)

	switch section {
	case context.SectionTagPicks:
		localIdx := dp.LocalCardIdx(idx)
		if localIdx < len(dp.TagPicks) {
			dp.SetSelectedCardIndex(localIdx)
			err := self.openPickResultFrom(dp.TagPicks, dp, func() {
				dp.SetSelectedCardIndex(idx) // restore global index before nav history snapshot
			})
			return err
		}
	case context.SectionTodoPicks:
		localIdx := dp.LocalCardIdx(idx)
		if localIdx < len(dp.TodoPicks) {
			dp.SetSelectedCardIndex(localIdx)
			err := self.openPickResultFrom(dp.TodoPicks, dp, func() {
				dp.SetSelectedCardIndex(idx) // restore global index before nav history snapshot
			})
			return err
		}
	case context.SectionNotes:
		localIdx := dp.LocalCardIdx(idx)
		if localIdx < len(dp.Notes) {
			note := dp.Notes[localIdx]
			title := displayTitleForNote(note.Title)
			return self.c.Helpers().Navigator().NavigateTo("cardList", title, func() error {
				source := self.c.Helpers().Preview().NewSingleNoteSource(note.UUID)
				self.c.Helpers().Preview().ShowCardList(title, []models.Note{note}, source)
				return nil
			})
		}
	}
	return nil
}

// OpenNoteByUUID loads a note by UUID and displays it in the preview.
func (self *PreviewNavHelper) OpenNoteByUUID(uuid string) error {
	opts := self.c.Helpers().Preview().BuildSearchOptions()
	note, err := self.c.RuinCmd().Search.Get(uuid, opts)
	if err != nil || note == nil {
		return nil
	}
	noteCopy := *note
	return self.c.Helpers().Navigator().NavigateTo("cardList", noteCopy.Title, func() error {
		source := self.c.Helpers().Preview().NewSingleNoteSource(noteCopy.UUID)
		self.c.Helpers().Preview().ShowCardList(noteCopy.Title, []models.Note{noteCopy}, source)
		return nil
	})
}

// openPickResultFrom is the shared implementation for opening a pick result
// from any context that holds pick results (PickResults or PickDialog).
// It resolves the cursor target line, fetches the full note, optionally runs
// beforeNav (e.g. to close a dialog), then navigates to card-list view with
// cursor pre-positioned on the matched line.
func (self *PreviewNavHelper) openPickResultFrom(results []models.PickResult, ctx context.IPreviewContext, beforeNav func()) error {
	idx := ctx.SelectedCardIndex()
	if idx >= len(results) {
		return nil
	}
	result := results[idx]

	lineTarget := self.c.Helpers().PreviewLineOps().ResolveTarget()

	opts := self.c.Helpers().Preview().BuildSearchOptions()
	note, err := self.c.RuinCmd().Search.Get(result.UUID, opts)
	if err != nil || note == nil {
		return nil
	}
	noteCopy := *note
	title := displayTitleForNote(noteCopy.Title)

	if beforeNav != nil {
		beforeNav()
	}
	return self.c.Helpers().Navigator().NavigateTo("cardList", title, func() error {
		source := self.c.Helpers().Preview().NewSingleNoteSource(noteCopy.UUID)
		self.c.Helpers().Preview().ShowCardList(title, []models.Note{noteCopy}, source)
		if lineTarget != nil {
			self.positionCursorAtContentLine(&noteCopy, lineTarget.LineNum)
		}
		return nil
	})
}

// OpenPickResult opens the currently selected pick result as a full note in
// card-list view, with the cursor pre-positioned on the matched line.
func (self *PreviewNavHelper) OpenPickResult() error {
	pr := self.c.GuiCommon().Contexts().PickResults
	if len(pr.Results) == 0 {
		return nil
	}
	return self.openPickResultFrom(pr.Results, pr, nil)
}

// OpenPickDialogResult opens the selected pick dialog result as a full note,
// closing the dialog first. Cursor is pre-positioned on the matched line.
func (self *PreviewNavHelper) OpenPickDialogResult() error {
	pd := self.c.GuiCommon().Contexts().PickDialog
	if len(pd.Results) == 0 {
		return nil
	}
	return self.openPickResultFrom(pd.Results, pd, func() {
		self.c.Helpers().Pick().ClosePickDialog()
	})
}

// positionCursorAtContentLine repositions the card-list cursor to the visual
// line corresponding to the given 1-indexed content line number.
// Uses deterministic UUID + LineNum matching against BuildCardContent output.
func (self *PreviewNavHelper) positionCursorAtContentLine(note *models.Note, contentLineNum int) {
	v := self.view()
	if v == nil {
		return
	}
	gui := self.c.GuiCommon()
	contentWidth := types.PreviewContentWidth(v)
	cardLines := gui.BuildCardContent(*note, contentWidth)

	for i, sl := range cardLines {
		if sl.UUID == note.UUID && sl.LineNum == contentLineNum {
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
			self.renderActive()
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
		self.renderActive()
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
	self.renderActive()
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
	self.renderActive()
	return nil
}

func (self *PreviewNavHelper) allHeaders() []int {
	ns := self.activeCtx().NavState()
	headers := ns.HeaderLines
	gui := self.c.GuiCommon()
	if gui.Contexts().ActivePreviewKey == "datePreview" {
		dp := gui.Contexts().DatePreview
		headers = mergeSorted(headers, dp.SectionHeaderLines)
	}
	return headers
}

func mergeSorted(a, b []int) []int {
	result := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			result = append(result, a[i])
			i++
		} else {
			result = append(result, b[j])
			j++
		}
	}
	result = append(result, a[i:]...)
	result = append(result, b[j:]...)
	return result
}

// NextHeader jumps to the next markdown header (or section header in datePreview).
func (self *PreviewNavHelper) NextHeader() error {
	cursor := self.activeCtx().NavState().CursorLine
	for _, h := range self.allHeaders() {
		if h > cursor {
			self.activeCtx().NavState().CursorLine = h
			self.SyncCardIndexFromCursor()
			self.renderActive()
			return nil
		}
	}
	return nil
}

// PrevHeader jumps to the previous markdown header (or section header in datePreview).
func (self *PreviewNavHelper) PrevHeader() error {
	cursor := self.activeCtx().NavState().CursorLine
	headers := self.allHeaders()
	for i := len(headers) - 1; i >= 0; i-- {
		if headers[i] < cursor {
			self.activeCtx().NavState().CursorLine = headers[i]
			self.SyncCardIndexFromCursor()
			self.renderActive()
			return nil
		}
	}
	return nil
}

// NextSection jumps to the next section header (datePreview only).
func (self *PreviewNavHelper) NextSection() error {
	gui := self.c.GuiCommon()
	if gui.Contexts().ActivePreviewKey != "datePreview" {
		return nil
	}
	dp := gui.Contexts().DatePreview
	cursor := dp.NavState().CursorLine
	for _, h := range dp.SectionHeaderLines {
		if h > cursor {
			target := h + 2 // skip header + blank spacer
			if self.isContentLine(target) {
				dp.NavState().CursorLine = target
			} else {
				dp.NavState().CursorLine = h
			}
			self.SyncCardIndexFromCursor()
			self.renderActive()
			return nil
		}
	}
	return nil
}

// PrevSection jumps to the previous section header (datePreview only).
func (self *PreviewNavHelper) PrevSection() error {
	gui := self.c.GuiCommon()
	if gui.Contexts().ActivePreviewKey != "datePreview" {
		return nil
	}
	dp := gui.Contexts().DatePreview
	cursor := dp.NavState().CursorLine
	for i := len(dp.SectionHeaderLines) - 1; i >= 0; i-- {
		if dp.SectionHeaderLines[i] < cursor {
			target := dp.SectionHeaderLines[i] + 2
			if self.isContentLine(target) {
				dp.NavState().CursorLine = target
			} else {
				dp.NavState().CursorLine = dp.SectionHeaderLines[i]
			}
			self.SyncCardIndexFromCursor()
			self.renderActive()
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
		for _, seg := range link.Segments {
			if seg.Line == absY && absX >= seg.Col && absX < seg.Col+seg.Len {
				return links.FollowLink(link)
			}
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

	// Push the appropriate preview context based on ActivePreviewKey.
	// Focusing a hover preview via click commits it to navigation history,
	// matching the committed-focus semantics of Enter.
	gui := self.c.GuiCommon()
	self.c.Helpers().Navigator().CommitHover()
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
	if notes.SelectByUUID(card.UUID) {
		gui.PushContextByKey("notes")
		gui.RenderNotes()
	}
	return nil
}

// OpenInEditor opens the file under the cursor in $EDITOR. It resolves which
// file to edit from the cursor line's source identity, falling back to the
// current card's path. Works for cardList, compose, and datePreview contexts.
func (self *PreviewNavHelper) OpenInEditor() error {
	ctx := self.activeCtx()
	ns := ctx.NavState()
	path := self.resolvePathAtCursor(ns)
	if path == "" {
		if card := self.c.Helpers().Preview().CurrentPreviewCard(); card != nil {
			path = card.Path
		}
	}
	if path == "" {
		return nil
	}

	if err := self.c.Helpers().Editor().OpenFileInEditor(path); err != nil {
		return err
	}

	self.c.Helpers().Preview().ReloadActivePreview()
	self.c.GuiCommon().RenderAll()
	return nil
}

// resolvePathAtCursor returns the source file path for the line at the cursor.
// If the cursor is on a non-content line (separator, title bar), it scans
// backward to find the nearest content line with a path.
func (self *PreviewNavHelper) resolvePathAtCursor(ns *context.PreviewNavState) string {
	if ns.CursorLine >= 0 && ns.CursorLine < len(ns.Lines) {
		if p := ns.Lines[ns.CursorLine].Path; p != "" {
			return p
		}
	}
	for i := ns.CursorLine - 1; i >= 0; i-- {
		if i < len(ns.Lines) {
			if p := ns.Lines[i].Path; p != "" {
				return p
			}
		}
	}
	return ""
}
