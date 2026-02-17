package gui

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// multiCardCount returns the number of items in the current preview.
func (gui *Gui) multiCardCount() int {
	switch gui.state.Preview.Mode {
	case PreviewModeCardList:
		return len(gui.state.Preview.Cards)
	case PreviewModePickResults:
		return len(gui.state.Preview.PickResults)
	default:
		return 0
	}
}

// isContentLine returns true if lineNum is a body/content line (not a separator or blank).
func (gui *Gui) isContentLine(lineNum int) bool {
	for _, r := range gui.state.Preview.CardLineRanges {
		if lineNum > r[0] && lineNum < r[1]-1 {
			return true
		}
	}
	return false
}

// syncCardIndexFromCursor updates SelectedCardIndex based on CursorLine position.
func (gui *Gui) syncCardIndexFromCursor() {
	ranges := gui.state.Preview.CardLineRanges
	cursor := gui.state.Preview.CursorLine
	for i, r := range ranges {
		if cursor >= r[0] && cursor < r[1] {
			gui.state.Preview.SelectedCardIndex = i
			return
		}
	}
	// On blank line between cards - attribute to next card
	for i := 0; i < len(ranges)-1; i++ {
		if cursor >= ranges[i][1] && cursor < ranges[i+1][0] {
			gui.state.Preview.SelectedCardIndex = i + 1
			return
		}
	}
}

func (gui *Gui) previewDown(g *gocui.Gui, v *gocui.View) error {
	ranges := gui.state.Preview.CardLineRanges
	if len(ranges) > 0 {
		maxLine := ranges[len(ranges)-1][1] - 1
		cursor := gui.state.Preview.CursorLine
		for cursor < maxLine {
			cursor++
			if gui.isContentLine(cursor) {
				break
			}
		}
		if gui.isContentLine(cursor) && cursor != gui.state.Preview.CursorLine {
			gui.state.Preview.CursorLine = cursor
			gui.syncCardIndexFromCursor()
			gui.renderPreview()
		}
	}
	return nil
}

func (gui *Gui) previewUp(g *gocui.Gui, v *gocui.View) error {
	cursor := gui.state.Preview.CursorLine
	for cursor > 0 {
		cursor--
		if gui.isContentLine(cursor) {
			break
		}
	}
	if gui.isContentLine(cursor) && cursor != gui.state.Preview.CursorLine {
		gui.state.Preview.CursorLine = cursor
		gui.syncCardIndexFromCursor()
		gui.renderPreview()
	}
	return nil
}

// previewCardDown jumps to the next card (J).
func (gui *Gui) previewCardDown(g *gocui.Gui, v *gocui.View) error {

	if listMove(&gui.state.Preview.SelectedCardIndex, gui.multiCardCount(), 1) {
		ranges := gui.state.Preview.CardLineRanges
		idx := gui.state.Preview.SelectedCardIndex
		if idx < len(ranges) {
			gui.state.Preview.CursorLine = ranges[idx][0] + 1 // first content line
		}
		gui.renderPreview()
	}
	return nil
}

// previewNextHeader jumps to the next markdown header (]).
func (gui *Gui) previewNextHeader(g *gocui.Gui, v *gocui.View) error {

	cursor := gui.state.Preview.CursorLine
	for _, h := range gui.state.Preview.HeaderLines {
		if h > cursor {
			gui.state.Preview.CursorLine = h
			gui.syncCardIndexFromCursor()
			gui.renderPreview()
			return nil
		}
	}
	return nil
}

// previewPrevHeader jumps to the previous markdown header ([).
func (gui *Gui) previewPrevHeader(g *gocui.Gui, v *gocui.View) error {

	cursor := gui.state.Preview.CursorLine
	for i := len(gui.state.Preview.HeaderLines) - 1; i >= 0; i-- {
		if gui.state.Preview.HeaderLines[i] < cursor {
			gui.state.Preview.CursorLine = gui.state.Preview.HeaderLines[i]
			gui.syncCardIndexFromCursor()
			gui.renderPreview()
			return nil
		}
	}
	return nil
}

// previewCardUp jumps to the previous card (K).
func (gui *Gui) previewCardUp(g *gocui.Gui, v *gocui.View) error {

	if listMove(&gui.state.Preview.SelectedCardIndex, gui.multiCardCount(), -1) {
		ranges := gui.state.Preview.CardLineRanges
		idx := gui.state.Preview.SelectedCardIndex
		if idx < len(ranges) {
			gui.state.Preview.CursorLine = ranges[idx][0] + 1 // first content line
		}
		gui.renderPreview()
	}
	return nil
}

// isTodoLine checks if a line is a markdown todo item.
func isTodoLine(line string) (isTodo bool, isComplete bool) {
	trimmed := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(trimmed, "- [ ] ") {
		return true, false
	}
	if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "- [X] ") {
		return true, true
	}
	return false, false
}

// findTodoGroup finds the contiguous block of todo lines containing the line at idx.
// Returns start (inclusive) and end (exclusive) indices.
func findTodoGroup(lines []string, idx int) (int, int) {
	start := idx
	for start > 0 {
		if ok, _ := isTodoLine(lines[start-1]); ok {
			start--
		} else {
			break
		}
	}
	end := idx + 1
	for end < len(lines) {
		if ok, _ := isTodoLine(lines[end]); ok {
			end++
		} else {
			break
		}
	}
	return start, end
}

