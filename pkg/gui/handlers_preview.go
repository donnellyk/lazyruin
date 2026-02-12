package gui

import (
	"os"
	"strings"

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
	_, cy := v.Cursor()
	_, oy := v.Origin()
	absY := cy + oy

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
	savedNoteIdx := gui.state.Notes.SelectedIndex
	gui.loadNotesForCurrentTabPreserve()
	if savedNoteIdx < len(gui.state.Notes.Items) {
		gui.state.Notes.SelectedIndex = savedNoteIdx
	}
	gui.renderNotes()

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
			notes, err := gui.ruinCmd.Queries.Run(query.Name)
			if err == nil {
				gui.state.Preview.Cards = notes
			}
		}
	}

	gui.renderPreview()
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

	// Read both files' raw content (after stripping frontmatter)
	targetContent, err := gui.loadNoteContent(target.Path)
	if err != nil {
		gui.showError(err)
		return nil
	}
	sourceContent, err := gui.loadNoteContent(source.Path)
	if err != nil {
		gui.showError(err)
		return nil
	}

	// Merge tags (union)
	tagSet := make(map[string]bool)
	for _, t := range target.Tags {
		tagSet[t] = true
	}
	for _, t := range source.Tags {
		tagSet[t] = true
	}
	var mergedTags []string
	for t := range tagSet {
		mergedTags = append(mergedTags, t)
	}

	// Merge inline tags (union)
	inlineTagSet := make(map[string]bool)
	for _, t := range target.InlineTags {
		inlineTagSet[t] = true
	}
	for _, t := range source.InlineTags {
		inlineTagSet[t] = true
	}
	var mergedInlineTags []string
	for t := range inlineTagSet {
		mergedInlineTags = append(mergedInlineTags, t)
	}

	// Combine content
	combined := strings.TrimRight(targetContent, "\n") + "\n\n" + strings.TrimRight(sourceContent, "\n") + "\n"

	// Rewrite target file
	err = gui.writeNoteFile(target.Path, combined, mergedTags, mergedInlineTags)
	if err != nil {
		gui.showError(err)
		return nil
	}

	// Delete source file
	os.Remove(source.Path)

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

// writeNoteFile rewrites a note file preserving uuid/created/updated, with merged tags and new content.
func (gui *Gui) writeNoteFile(path, content string, tags, inlineTags []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Extract existing frontmatter fields
	raw := string(data)
	uuid := ""
	created := ""
	updated := ""
	title := ""

	if strings.HasPrefix(raw, "---") {
		rest := raw[3:]
		if idx := strings.Index(rest, "\n---"); idx != -1 {
			fmBlock := rest[:idx]
			for _, line := range strings.Split(fmBlock, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "uuid:") {
					uuid = strings.TrimSpace(strings.TrimPrefix(line, "uuid:"))
				} else if strings.HasPrefix(line, "created:") {
					created = strings.TrimSpace(strings.TrimPrefix(line, "created:"))
				} else if strings.HasPrefix(line, "updated:") {
					updated = strings.TrimSpace(strings.TrimPrefix(line, "updated:"))
				} else if strings.HasPrefix(line, "title:") {
					title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
				}
			}
		}
	}

	// Build new frontmatter
	var fm strings.Builder
	fm.WriteString("---\n")
	if uuid != "" {
		fm.WriteString("uuid: " + uuid + "\n")
	}
	if created != "" {
		fm.WriteString("created: " + created + "\n")
	}
	if updated != "" {
		fm.WriteString("updated: " + updated + "\n")
	}
	if title != "" {
		fm.WriteString("title: " + title + "\n")
	}
	if len(tags) > 0 {
		fm.WriteString("tags:\n")
		for _, t := range tags {
			fm.WriteString("  - " + t + "\n")
		}
	} else {
		fm.WriteString("tags: []\n")
	}
	if len(inlineTags) > 0 {
		fm.WriteString("inline-tags:\n")
		for _, t := range inlineTags {
			fm.WriteString("  - " + t + "\n")
		}
	}
	fm.WriteString("---\n")

	return os.WriteFile(path, []byte(fm.String()+content), 0644)
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
