package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// regex patterns for line operations
var (
	inlineTagRe  = regexp.MustCompile(`#[\w-]+`)
	inlineDateRe = regexp.MustCompile(`@\d{4}-\d{2}-\d{2}`)
)

// PreviewHelper encapsulates all preview operations: navigation history,
// content reloading, link handling, card mutations, display toggles, and
// line operations.
type PreviewHelper struct {
	c *HelperCommon
}

// NewPreviewHelper creates a new PreviewHelper.
func NewPreviewHelper(c *HelperCommon) *PreviewHelper {
	return &PreviewHelper{c: c}
}

// --- accessors ---

func (self *PreviewHelper) ctx() *context.PreviewContext {
	return self.c.GuiCommon().Contexts().Preview
}

func (self *PreviewHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// --- shared helpers (used by other helpers via gui_common adapters) ---

// CurrentPreviewCard returns the currently selected card, or nil if none.
func (self *PreviewHelper) CurrentPreviewCard() *models.Note {
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	if idx >= len(pc.Cards) {
		return nil
	}
	return &pc.Cards[idx]
}

// UpdatePreviewForNotes updates the preview pane to show the selected note.
func (self *PreviewHelper) UpdatePreviewForNotes() {
	gui := self.c.GuiCommon()
	notes := gui.Contexts().Notes
	if len(notes.Items) == 0 {
		return
	}
	idx := notes.GetSelectedLineIdx()
	if idx >= len(notes.Items) {
		return
	}
	note := notes.Items[idx]
	self.PushNavHistory()
	pc := self.ctx()
	pc.Mode = context.PreviewModeCardList
	pc.Cards = []models.Note{note}
	pc.SelectedCardIndex = 0
	pc.CursorLine = 1
	pc.ScrollOffset = 0
	v := self.view()
	if v != nil {
		v.Title = " " + note.Title + " "
		gui.RenderPreview()
	}
}

// UpdatePreviewCardList loads a card list into the preview.
func (self *PreviewHelper) UpdatePreviewCardList(title string, loadFn func() ([]models.Note, error)) {
	notes, err := loadFn()
	if err != nil {
		return
	}
	self.PushNavHistory()
	pc := self.ctx()
	pc.Mode = context.PreviewModeCardList
	pc.Cards = notes
	pc.SelectedCardIndex = 0
	pc.CursorLine = 1
	pc.ScrollOffset = 0
	v := self.view()
	if v != nil {
		v.Title = title
		self.c.GuiCommon().RenderPreview()
	}
}

// --- nav history ---

// PushNavHistory captures the current preview state onto the nav history stack.
func (self *PreviewHelper) PushNavHistory() {
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

func (self *PreviewHelper) captureCurrentNavEntry() context.NavEntry {
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

func (self *PreviewHelper) restoreNavEntry(entry context.NavEntry) {
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
func (self *PreviewHelper) NavBack() error {
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
func (self *PreviewHelper) NavForward() error {
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
func (self *PreviewHelper) ShowNavHistory() error {
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
func (self *PreviewHelper) OpenNoteByUUID(uuid string) error {
	opts := self.c.GuiCommon().BuildSearchOptions()
	note, err := self.c.RuinCmd().Search.Get(uuid, opts)
	if err != nil || note == nil {
		return nil
	}
	self.PushNavHistory()
	pc := self.ctx()
	pc.Mode = context.PreviewModeCardList
	pc.Cards = []models.Note{*note}
	pc.SelectedCardIndex = 0
	pc.CursorLine = 1
	pc.ScrollOffset = 0
	if v := self.view(); v != nil {
		v.Title = " " + note.Title + " "
	}
	self.c.GuiCommon().PushContextByKey("preview")
	self.c.GuiCommon().RenderPreview()
	return nil
}

// --- content reload ---

// ReloadContent reloads notes and preview cards with current toggle settings.
func (self *PreviewHelper) ReloadContent() {
	gui := self.c.GuiCommon()
	gui.RefreshNotes(true)

	pc := self.ctx()
	if len(pc.Cards) > 0 {
		savedCardIdx := pc.SelectedCardIndex
		self.reloadPreviewCards()
		if savedCardIdx < len(pc.Cards) {
			pc.SelectedCardIndex = savedCardIdx
		}
	}
	gui.RenderPreview()
}

func (self *PreviewHelper) reloadPreviewCards() {
	gui := self.c.GuiCommon()
	pc := self.ctx()
	pc.TemporarilyMoved = nil
	opts := gui.BuildSearchOptions()

	if gui.GetSearchQuery() != "" {
		notes, err := self.c.RuinCmd().Search.Search(gui.GetSearchQuery(), opts)
		if err == nil {
			pc.Cards = notes
		}
		gui.RenderPreview()
		return
	}

	switch gui.PreviousContextKey() {
	case "notes":
		self.reloadPreviewCardsFromNotes()
	case "tags":
		tagsCtx := gui.Contexts().Tags
		if len(tagsCtx.Items) > 0 {
			tag := tagsCtx.Items[tagsCtx.GetSelectedLineIdx()]
			notes, err := self.c.RuinCmd().Search.Search(tag.Name, opts)
			if err == nil {
				pc.Cards = notes
			}
		}
	case "queries":
		queriesCtx := gui.Contexts().Queries
		if queriesCtx.CurrentTab == "parents" {
			if len(queriesCtx.Parents) > 0 {
				parent := queriesCtx.Parents[queriesCtx.ParentsTrait().GetSelectedLineIdx()]
				composed, err := self.c.RuinCmd().Parent.ComposeFlat(parent.UUID, parent.Title)
				if err == nil {
					pc.Cards = []models.Note{composed}
				}
			}
		} else if len(queriesCtx.Queries) > 0 {
			query := queriesCtx.Queries[queriesCtx.QueriesTrait().GetSelectedLineIdx()]
			notes, err := self.c.RuinCmd().Queries.Run(query.Name, opts)
			if err == nil {
				pc.Cards = notes
			}
		}
	default:
		self.reloadPreviewCardsFromNotes()
	}

	gui.RenderPreview()
}

func (self *PreviewHelper) reloadPreviewCardsFromNotes() {
	pc := self.ctx()
	opts := self.c.GuiCommon().BuildSearchOptions()
	updated := make([]models.Note, 0, len(pc.Cards))
	for _, card := range pc.Cards {
		fresh, err := self.c.RuinCmd().Search.Get(card.UUID, opts)
		if err == nil && fresh != nil {
			if len(fresh.InlineTags) == 0 && len(card.InlineTags) > 0 {
				fresh.InlineTags = card.InlineTags
			}
			updated = append(updated, *fresh)
		} else {
			card.Content = ""
			updated = append(updated, card)
		}
	}
	pc.Cards = updated
}

// --- navigation ---

func (self *PreviewHelper) multiCardCount() int {
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

func (self *PreviewHelper) isContentLine(lineNum int) bool {
	for _, r := range self.ctx().CardLineRanges {
		if lineNum > r[0] && lineNum < r[1]-1 {
			return true
		}
	}
	return false
}

// SyncCardIndexFromCursor updates SelectedCardIndex based on CursorLine.
func (self *PreviewHelper) SyncCardIndexFromCursor() {
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

// MoveDown moves the cursor to the next content line (j / down).
func (self *PreviewHelper) MoveDown() error {
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

// MoveUp moves the cursor to the previous content line (k / up).
func (self *PreviewHelper) MoveUp() error {
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

// CardDown jumps to the next card (J).
func (self *PreviewHelper) CardDown() error {
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

// CardUp jumps to the previous card (K).
func (self *PreviewHelper) CardUp() error {
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

// NextHeader jumps to the next markdown header (}).
func (self *PreviewHelper) NextHeader() error {
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

// PrevHeader jumps to the previous markdown header ({).
func (self *PreviewHelper) PrevHeader() error {
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

// ScrollDown scrolls the preview viewport down.
func (self *PreviewHelper) ScrollDown() error {
	// If palette is active, scroll that instead
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
func (self *PreviewHelper) ScrollUp() error {
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

// Click handles a mouse click on the preview.
func (self *PreviewHelper) Click() error {
	v := self.view()
	if v == nil {
		return nil
	}
	cx, cy := v.Cursor()
	ox, oy := v.Origin()
	absX := cx + ox
	absY := cy + oy

	// Check if click lands on a link
	self.ExtractLinks()
	pc := self.ctx()
	for _, link := range pc.Links {
		if link.Line == absY && absX >= link.Col && absX < link.Col+link.Len {
			return self.FollowLink(link)
		}
	}

	// Snap click to nearest content line within the card
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
func (self *PreviewHelper) Back() error {
	self.c.GuiCommon().PopContext()
	return nil
}

// FocusNote focuses the notes panel on the currently previewed card.
func (self *PreviewHelper) FocusNote() error {
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
func (self *PreviewHelper) OpenInEditor() error {
	card := self.CurrentPreviewCard()
	if card == nil {
		return nil
	}
	return self.c.Helpers().Editor().OpenInEditor(card.Path)
}

// --- links ---

// ExtractLinks parses the preview content for wiki-links and URLs.
func (self *PreviewHelper) ExtractLinks() {
	pc := self.ctx()
	pc.Links = nil
	v := self.view()
	if v == nil {
		return
	}

	lines := v.ViewBufferLines()
	wikiRe := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	urlRe := regexp.MustCompile(`https?://[^\s)\]>]+`)

	for lineNum, line := range lines {
		plain := stripAnsi(line)
		for _, match := range wikiRe.FindAllStringIndex(plain, -1) {
			text := plain[match[0]:match[1]]
			pc.Links = append(pc.Links, context.PreviewLink{
				Text: text, Line: lineNum, Col: match[0], Len: match[1] - match[0],
			})
		}
		for _, match := range urlRe.FindAllStringIndex(plain, -1) {
			text := plain[match[0]:match[1]]
			pc.Links = append(pc.Links, context.PreviewLink{
				Text: text, Line: lineNum, Col: match[0], Len: match[1] - match[0],
			})
		}
	}
}

// HighlightNextLink cycles to the next link (l).
func (self *PreviewHelper) HighlightNextLink() error {
	self.ExtractLinks()
	pc := self.ctx()
	links := pc.Links
	if len(links) == 0 {
		return nil
	}
	cur := pc.RenderedLink
	next := cur + 1
	if next >= len(links) {
		next = 0
	}
	pc.HighlightedLink = next
	pc.CursorLine = links[next].Line
	self.SyncCardIndexFromCursor()
	self.c.GuiCommon().RenderPreview()
	return nil
}

// HighlightPrevLink cycles to the previous link (L).
func (self *PreviewHelper) HighlightPrevLink() error {
	self.ExtractLinks()
	pc := self.ctx()
	links := pc.Links
	if len(links) == 0 {
		return nil
	}
	cur := pc.RenderedLink
	prev := cur - 1
	if prev < 0 {
		prev = len(links) - 1
	}
	pc.HighlightedLink = prev
	pc.CursorLine = links[prev].Line
	self.SyncCardIndexFromCursor()
	self.c.GuiCommon().RenderPreview()
	return nil
}

// OpenLink opens the currently highlighted link.
func (self *PreviewHelper) OpenLink() error {
	pc := self.ctx()
	links := pc.Links
	hl := pc.RenderedLink
	if hl < 0 || hl >= len(links) {
		return nil
	}
	return self.FollowLink(links[hl])
}

// FollowLink navigates to a wiki-link target or opens a URL in the browser.
func (self *PreviewHelper) FollowLink(link context.PreviewLink) error {
	text := link.Text

	if strings.HasPrefix(text, "[[") && strings.HasSuffix(text, "]]") {
		target := text[2 : len(text)-2]
		if i := strings.Index(target, "#"); i >= 0 {
			target = target[:i]
		}
		if target == "" {
			return nil
		}
		opts := self.c.GuiCommon().BuildSearchOptions()
		note, err := self.c.RuinCmd().Search.GetByTitle(target, opts)
		if err != nil || note == nil {
			return nil
		}
		self.PushNavHistory()
		pc := self.ctx()
		pc.Mode = context.PreviewModeCardList
		pc.Cards = []models.Note{*note}
		pc.SelectedCardIndex = 0
		pc.CursorLine = 1
		pc.ScrollOffset = 0
		if v := self.view(); v != nil {
			v.Title = " " + note.Title + " "
		}
		self.c.GuiCommon().RenderPreview()
		return nil
	}

	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		exec.Command("open", text).Start()
	}
	return nil
}

// --- card mutations ---

// DeleteCard deletes the currently selected card.
func (self *PreviewHelper) DeleteCard() error {
	pc := self.ctx()
	if len(pc.Cards) == 0 {
		return nil
	}

	card := pc.Cards[pc.SelectedCardIndex]
	title := card.Title
	if title == "" {
		title = card.Path
	}
	if len(title) > 30 {
		title = title[:30] + "..."
	}

	gui := self.c.GuiCommon()
	gui.ShowConfirm("Delete Note", "Delete \""+title+"\"?", func() error {
		err := self.c.RuinCmd().Note.Delete(card.UUID)
		if err != nil {
			gui.ShowError(err)
			return nil
		}
		idx := pc.SelectedCardIndex
		pc.Cards = append(pc.Cards[:idx], pc.Cards[idx+1:]...)
		if pc.SelectedCardIndex >= len(pc.Cards) && pc.SelectedCardIndex > 0 {
			pc.SelectedCardIndex--
		}
		gui.RefreshNotes(false)
		gui.RenderPreview()
		return nil
	})
	return nil
}

// MoveCardDialog shows the move direction menu.
func (self *PreviewHelper) MoveCardDialog() error {
	pc := self.ctx()
	if len(pc.Cards) <= 1 {
		return nil
	}
	self.c.GuiCommon().ShowMenuDialog("Move", []types.MenuItem{
		{Label: "Move card up", Key: "u", OnRun: func() error { return self.moveCard("up") }},
		{Label: "Move card down", Key: "d", OnRun: func() error { return self.moveCard("down") }},
	})
	return nil
}

func (self *PreviewHelper) moveCard(direction string) error {
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	if direction == "up" {
		if idx <= 0 {
			return nil
		}
		pc.Cards[idx], pc.Cards[idx-1] = pc.Cards[idx-1], pc.Cards[idx]
		pc.SelectedCardIndex--
	} else {
		if idx >= len(pc.Cards)-1 {
			return nil
		}
		pc.Cards[idx], pc.Cards[idx+1] = pc.Cards[idx+1], pc.Cards[idx]
		pc.SelectedCardIndex++
	}

	if pc.TemporarilyMoved == nil {
		pc.TemporarilyMoved = make(map[int]bool)
	}
	pc.TemporarilyMoved[pc.SelectedCardIndex] = true

	gui := self.c.GuiCommon()
	gui.RenderPreview()
	newIdx := pc.SelectedCardIndex
	if newIdx < len(pc.CardLineRanges) {
		pc.CursorLine = pc.CardLineRanges[newIdx][0] + 1
	}
	gui.RenderPreview()
	return nil
}

// MergeCardDialog shows the merge direction menu.
func (self *PreviewHelper) MergeCardDialog() error {
	pc := self.ctx()
	if len(pc.Cards) <= 1 {
		return nil
	}
	self.c.GuiCommon().ShowMenuDialog("Merge", []types.MenuItem{
		{Label: "Merge card below into this one", Key: "d", OnRun: func() error { return self.executeMerge("down") }},
		{Label: "Merge card above into this one", Key: "u", OnRun: func() error { return self.executeMerge("up") }},
	})
	return nil
}

func (self *PreviewHelper) executeMerge(direction string) error {
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	var targetIdx, sourceIdx int
	if direction == "down" {
		if idx >= len(pc.Cards)-1 {
			return nil
		}
		targetIdx = idx
		sourceIdx = idx + 1
	} else {
		if idx <= 0 {
			return nil
		}
		targetIdx = idx
		sourceIdx = idx - 1
	}

	target := pc.Cards[targetIdx]
	source := pc.Cards[sourceIdx]

	result, err := self.c.RuinCmd().Note.Merge(target.UUID, source.UUID, true, false)
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	pc.Cards[targetIdx].Content = ""
	if len(result.TagsMerged) > 0 {
		pc.Cards[targetIdx].Tags = result.TagsMerged
	}

	pc.Cards = append(pc.Cards[:sourceIdx], pc.Cards[sourceIdx+1:]...)
	if pc.SelectedCardIndex >= len(pc.Cards) {
		pc.SelectedCardIndex = len(pc.Cards) - 1
	}
	if pc.SelectedCardIndex < 0 {
		pc.SelectedCardIndex = 0
	}

	gui := self.c.GuiCommon()
	gui.RefreshNotes(false)
	gui.RenderPreview()
	return nil
}

// OrderCards persists the current card order to frontmatter order fields.
func (self *PreviewHelper) OrderCards() error {
	pc := self.ctx()
	for i, card := range pc.Cards {
		if err := self.c.RuinCmd().Note.SetOrder(card.UUID, i+1); err != nil {
			self.c.GuiCommon().ShowError(err)
			return nil
		}
	}
	pc.TemporarilyMoved = nil
	self.ReloadContent()
	return nil
}

// --- display toggles ---

// ToggleMarkdown toggles markdown rendering.
func (self *PreviewHelper) ToggleMarkdown() error {
	self.ctx().RenderMarkdown = !self.ctx().RenderMarkdown
	self.c.GuiCommon().RenderPreview()
	return nil
}

// ToggleFrontmatter toggles frontmatter display.
func (self *PreviewHelper) ToggleFrontmatter() error {
	self.ctx().ShowFrontmatter = !self.ctx().ShowFrontmatter
	self.c.GuiCommon().RenderPreview()
	return nil
}

// ToggleTitle toggles title display.
func (self *PreviewHelper) ToggleTitle() error {
	self.ctx().ShowTitle = !self.ctx().ShowTitle
	self.ReloadContent()
	return nil
}

// ToggleGlobalTags toggles global tags display.
func (self *PreviewHelper) ToggleGlobalTags() error {
	self.ctx().ShowGlobalTags = !self.ctx().ShowGlobalTags
	self.ReloadContent()
	return nil
}

// ViewOptionsDialog shows the view options menu.
func (self *PreviewHelper) ViewOptionsDialog() error {
	pc := self.ctx()
	fmLabel := "Show frontmatter"
	if pc.ShowFrontmatter {
		fmLabel = "Hide frontmatter"
	}
	titleLabel := "Show title"
	if pc.ShowTitle {
		titleLabel = "Hide title"
	}
	tagsLabel := "Show global tags"
	if pc.ShowGlobalTags {
		tagsLabel = "Hide global tags"
	}
	mdLabel := "Render markdown"
	if pc.RenderMarkdown {
		mdLabel = "Raw markdown"
	}

	self.c.GuiCommon().ShowMenuDialog("View Options", []types.MenuItem{
		{Label: fmLabel, Key: "f", OnRun: func() error { return self.ToggleFrontmatter() }},
		{Label: titleLabel, Key: "t", OnRun: func() error { return self.ToggleTitle() }},
		{Label: tagsLabel, Key: "T", OnRun: func() error { return self.ToggleGlobalTags() }},
		{Label: mdLabel, Key: "M", OnRun: func() error { return self.ToggleMarkdown() }},
	})
	return nil
}

// --- line operations ---

// resolveSourceLine maps the current visual cursor position to a 1-indexed
// content line number in the raw source file (after frontmatter).
func (self *PreviewHelper) resolveSourceLine() int {
	v := self.view()
	if v == nil {
		return -1
	}
	card := self.CurrentPreviewCard()
	if card == nil {
		return -1
	}
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	ranges := pc.CardLineRanges
	if idx >= len(ranges) {
		return -1
	}

	cardStart := ranges[idx][0]
	lineOffset := pc.CursorLine - cardStart - 1
	if lineOffset < 0 {
		return -1
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)
	cardLines := self.c.GuiCommon().BuildCardContent(*card, contentWidth)
	if lineOffset >= len(cardLines) {
		return -1
	}
	visibleLine := strings.TrimSpace(stripAnsi(cardLines[lineOffset]))
	if visibleLine == "" {
		return -1
	}

	data, err := os.ReadFile(card.Path)
	if err != nil {
		return -1
	}
	fileLines := strings.Split(string(data), "\n")

	contentStart := 0
	if len(fileLines) > 0 && strings.HasPrefix(fileLines[0], "---") {
		for i := 1; i < len(fileLines); i++ {
			if strings.TrimSpace(fileLines[i]) == "---" {
				contentStart = i + 1
				break
			}
		}
	}

	for i := contentStart; i < len(fileLines); i++ {
		if strings.TrimSpace(fileLines[i]) == visibleLine {
			return i - contentStart + 1
		}
	}
	return -1
}

// readSourceLine reads the raw source file and returns the content line at
// the given 1-indexed content line number.
func readSourceLine(path string, lineNum int) (string, []string, int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, 0
	}
	fileLines := strings.Split(string(data), "\n")
	contentStart := 0
	if len(fileLines) > 0 && strings.HasPrefix(fileLines[0], "---") {
		for i := 1; i < len(fileLines); i++ {
			if strings.TrimSpace(fileLines[i]) == "---" {
				contentStart = i + 1
				break
			}
		}
	}
	absIdx := contentStart + lineNum - 1
	if absIdx < 0 || absIdx >= len(fileLines) {
		return "", fileLines, contentStart
	}
	return fileLines[absIdx], fileLines, contentStart
}

// ToggleTodo toggles a todo checkbox on the current line.
func (self *PreviewHelper) ToggleTodo() error {
	card := self.CurrentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := self.resolveSourceLine()
	if lineNum < 1 {
		return nil
	}

	err := self.c.RuinCmd().Note.ToggleTodo(card.UUID, lineNum)
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	self.ctx().Cards[self.ctx().SelectedCardIndex].Content = ""
	self.ReloadContent()
	return nil
}

// AppendDone toggles #done on the current line.
func (self *PreviewHelper) AppendDone() error {
	card := self.CurrentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := self.resolveSourceLine()
	if lineNum < 1 {
		return nil
	}

	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	hasDone := false
	for _, m := range inlineTagRe.FindAllString(srcLine, -1) {
		if strings.EqualFold(m, "#done") {
			hasDone = true
			break
		}
	}

	var err error
	if hasDone {
		err = self.c.RuinCmd().Note.RemoveTagFromLine(card.UUID, "#done", lineNum)
	} else {
		err = self.c.RuinCmd().Note.AddTagToLine(card.UUID, "#done", lineNum)
	}
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	self.ReloadContent()
	self.c.GuiCommon().RefreshTags(false)
	return nil
}

// ToggleInlineTag opens the input popup to toggle an inline tag on the cursor line.
func (self *PreviewHelper) ToggleInlineTag() error {
	card := self.CurrentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := self.resolveSourceLine()
	if lineNum < 1 {
		return nil
	}

	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	existingTags := make(map[string]bool)
	for _, m := range inlineTagRe.FindAllString(srcLine, -1) {
		existingTags[strings.ToLower(m)] = true
	}

	uuid := card.UUID
	gui := self.c.GuiCommon()
	gui.OpenInputPopup(&types.InputPopupConfig{
		Title:  "Toggle Inline Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{{Prefix: "#", Candidates: func(filter string) []types.CompletionItem {
				items := gui.TagCandidates(filter)
				var onLine, rest []types.CompletionItem
				for _, item := range items {
					if existingTags[strings.ToLower(item.Label)] {
						item.Detail = "*"
						onLine = append(onLine, item)
					} else {
						rest = append(rest, item)
					}
				}
				return append(onLine, rest...)
			}}}
		},
		OnAccept: func(_ string, item *types.CompletionItem) error {
			tag := ""
			if item != nil {
				tag = item.Label
			}
			if tag == "" {
				return nil
			}
			if !strings.HasPrefix(tag, "#") {
				tag = "#" + tag
			}

			if existingTags[strings.ToLower(tag)] {
				if err := self.c.RuinCmd().Note.RemoveTagFromLine(uuid, tag, lineNum); err != nil {
					gui.ShowError(err)
					return nil
				}
			} else {
				if err := self.c.RuinCmd().Note.AddTagToLine(uuid, tag, lineNum); err != nil {
					gui.ShowError(err)
					return nil
				}
			}
			self.ReloadContent()
			gui.RefreshTags(false)
			return nil
		},
	})
	return nil
}

// ToggleInlineDate opens the input popup to toggle an inline date on the cursor line.
func (self *PreviewHelper) ToggleInlineDate() error {
	card := self.CurrentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := self.resolveSourceLine()
	if lineNum < 1 {
		return nil
	}

	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	existingDates := make(map[string]bool)
	for _, m := range inlineDateRe.FindAllString(srcLine, -1) {
		existingDates[m] = true
	}

	uuid := card.UUID
	gui := self.c.GuiCommon()
	gui.OpenInputPopup(&types.InputPopupConfig{
		Title:  "Toggle Inline Date",
		Footer: " @ for dates | Tab: accept | Esc: cancel ",
		Seed:   "@",
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{{Prefix: "@", Candidates: func(filter string) []types.CompletionItem {
				items := gui.AtDateCandidates(filter)
				var onLine, rest []types.CompletionItem
				for _, item := range items {
					if existingDates[item.InsertText] {
						item.Detail = "*"
						onLine = append(onLine, item)
					} else {
						rest = append(rest, item)
					}
				}
				return append(onLine, rest...)
			}}}
		},
		OnAccept: func(_ string, item *types.CompletionItem) error {
			if item == nil || item.InsertText == "" {
				return nil
			}
			dateArg := strings.TrimPrefix(item.InsertText, "@")

			if existingDates[item.InsertText] {
				if err := self.c.RuinCmd().Note.RemoveDateFromLine(uuid, dateArg, lineNum); err != nil {
					gui.ShowError(err)
					return nil
				}
			} else {
				if err := self.c.RuinCmd().Note.AddDateToLine(uuid, dateArg, lineNum); err != nil {
					gui.ShowError(err)
					return nil
				}
			}
			self.ReloadContent()
			return nil
		},
	})
	return nil
}

// --- info dialog ---

// ShowInfoDialog shows parent structure / TOC for the current card.
func (self *PreviewHelper) ShowInfoDialog() error {
	card := self.CurrentPreviewCard()
	if card == nil {
		return nil
	}

	var items []types.MenuItem
	items = append(items, types.MenuItem{Label: "Info: " + card.Title, IsHeader: true})

	if card.Order != nil {
		items = append(items, types.MenuItem{Label: "Order: " + fmt.Sprintf("%d", *card.Order)})
	}

	treeRef := card.UUID
	if card.Parent != "" {
		treeRef = card.Parent
	}
	tree, err := self.c.RuinCmd().Parent.Tree(treeRef)
	if err == nil && (card.Parent != "" || len(tree.Children) > 0) {
		items = append(items, types.MenuItem{})
		items = append(items, types.MenuItem{Label: "Parent", IsHeader: true})
		rootUUID := tree.UUID
		items = append(items, types.MenuItem{Label: "* " + tree.Title, OnRun: func() error {
			return self.OpenNoteByUUID(rootUUID)
		}})
		items = self.appendTreeItems(items, tree.Children, "", 5)
	}

	headerItems := self.buildHeaderTOC()
	if len(headerItems) > 0 {
		items = append(items, types.MenuItem{})
		items = append(items, types.MenuItem{Label: "Headers", IsHeader: true})
		items = append(items, headerItems...)
	}

	self.c.GuiCommon().ShowMenuDialog("Info", items)
	return nil
}

func (self *PreviewHelper) appendTreeItems(items []types.MenuItem, children []commands.TreeNode, indent string, maxDepth int) []types.MenuItem {
	if maxDepth <= 0 || len(children) == 0 {
		return items
	}
	for _, child := range children {
		childUUID := child.UUID
		items = append(items, types.MenuItem{
			Label: indent + "  * " + child.Title,
			OnRun: func() error {
				return self.OpenNoteByUUID(childUUID)
			},
		})
		items = self.appendTreeItems(items, child.Children, indent+"  ", maxDepth-1)
	}
	return items
}

func (self *PreviewHelper) buildHeaderTOC() []types.MenuItem {
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	if idx >= len(pc.CardLineRanges) {
		return nil
	}
	ranges := pc.CardLineRanges[idx]

	var viewLines []string
	if v := self.view(); v != nil {
		viewLines = v.ViewBufferLines()
	}

	type header struct {
		level    int
		title    string
		viewLine int
	}
	var headers []header

	for _, hLine := range pc.HeaderLines {
		if hLine < ranges[0] || hLine >= ranges[1] {
			continue
		}
		if hLine < len(viewLines) {
			raw := strings.TrimSpace(stripAnsi(viewLines[hLine]))
			level := 0
			for _, r := range raw {
				if r == '#' {
					level++
				} else {
					break
				}
			}
			title := strings.TrimSpace(strings.TrimLeft(raw, "#"))
			headers = append(headers, header{level: level, title: title, viewLine: hLine})
		}
	}

	if len(headers) == 0 {
		return nil
	}

	minLevel := headers[0].level
	for _, h := range headers[1:] {
		if h.level < minLevel {
			minLevel = h.level
		}
	}

	var items []types.MenuItem
	for _, h := range headers {
		depth := h.level - minLevel
		indent := strings.Repeat("  ", depth)
		targetLine := h.viewLine
		items = append(items, types.MenuItem{
			Label: indent + "* " + h.title,
			OnRun: func() error {
				pc.CursorLine = targetLine
				self.c.GuiCommon().RenderPreview()
				return nil
			},
		})
	}
	return items
}

// --- utility functions ---

// stripAnsi removes ANSI escape sequences from a string.
func stripAnsi(s string) string {
	var sb strings.Builder
	inEsc := false
	for _, r := range s {
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == '\033' {
			inEsc = true
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}