// reorderTodoGroup moves a toggled todo within its group.
// Completing: move to bottom. Uncompleting: move just before first completed item.
// Returns the modified lines and the new absolute index of the toggled line.
func reorderTodoGroup(lines []string, groupStart, groupEnd, toggledIdx int, completing bool) ([]string, int) {
	// Extract the toggled line
	toggled := lines[toggledIdx]
	group := make([]string, 0, groupEnd-groupStart-1)
	for i := groupStart; i < groupEnd; i++ {
		if i != toggledIdx {
			group = append(group, lines[i])
		}
	}

	var newGroup []string
	var newRelIdx int
	if completing {
		// Append at end of group
		newGroup = append(group, toggled)
		newRelIdx = len(newGroup) - 1
	} else {
		// Insert just before first completed item
		insertAt := len(group)
		for i, l := range group {
			if _, complete := isTodoLine(l); complete {
				insertAt = i
				break
			}
		}
		newGroup = make([]string, 0, len(group)+1)
		newGroup = append(newGroup, group[:insertAt]...)
		newGroup = append(newGroup, toggled)
		newGroup = append(newGroup, group[insertAt:]...)
		newRelIdx = insertAt
	}

	result := make([]string, 0, len(lines))
	result = append(result, lines[:groupStart]...)
	result = append(result, newGroup...)
	result = append(result, lines[groupEnd:]...)
	return result, groupStart + newRelIdx
}

func (gui *Gui) toggleTodo(g *gocui.Gui, v *gocui.View) error {
	idx := gui.state.Preview.SelectedCardIndex
	ranges := gui.state.Preview.CardLineRanges
	if idx >= len(ranges) || idx >= len(gui.state.Preview.Cards) {
		return nil
	}

	// Get the rendered line at CursorLine
	note := gui.state.Preview.Cards[idx]
	cardStart := ranges[idx][0]
	lineOffset := gui.state.Preview.CursorLine - cardStart - 1 // -1 for upper separator

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)
	cardLines := gui.buildCardContent(note, contentWidth)
	if lineOffset < 0 || lineOffset >= len(cardLines) {
		return nil
	}

	visibleLine := strings.TrimSpace(stripAnsi(cardLines[lineOffset]))
	isTodo, wasComplete := isTodoLine(visibleLine)
	if !isTodo {
		return nil
	}

	// Read raw file
	data, err := os.ReadFile(note.Path)
	if err != nil {
		return nil
	}
	fileLines := strings.Split(string(data), "\n")

	// Find content start (after frontmatter)
	contentStart := 0
	if strings.HasPrefix(fileLines[0], "---") {
		for i := 1; i < len(fileLines); i++ {
			if strings.TrimSpace(fileLines[i]) == "---" {
				contentStart = i + 1
				break
			}
		}
	}
	// Skip leading blank lines after frontmatter (loadNoteContent does TrimLeft \n)
	for contentStart < len(fileLines) && strings.TrimSpace(fileLines[contentStart]) == "" {
		contentStart++
	}

	// Find matching source line by content
	matchIdx := -1
	for i := contentStart; i < len(fileLines); i++ {
		if strings.TrimSpace(fileLines[i]) == visibleLine {
			matchIdx = i
			break
		}
	}
	if matchIdx == -1 {
		return nil
	}

	// Toggle and reorder in the file (for persistence)
	toggleLine := func(line string, complete bool) string {
		if complete {
			line = strings.Replace(line, "- [x] ", "- [ ] ", 1)
			line = strings.Replace(line, "- [X] ", "- [ ] ", 1)
		} else {
			line = strings.Replace(line, "- [ ] ", "- [x] ", 1)
		}
		return line
	}

	fileLines[matchIdx] = toggleLine(fileLines[matchIdx], wasComplete)
	groupStart, groupEnd := findTodoGroup(fileLines, matchIdx)
	fileLines, newMatchIdx := reorderTodoGroup(fileLines, groupStart, groupEnd, matchIdx, !wasComplete)

	// Move cursor to follow the toggled line
	gui.state.Preview.CursorLine += newMatchIdx - matchIdx

	if err := os.WriteFile(note.Path, []byte(strings.Join(fileLines, "\n")), 0644); err != nil {
		return nil
	}

	// Apply the same toggle+reorder to the cached (already-stripped) content
	contentLines := strings.Split(note.Content, "\n")
	contentMatchIdx := -1
	for i, l := range contentLines {
		if strings.TrimSpace(l) == visibleLine {
			contentMatchIdx = i
			break
		}
	}
	if contentMatchIdx >= 0 {
		contentLines[contentMatchIdx] = toggleLine(contentLines[contentMatchIdx], wasComplete)
		cStart, cEnd := findTodoGroup(contentLines, contentMatchIdx)
		contentLines, _ = reorderTodoGroup(contentLines, cStart, cEnd, contentMatchIdx, !wasComplete)
		gui.state.Preview.Cards[idx].Content = strings.Join(contentLines, "\n")
	} else {
		gui.state.Preview.Cards[idx].Content = ""
	}

	gui.renderPreview()
	return nil
}

func (gui *Gui) previewScrollDown(g *gocui.Gui, v *gocui.View) error {
	if gui.state.PaletteMode {
		if gui.views.PaletteList != nil {
			scrollViewport(gui.views.PaletteList, 3)
		}
		return nil
	}
	if v == nil || v.Name() != PreviewView {
		return nil
	}
	gui.state.Preview.ScrollOffset += 3
	v.SetOrigin(0, gui.state.Preview.ScrollOffset)
	return nil
}

