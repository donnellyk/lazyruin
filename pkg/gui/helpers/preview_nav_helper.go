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
	ctx := self.activeCtx()
	ns := ctx.NavState()

	entry := context.NavEntry{
		SelectedCardIndex: ctx.SelectedCardIndex(),
		CursorLine:        ns.CursorLine,
		ScrollOffset:      ns.ScrollOffset,
		Title:             ctx.Title(),
		ContextKey:        contexts.ActivePreviewKey,
	}

	switch contexts.ActivePreviewKey {
	case "pickResults":
		pr := contexts.PickResults
		entry.PickResults = append([]models.PickResult(nil), pr.Results...)
	case "compose":
		comp := contexts.Compose
		entry.Cards = []models.Note{comp.Note}
		entry.SourceMap = append([]models.SourceMapEntry(nil), comp.SourceMap...)
		entry.ParentUUID = comp.ParentUUID
		entry.ParentTitle = comp.ParentTitle
	case "datePreview":
		dp := contexts.DatePreview
		entry.DateTargetDate = dp.TargetDate
		entry.DateTagPicks = append([]models.PickResult(nil), dp.TagPicks...)
		entry.DateTodoPicks = append([]models.PickResult(nil), dp.TodoPicks...)
		entry.DateNotes = append([]models.Note(nil), dp.Notes...)
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
		comp.SourceMap = append([]models.SourceMapEntry(nil), entry.SourceMap...)
		comp.ParentUUID = entry.ParentUUID
		comp.ParentTitle = entry.ParentTitle
		comp.SelectedCardIdx = entry.SelectedCardIndex
		ns := comp.NavState()
		ns.CursorLine = entry.CursorLine
		ns.ScrollOffset = entry.ScrollOffset
	case "datePreview":
		dp := contexts.DatePreview
		dp.TargetDate = entry.DateTargetDate
		dp.TagPicks = append([]models.PickResult(nil), entry.DateTagPicks...)
		dp.TodoPicks = append([]models.PickResult(nil), entry.DateTodoPicks...)
		dp.Notes = append([]models.Note(nil), entry.DateNotes...)
		dp.SelectedCardIdx = entry.SelectedCardIndex
		ns := dp.NavState()
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

	// Restore the title into the context state; layout reads it on next draw.
	activeCtx := gui.Contexts().ActivePreview()
	activeCtx.SetTitle(entry.Title)

	if v := self.view(); v != nil {
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
	case "datePreview":
		return self.OpenDatePreviewResult()
	case "cardList":
		return self.FocusNote()
	default:
		return nil
	}
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
			self.PushNavHistory()
			self.c.Helpers().Preview().ShowCardList(note.Title, []models.Note{note})
			gui.PushContextByKey("cardList")
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
	self.PushNavHistory()
	self.c.Helpers().Preview().ShowCardList(note.Title, []models.Note{*note})
	self.c.GuiCommon().PushContextByKey("cardList")
	return nil
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

	if beforeNav != nil {
		beforeNav()
	}
	self.PushNavHistory()
	self.c.Helpers().Preview().ShowCardList(note.Title, []models.Note{*note})
	self.c.GuiCommon().PushContextByKey("cardList")

	if lineTarget != nil {
		self.positionCursorAtContentLine(note, lineTarget.LineNum)
	}
	return nil
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

// OpenComposeInEditor opens the child file under the cursor in $EDITOR,
// runs doctor, reloads the compose preview, and preserves the cursor line.
func (self *PreviewNavHelper) OpenComposeInEditor() error {
	ns := self.c.GuiCommon().Contexts().Compose.NavState()
	path := self.resolveComposePath(ns)
	if path == "" {
		return nil
	}

	savedCursor := ns.CursorLine
	savedScroll := ns.ScrollOffset

	if err := self.c.Helpers().Editor().OpenFileInEditor(path); err != nil {
		return err
	}

	self.c.Helpers().Preview().ReloadActivePreview()
	ns.CursorLine = savedCursor
	ns.ScrollOffset = savedScroll
	self.c.GuiCommon().RenderAll()
	return nil
}

// resolveComposePath returns the file path for the child note at the cursor.
// If the cursor is on a non-content line (separator, title bar), it scans
// backward to find the nearest content line with a path.
func (self *PreviewNavHelper) resolveComposePath(ns *context.PreviewNavState) string {
	if ns.CursorLine >= 0 && ns.CursorLine < len(ns.Lines) {
		if p := ns.Lines[ns.CursorLine].Path; p != "" {
			return p
		}
	}
	// Scan backward for nearest line with a path.
	for i := ns.CursorLine - 1; i >= 0; i-- {
		if i < len(ns.Lines) {
			if p := ns.Lines[i].Path; p != "" {
				return p
			}
		}
	}
	return ""
}