func (gui *Gui) previewScrollUp(g *gocui.Gui, v *gocui.View) error {
	if gui.state.PaletteMode {
		if gui.views.PaletteList != nil {
			scrollViewport(gui.views.PaletteList, -3)
		}
		return nil
	}
	if v == nil || v.Name() != PreviewView {
		return nil
	}
	gui.state.Preview.ScrollOffset -= 3
	if gui.state.Preview.ScrollOffset < 0 {
		gui.state.Preview.ScrollOffset = 0
	}
	v.SetOrigin(0, gui.state.Preview.ScrollOffset)
	return nil
}

func (gui *Gui) previewClick(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	ox, oy := v.Origin()
	absX := cx + ox
	absY := cy + oy

	// Check if click lands on a link
	gui.extractLinks()
	for _, link := range gui.state.Preview.Links {
		if link.Line == absY && absX >= link.Col && absX < link.Col+link.Len {
			return gui.followLink(link)
		}
	}

	// Snap click to nearest content line within the card
	clickLine := absY
	for i, lr := range gui.state.Preview.CardLineRanges {
		if absY >= lr[0] && absY < lr[1] {
			gui.state.Preview.SelectedCardIndex = i
			if !gui.isContentLine(clickLine) {
				clickLine = lr[0] + 1 // first content line
			}
			break
		}
	}
	gui.state.Preview.CursorLine = clickLine

	gui.setContext(PreviewContext)
	gui.renderPreview()
	return nil
}

func (gui *Gui) previewBack(g *gocui.Gui, v *gocui.View) error {
	gui.setContext(gui.state.PreviousContext)
	return nil
}

func (gui *Gui) focusNoteFromPreview(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Preview.Cards) == 0 {
		return nil
	}

	card := gui.state.Preview.Cards[gui.state.Preview.SelectedCardIndex]

	for i, note := range gui.state.Notes.Items {
		if note.UUID == card.UUID {
			gui.state.Notes.SelectedIndex = i
			gui.setContext(NotesContext)
			gui.renderNotes()
			return nil
		}
	}

	return nil
}

func (gui *Gui) toggleMarkdown(g *gocui.Gui, v *gocui.View) error {
	gui.state.Preview.RenderMarkdown = !gui.state.Preview.RenderMarkdown
	gui.renderPreview()
	return nil
}

func (gui *Gui) toggleFrontmatter(g *gocui.Gui, v *gocui.View) error {
	gui.state.Preview.ShowFrontmatter = !gui.state.Preview.ShowFrontmatter
	gui.renderPreview()
	return nil
}

func (gui *Gui) toggleTitle(g *gocui.Gui, v *gocui.View) error {
	gui.state.Preview.ShowTitle = !gui.state.Preview.ShowTitle
	gui.reloadContent()
	return nil
}

func (gui *Gui) toggleGlobalTags(g *gocui.Gui, v *gocui.View) error {
	gui.state.Preview.ShowGlobalTags = !gui.state.Preview.ShowGlobalTags
	gui.reloadContent()
	return nil
}

// reloadContent reloads notes from CLI with current toggle settings,
// preserving selection indices and preview mode.
func (gui *Gui) reloadContent() {
	// Reload notes for the Notes pane, preserving selection
	gui.fetchNotesForCurrentTab(true)

	// Reload cards in Preview pane
	if len(gui.state.Preview.Cards) > 0 {
		savedCardIdx := gui.state.Preview.SelectedCardIndex
		gui.reloadPreviewCards()
		if savedCardIdx < len(gui.state.Preview.Cards) {
			gui.state.Preview.SelectedCardIndex = savedCardIdx
		}
	}
	gui.renderPreview()
}

// reloadPreviewCards reloads the preview cards based on what generated them
func (gui *Gui) reloadPreviewCards() {
	gui.state.Preview.TemporarilyMoved = nil
	opts := gui.buildSearchOptions()

	// If there's an active search query, reload search results
	if gui.state.SearchQuery != "" {
		notes, err := gui.ruinCmd.Search.Search(gui.state.SearchQuery, opts)
		if err == nil {
			gui.state.Preview.Cards = notes
		}
		gui.renderPreview()
		return
	}

	// Otherwise, reload based on previous context
	switch gui.state.PreviousContext {
	case NotesContext:
		// The notes list was already refreshed by reloadContent().
		// Find the updated note(s) by UUID.
		gui.reloadPreviewCardsFromNotes()
	case TagsContext:
		if len(gui.state.Tags.Items) > 0 {
			tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]
			notes, err := gui.ruinCmd.Search.Search(tag.Name, opts)
			if err == nil {
				gui.state.Preview.Cards = notes
			}
		}
	case QueriesContext:
		if gui.state.Queries.CurrentTab == QueriesTabParents {
			if len(gui.state.Parents.Items) > 0 {
				parent := gui.state.Parents.Items[gui.state.Parents.SelectedIndex]
				composed, err := gui.ruinCmd.Parent.ComposeFlat(parent.UUID, parent.Title)
				if err == nil {
					gui.state.Preview.Cards = []models.Note{composed}
				}
			}
		} else if len(gui.state.Queries.Items) > 0 {
			query := gui.state.Queries.Items[gui.state.Queries.SelectedIndex]
			notes, err := gui.ruinCmd.Queries.Run(query.Name, opts)
			if err == nil {
				gui.state.Preview.Cards = notes
			}
		}
	default:
		gui.reloadPreviewCardsFromNotes()
	}

	gui.renderPreview()
}

// reloadPreviewCardsFromNotes re-fetches each preview card by UUID using
// the current search options (respecting title/tag toggle state).
func (gui *Gui) reloadPreviewCardsFromNotes() {
	opts := gui.buildSearchOptions()
	updated := make([]models.Note, 0, len(gui.state.Preview.Cards))
	for _, card := range gui.state.Preview.Cards {
		fresh, err := gui.ruinCmd.Search.Get(card.UUID, opts)
		if err == nil && fresh != nil {
			// ruin get doesn't return inline_tags; preserve from original
			if len(fresh.InlineTags) == 0 && len(card.InlineTags) > 0 {
				fresh.InlineTags = card.InlineTags
			}
			updated = append(updated, *fresh)
		} else {
			// Fallback: clear content so buildCardContent reads from disk
			card.Content = ""
			updated = append(updated, card)
		}
	}
	gui.state.Preview.Cards = updated
}

func (gui *Gui) updatePreviewForNotes() {
	if len(gui.state.Notes.Items) == 0 {
		return
	}
	idx := gui.state.Notes.SelectedIndex
	if idx >= len(gui.state.Notes.Items) {
		return
	}
	note := gui.state.Notes.Items[idx]
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = []models.Note{note}
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.CursorLine = 1
	gui.state.Preview.ScrollOffset = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Preview "
		gui.renderPreview()
	}
}

func (gui *Gui) deleteCardFromPreview(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Preview.Cards) == 0 {
		return nil
	}

	card := gui.state.Preview.Cards[gui.state.Preview.SelectedCardIndex]
	title := card.Title
	if title == "" {
		title = card.Path
	}
	if len(title) > 30 {
		title = title[:30] + "..."
	}

	gui.showConfirm("Delete Note", "Delete \""+title+"\"?", func() error {
		err := os.Remove(card.Path)
		if err != nil {
			gui.showError(err)
			return nil
		}
		idx := gui.state.Preview.SelectedCardIndex
		gui.state.Preview.Cards = append(gui.state.Preview.Cards[:idx], gui.state.Preview.Cards[idx+1:]...)
		if gui.state.Preview.SelectedCardIndex >= len(gui.state.Preview.Cards) && gui.state.Preview.SelectedCardIndex > 0 {
			gui.state.Preview.SelectedCardIndex--
		}
		gui.refreshNotes(false)
		gui.renderPreview()
		return nil
	})
	return nil
}

func (gui *Gui) moveCardHandler(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Preview.Cards) <= 1 {
		return nil
	}
	gui.showMoveOverlay()
	return nil
}

func (gui *Gui) showMoveOverlay() {
	gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "menu",
		Title:  "Move",
		MenuItems: []MenuItem{
			{Label: "Move card up", Key: "u", OnRun: func() error { return gui.moveCard("up") }},
			{Label: "Move card down", Key: "d", OnRun: func() error { return gui.moveCard("down") }},
		},
		MenuSelection: 0,
	}
}

func (gui *Gui) moveCard(direction string) error {
	idx := gui.state.Preview.SelectedCardIndex
	if direction == "up" {
		if idx <= 0 {
			return nil
		}
		gui.state.Preview.Cards[idx], gui.state.Preview.Cards[idx-1] = gui.state.Preview.Cards[idx-1], gui.state.Preview.Cards[idx]
		gui.state.Preview.SelectedCardIndex--
	} else {
		if idx >= len(gui.state.Preview.Cards)-1 {
			return nil
		}
		gui.state.Preview.Cards[idx], gui.state.Preview.Cards[idx+1] = gui.state.Preview.Cards[idx+1], gui.state.Preview.Cards[idx]
		gui.state.Preview.SelectedCardIndex++
	}

	if gui.state.Preview.TemporarilyMoved == nil {
		gui.state.Preview.TemporarilyMoved = make(map[int]bool)
	}
	gui.state.Preview.TemporarilyMoved[gui.state.Preview.SelectedCardIndex] = true

	// Render once to compute CardLineRanges for the new order,
	// then move cursor to the card's new position and render again.
	gui.renderPreview()
	newIdx := gui.state.Preview.SelectedCardIndex
	if newIdx < len(gui.state.Preview.CardLineRanges) {
		gui.state.Preview.CursorLine = gui.state.Preview.CardLineRanges[newIdx][0] + 1
	}
	gui.renderPreview()
	return nil
}

func (gui *Gui) mergeCardHandler(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Preview.Cards) <= 1 {
		return nil
	}
	gui.showMergeOverlay()
	return nil
}

func (gui *Gui) executeMerge(direction string) error {
	idx := gui.state.Preview.SelectedCardIndex
	var targetIdx, sourceIdx int
	if direction == "down" {
		if idx >= len(gui.state.Preview.Cards)-1 {
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

	target := gui.state.Preview.Cards[targetIdx]
	source := gui.state.Preview.Cards[sourceIdx]

	result, err := gui.ruinCmd.Note.Merge(target.UUID, source.UUID, true, false)
	if err != nil {
		gui.showError(err)
		return nil
	}

	// Clear target content so it re-reads from disk
	gui.state.Preview.Cards[targetIdx].Content = ""
	if len(result.TagsMerged) > 0 {
		gui.state.Preview.Cards[targetIdx].Tags = result.TagsMerged
	}

	// Remove source from cards
	gui.state.Preview.Cards = append(gui.state.Preview.Cards[:sourceIdx], gui.state.Preview.Cards[sourceIdx+1:]...)
	if gui.state.Preview.SelectedCardIndex >= len(gui.state.Preview.Cards) {
		gui.state.Preview.SelectedCardIndex = len(gui.state.Preview.Cards) - 1
	}
	if gui.state.Preview.SelectedCardIndex < 0 {
		gui.state.Preview.SelectedCardIndex = 0
	}

	gui.refreshNotes(false)
	gui.renderPreview()
	return nil
}

// currentPreviewCard returns the currently selected card, or nil if none.
func (gui *Gui) currentPreviewCard() *models.Note {
	idx := gui.state.Preview.SelectedCardIndex
	if idx >= len(gui.state.Preview.Cards) {
		return nil
	}
	return &gui.state.Preview.Cards[idx]
}

// resolveSourceLine maps the current visual cursor position to a 1-indexed
// content line number in the raw source file (after frontmatter). This accounts
// for title and global-tag lines that may be stripped from note.Content by
// search options. Returns -1 if the cursor is not on a matchable content line.
func (gui *Gui) resolveSourceLine(v *gocui.View) int {
	if v == nil {
		return -1
	}
	card := gui.currentPreviewCard()
	if card == nil {
		return -1
	}
	idx := gui.state.Preview.SelectedCardIndex
	ranges := gui.state.Preview.CardLineRanges
	if idx >= len(ranges) {
		return -1
	}

	// Get the visible text at the cursor
	cardStart := ranges[idx][0]
	lineOffset := gui.state.Preview.CursorLine - cardStart - 1 // -1 for separator
	if lineOffset < 0 {
		return -1
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)
	cardLines := gui.buildCardContent(*card, contentWidth)
	if lineOffset >= len(cardLines) {
		return -1
	}
	visibleLine := strings.TrimSpace(stripAnsi(cardLines[lineOffset]))
	if visibleLine == "" {
		return -1
	}

	// Read the raw file and find the content start (after frontmatter)
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
	// Do NOT skip leading blank lines here — the CLI's --line N counts
	// every line after the closing --- delimiter, including blanks.

	// Match the visible text against raw file lines
	for i := contentStart; i < len(fileLines); i++ {
		if strings.TrimSpace(fileLines[i]) == visibleLine {
			return i - contentStart + 1 // 1-indexed content line
		}
	}
	return -1
}

// openCardInEditor opens the currently selected card in $EDITOR.
func (gui *Gui) openCardInEditor(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	return gui.openInEditor(card.Path)
}

// appendDone appends " #done" to the current line via note append.
func (gui *Gui) appendDone(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := gui.resolveSourceLine(v)
	if lineNum < 1 {
		return nil
	}

	// Check if #done already exists on this line
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
		err = gui.ruinCmd.Note.RemoveTagFromLine(card.UUID, "#done", lineNum)
	} else {
		err = gui.ruinCmd.Note.AddTagToLine(card.UUID, "#done", lineNum)
	}
	if err != nil {
		gui.showError(err)
		return nil
	}

	gui.reloadContent()
	gui.refreshTags(false)
	return nil
}

// viewOptionsDialog shows the view options menu (displaced toggles).
func (gui *Gui) viewOptionsDialog(g *gocui.Gui, v *gocui.View) error {
	fmLabel := "Show frontmatter"
	if gui.state.Preview.ShowFrontmatter {
		fmLabel = "Hide frontmatter"
	}
	titleLabel := "Show title"
	if gui.state.Preview.ShowTitle {
		titleLabel = "Hide title"
	}
	tagsLabel := "Show global tags"
	if gui.state.Preview.ShowGlobalTags {
		tagsLabel = "Hide global tags"
	}
	mdLabel := "Render markdown"
	if gui.state.Preview.RenderMarkdown {
		mdLabel = "Raw markdown"
	}

	gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "menu",
		Title:  "View Options",
		MenuItems: []MenuItem{
			{Label: fmLabel, Key: "f", OnRun: func() error { return gui.toggleFrontmatter(nil, nil) }},
			{Label: titleLabel, Key: "t", OnRun: func() error { return gui.toggleTitle(nil, nil) }},
			{Label: tagsLabel, Key: "T", OnRun: func() error { return gui.toggleGlobalTags(nil, nil) }},
			{Label: mdLabel, Key: "M", OnRun: func() error { return gui.toggleMarkdown(nil, nil) }},
		},
		MenuSelection: 0,
	}
	return nil
}

// openInputPopup opens the generic input popup with the given config.
func (gui *Gui) openInputPopup(config *InputPopupConfig) {
	gui.state.InputPopupMode = true
	gui.state.InputPopupCompletion = NewCompletionState()
	gui.state.InputPopupSeedDone = false
	gui.state.InputPopupConfig = config
}

// inputPopupEnter handles Enter in the input popup.
func (gui *Gui) inputPopupEnter(g *gocui.Gui, v *gocui.View) error {
	state := gui.state.InputPopupCompletion
	config := gui.state.InputPopupConfig

	raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
	var item *CompletionItem
	if state.Active && len(state.Items) > 0 {
		selected := state.Items[state.SelectedIndex]
		item = &selected
	}

	gui.closeInputPopup()
	if (raw == "" && item == nil) || config == nil || config.OnAccept == nil {
		return nil
	}
	return config.OnAccept(raw, item)
}

// inputPopupTab accepts the current completion in the input popup.
func (gui *Gui) inputPopupTab(g *gocui.Gui, v *gocui.View) error {
	if gui.state.InputPopupCompletion.Active && len(gui.state.InputPopupCompletion.Items) > 0 {
		return gui.inputPopupEnter(g, v)
	}
	return nil
}

// inputPopupEsc cancels the input popup (first press dismisses suggestions).
func (gui *Gui) inputPopupEsc(g *gocui.Gui, v *gocui.View) error {
	if gui.state.InputPopupCompletion.Active {
		gui.state.InputPopupCompletion.Dismiss()
		return nil
	}
	gui.closeInputPopup()
	return nil
}

// closeInputPopup closes the input popup and restores focus.
func (gui *Gui) closeInputPopup() {
	gui.state.InputPopupMode = false
	gui.state.InputPopupCompletion = NewCompletionState()
	gui.state.InputPopupConfig = nil
	gui.g.Cursor = false
	gui.g.DeleteView(InputPopupView)
	gui.g.DeleteView(InputPopupSuggestView)
	gui.g.SetCurrentView(gui.contextToView(gui.state.CurrentContext))
}

// setParentDialog opens the input popup with > / >> parent completion.
func (gui *Gui) setParentDialog(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	uuid := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Set Parent",
		Footer: " > bookmarks | >> all notes | / drill | Tab: accept | Esc: cancel ",
		Seed:   ">",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{
				{Prefix: ">", Candidates: gui.parentCandidatesFor(gui.state.InputPopupCompletion)},
			}
		},
		OnAccept: func(raw string, item *CompletionItem) error {
			parentRef := strings.TrimLeft(raw, ">")
			if item != nil {
				parentRef = item.Value
				if parentRef == "" {
					parentRef = item.Label
				}
			}
			if parentRef == "" {
				return nil
			}
			err := gui.ruinCmd.Note.SetParent(uuid, parentRef)
			if err != nil {
				gui.showError(err)
				return nil
			}
			gui.reloadContent()
			return nil
		},
	})
	return nil
}

// removeParent removes the parent from the current card.
func (gui *Gui) removeParent(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	err := gui.ruinCmd.Note.RemoveParent(card.UUID)
	if err != nil {
		gui.showError(err)
		return nil
	}
	gui.reloadContent()
	return nil
}

// addGlobalTag opens the input popup to add a global tag.
func (gui *Gui) addGlobalTag(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	uuid := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Add Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{{Prefix: "#", Candidates: gui.tagCandidates}}
		},
		OnAccept: func(_ string, item *CompletionItem) error {
			tag := ""
			if item != nil {
				tag = item.Label
			}
			if tag == "" {
				return nil
			}
			err := gui.ruinCmd.Note.AddTag(uuid, tag)
			if err != nil {
				gui.showError(err)
				return nil
			}
			gui.reloadContent()
			gui.refreshTags(false)
			return nil
		},
	})
	return nil
}

// inlineTagRe matches #tag patterns (hashtag followed by word characters).
var inlineTagRe = regexp.MustCompile(`#[\w-]+`)

// readSourceLine reads the raw source file and returns the content line at
// the given 1-indexed content line number (after frontmatter). Returns ""
// and the file lines if the line is out of range.
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

// inlineDateRe matches @YYYY-MM-DD patterns on a content line.
var inlineDateRe = regexp.MustCompile(`@\d{4}-\d{2}-\d{2}`)

// toggleInlineTag opens the input popup to add or remove an inline tag on the cursor line.
// If the selected tag already appears on the line, it is removed; otherwise it is appended.
func (gui *Gui) toggleInlineTag(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := gui.resolveSourceLine(v)
	if lineNum < 1 {
		return nil
	}

	// Read the source line to detect existing inline tags
	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	existingTags := make(map[string]bool)
	for _, m := range inlineTagRe.FindAllString(srcLine, -1) {
		existingTags[strings.ToLower(m)] = true
	}

	uuid := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Toggle Inline Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{{Prefix: "#", Candidates: func(filter string) []CompletionItem {
				items := gui.tagCandidates(filter)
				var onLine, rest []CompletionItem
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
		OnAccept: func(_ string, item *CompletionItem) error {
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
				// Remove: strip the tag from the source line via CLI
				if err := gui.ruinCmd.Note.RemoveTagFromLine(uuid, tag, lineNum); err != nil {
					gui.showError(err)
					return nil
				}
			} else {
				// Add: append tag to end of line via CLI
				if err := gui.ruinCmd.Note.AddTagToLine(uuid, tag, lineNum); err != nil {
					gui.showError(err)
					return nil
				}
			}
			gui.reloadContent()
			gui.refreshTags(false)
			return nil
		},
	})
	return nil
}

// toggleInlineDate opens the input popup to add or remove an inline date on the cursor line.
// If the selected date already appears on the line, it is removed; otherwise it is appended.
func (gui *Gui) toggleInlineDate(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := gui.resolveSourceLine(v)
	if lineNum < 1 {
		return nil
	}

	// Read the source line to detect existing inline dates
	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	existingDates := make(map[string]bool)
	for _, m := range inlineDateRe.FindAllString(srcLine, -1) {
		existingDates[m] = true // keep the @YYYY-MM-DD form for matching
	}

	uuid := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Toggle Inline Date",
		Footer: " @ for dates | Tab: accept | Esc: cancel ",
		Seed:   "@",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{{Prefix: "@", Candidates: func(filter string) []CompletionItem {
				items := atDateCandidates(filter)
				var onLine, rest []CompletionItem
				for _, item := range items {
					// InsertText is e.g. "@today" or "@2026-02-17"; check resolved form
					// against existing dates on the line
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
		OnAccept: func(_ string, item *CompletionItem) error {
			if item == nil || item.InsertText == "" {
				return nil
			}
			// Strip the @ prefix — CLI expects "today" or "2026-02-17"
			dateArg := strings.TrimPrefix(item.InsertText, "@")

			if existingDates[item.InsertText] {
				if err := gui.ruinCmd.Note.RemoveDateFromLine(uuid, dateArg, lineNum); err != nil {
					gui.showError(err)
					return nil
				}
			} else {
				if err := gui.ruinCmd.Note.AddDateToLine(uuid, dateArg, lineNum); err != nil {
					gui.showError(err)
					return nil
				}
			}
			gui.reloadContent()
			return nil
		},
	})
	return nil
}

// removeTag opens the input popup showing only the current card's tags.
func (gui *Gui) removeTag(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	allTags := append(card.Tags, card.InlineTags...)
	if len(allTags) == 0 {
		return nil
	}
	uuid := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Remove Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{{Prefix: "#", Candidates: gui.currentCardTagCandidates}}
		},
		OnAccept: func(_ string, item *CompletionItem) error {
			tag := ""
			if item != nil {
				tag = item.Label
			}
			if tag == "" {
				return nil
			}
			err := gui.ruinCmd.Note.RemoveTag(uuid, tag)
			if err != nil {
				gui.showError(err)
				return nil
			}
			gui.reloadContent()
			gui.refreshTags(false)
			return nil
		},
	})
	return nil
}

// toggleBookmark toggles a parent bookmark for the current card.
func (gui *Gui) toggleBookmark(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	// Check if a bookmark already exists for this note
	bookmarks, err := gui.ruinCmd.Parent.List()
	if err == nil {
		for _, bm := range bookmarks {
			if bm.UUID == card.UUID {
				// Remove existing bookmark
				gui.ruinCmd.Parent.Delete(bm.Name)
				gui.refreshParents(false)
				return nil
			}
		}
	}
	// No bookmark exists — prompt for a name
	defaultName := card.Title
	if defaultName == "" {
		defaultName = card.UUID[:8]
	}
	cardUUID := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Save Bookmark",
		Footer: " Enter: save | Esc: cancel ",
		Seed:   defaultName,
		OnAccept: func(raw string, _ *CompletionItem) error {
			if raw == "" {
				return nil
			}
			err := gui.ruinCmd.Parent.Save(raw, cardUUID)
			if err != nil {
				gui.showError(err)
				return nil
			}
			gui.refreshParents(false)
			return nil
		},
	})
	return nil
}

// showInfoDialog shows parent structure / TOC for the current card.
func (gui *Gui) showInfoDialog(g *gocui.Gui, v *gocui.View) error {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}

	var items []MenuItem
	items = append(items, MenuItem{Label: "Info: " + card.Title, IsHeader: true})

	if card.Order != nil {
		items = append(items, MenuItem{Label: fmt.Sprintf("Order: %d", *card.Order)})
	}

	// Parent tree (git-log-graph style)
	// If the note has a parent, show the tree rooted at the parent.
	// Otherwise, show the tree rooted at the current note (if it has children).
	treeRef := card.UUID
	if card.Parent != "" {
		treeRef = card.Parent
	}
	tree, err := gui.ruinCmd.Parent.Tree(treeRef)
	if err == nil && (card.Parent != "" || len(tree.Children) > 0) {
		items = append(items, MenuItem{})
		items = append(items, MenuItem{Label: "Parent", IsHeader: true})
		items = append(items, MenuItem{Label: "* " + tree.Title})
		items = appendTreeItems(items, tree.Children, "", 5)
	}

	// TOC from headers (git-log-graph style with actual header titles)
	headerItems := gui.buildHeaderTOC()
	if len(headerItems) > 0 {
		items = append(items, MenuItem{})
		items = append(items, MenuItem{Label: "Headers", IsHeader: true})
		items = append(items, headerItems...)
	}

	gui.state.Dialog = &DialogState{
		Active:        true,
		Type:          "menu",
		Title:         "Info",
		MenuItems:     items,
		MenuSelection: 0,
	}
	return nil
}

// appendTreeItems recursively adds children to the menu in git-log-graph style.
func appendTreeItems(items []MenuItem, children []commands.TreeNode, prefix string, maxDepth int) []MenuItem {
	if maxDepth <= 0 || len(children) == 0 {
		return items
	}
	for i, child := range children {
		isLast := i == len(children)-1
		connector := "|"
		if isLast {
			connector = "\\"
		}
		items = append(items, MenuItem{Label: prefix + connector + " * " + child.Title})
		childPrefix := prefix
		if isLast {
			childPrefix += "  "
		} else {
			childPrefix += "| "
		}
		items = appendTreeItems(items, child.Children, childPrefix, maxDepth-1)
	}
	return items
}

// buildHeaderTOC returns menu items for the current card's headers in git-log-graph style.
func (gui *Gui) buildHeaderTOC() []MenuItem {
	idx := gui.state.Preview.SelectedCardIndex
	if idx >= len(gui.state.Preview.CardLineRanges) {
		return nil
	}
	ranges := gui.state.Preview.CardLineRanges[idx]

	var viewLines []string
	if gui.views.Preview != nil {
		viewLines = gui.views.Preview.ViewBufferLines()
	}

	type header struct {
		level int
		title string
	}
	var headers []header

	for _, hLine := range gui.state.Preview.HeaderLines {
		if hLine < ranges[0] || hLine >= ranges[1] {
			continue
		}
		// Extract actual header text from the view buffer
		title := ""
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
			title = strings.TrimSpace(strings.TrimLeft(raw, "#"))
			headers = append(headers, header{level: level, title: title})
		}
	}

	if len(headers) == 0 {
		return nil
	}

	// Find the minimum header level to use as the base
	minLevel := headers[0].level
	for _, h := range headers[1:] {
		if h.level < minLevel {
			minLevel = h.level
		}
	}

	var items []MenuItem
	for _, h := range headers {
		depth := h.level - minLevel
		prefix := strings.Repeat("| ", depth)
		items = append(items, MenuItem{Label: prefix + "* " + h.title})
	}
	return items
}

// orderCards persists the current card order to frontmatter order fields.
func (gui *Gui) orderCards() error {
	for i, card := range gui.state.Preview.Cards {
		if err := gui.ruinCmd.Note.SetOrder(card.UUID, i+1); err != nil {
			gui.showError(err)
			return nil
		}
	}
	gui.state.Preview.TemporarilyMoved = nil
	gui.reloadContent()
	return nil
}

// extractLinks parses the preview content for wiki-links ([[...]]) and URLs.
func (gui *Gui) extractLinks() {
	gui.state.Preview.Links = nil
	v := gui.views.Preview
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
			gui.state.Preview.Links = append(gui.state.Preview.Links, PreviewLink{
				Text: text,
				Line: lineNum,
				Col:  match[0],
				Len:  match[1] - match[0],
			})
		}
		for _, match := range urlRe.FindAllStringIndex(plain, -1) {
			text := plain[match[0]:match[1]]
			gui.state.Preview.Links = append(gui.state.Preview.Links, PreviewLink{
				Text: text,
				Line: lineNum,
				Col:  match[0],
				Len:  match[1] - match[0],
			})
		}
	}
}

// highlightNextLink cycles to the next link (l).
func (gui *Gui) highlightNextLink(g *gocui.Gui, v *gocui.View) error {
	gui.extractLinks()
	links := gui.state.Preview.Links
	if len(links) == 0 {
		return nil
	}
	cur := gui.state.Preview.renderedLink
	next := cur + 1
	if next >= len(links) {
		next = 0
	}
	gui.state.Preview.HighlightedLink = next
	gui.state.Preview.CursorLine = links[next].Line
	gui.syncCardIndexFromCursor()
	gui.renderPreview()
	return nil
}

// highlightPrevLink cycles to the previous link (L).
func (gui *Gui) highlightPrevLink(g *gocui.Gui, v *gocui.View) error {
	gui.extractLinks()
	links := gui.state.Preview.Links
	if len(links) == 0 {
		return nil
	}
	cur := gui.state.Preview.renderedLink
	prev := cur - 1
	if prev < 0 {
		prev = len(links) - 1
	}
	gui.state.Preview.HighlightedLink = prev
	gui.state.Preview.CursorLine = links[prev].Line
	gui.syncCardIndexFromCursor()
	gui.renderPreview()
	return nil
}

// openLink opens the currently highlighted link.
func (gui *Gui) openLink(g *gocui.Gui, v *gocui.View) error {
	links := gui.state.Preview.Links
	hl := gui.state.Preview.renderedLink
	if hl < 0 || hl >= len(links) {
		return nil
	}
	return gui.followLink(links[hl])
}

// followLink navigates to a wiki-link target or opens a URL in the browser.
func (gui *Gui) followLink(link PreviewLink) error {
	text := link.Text

	// Wiki-link: strip [[ and ]]
	if strings.HasPrefix(text, "[[") && strings.HasSuffix(text, "]]") {
		target := text[2 : len(text)-2]
		// Strip header fragment
		if i := strings.Index(target, "#"); i >= 0 {
			target = target[:i]
		}
		if target == "" {
			return nil
		}
		opts := gui.buildSearchOptions()
		note, err := gui.ruinCmd.Search.GetByTitle(target, opts)
		if err != nil || note == nil {
			return nil
		}
		gui.state.Preview.Mode = PreviewModeCardList
		gui.state.Preview.Cards = []models.Note{*note}
		gui.state.Preview.SelectedCardIndex = 0
		gui.state.Preview.CursorLine = 1
		gui.state.Preview.ScrollOffset = 0
		if gui.views.Preview != nil {
			gui.views.Preview.Title = " " + note.Title + " "
		}
		gui.renderPreview()
		return nil
	}

	// URL: open in browser
	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		exec.Command("open", text).Start()
	}
	return nil
}

// updatePreviewCardList is a shared helper for updating the preview with a card list.
func (gui *Gui) updatePreviewCardList(title string, loadFn func() ([]models.Note, error)) {
	notes, err := loadFn()
	if err != nil {
		return
	}
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.CursorLine = 1
	gui.state.Preview.ScrollOffset = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = title
		gui.renderPreview()
	}
}
