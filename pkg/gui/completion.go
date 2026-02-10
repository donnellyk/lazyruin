package gui

import (
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/commands"

	"github.com/jesseduffield/gocui"
)

// CompletionItem represents a single suggestion in the completion dropdown.
type CompletionItem struct {
	Label             string // display text (e.g. "#project")
	InsertText        string // text to insert (e.g. "#project")
	Detail            string // right-aligned detail (e.g. "(5)")
	ContinueCompleting bool   // if true, don't add trailing space -- allows chaining into next trigger
	Value              string // opaque data (e.g. UUID) for use by accept handlers
}

// CompletionTrigger defines a prefix that activates completion with a candidate provider.
type CompletionTrigger struct {
	Prefix     string
	Candidates func(filter string) []CompletionItem
}

// ParentDrillEntry records a parent selected during drill-down navigation.
type ParentDrillEntry struct {
	Name string
	UUID string
}

// CompletionState tracks the current state of a completion session.
type CompletionState struct {
	Active        bool
	TriggerStart  int // byte offset where the trigger token starts
	Items         []CompletionItem
	SelectedIndex int
	ParentDrill   []ParentDrillEntry // stack of drilled-into parents for > completion
}

// NewCompletionState returns an initialized CompletionState.
func NewCompletionState() *CompletionState {
	return &CompletionState{}
}

// cursorBytePos converts a TextArea's 2D cursor position (cx, cy) into a byte
// offset within the unwrapped content string.
func cursorBytePos(content string, cx, cy int) int {
	lines := strings.Split(content, "\n")
	pos := 0
	for i := 0; i < cy && i < len(lines); i++ {
		pos += len(lines[i]) + 1 // +1 for newline
	}
	if cy < len(lines) {
		lineLen := len(lines[cy])
		if cx > lineLen {
			cx = lineLen
		}
		pos += cx
	}
	if pos > len(content) {
		pos = len(content)
	}
	return pos
}

// extractTokenAtCursor scans backward from cursorPos to find the current token
// (delimited by whitespace or start of string). Returns the token and its start position.
func extractTokenAtCursor(content string, cursorPos int) (string, int) {
	if cursorPos > len(content) {
		cursorPos = len(content)
	}
	start := cursorPos
	for start > 0 {
		ch := content[start-1]
		if ch == ' ' || ch == '\t' || ch == '\n' {
			break
		}
		start--
	}
	return content[start:cursorPos], start
}

// detectTrigger checks if the token at the cursor matches any trigger prefix.
// Returns the matching trigger and the filter text (portion after the prefix), or nil.
// As a fallback, scans backward for unclosed [[ to support bracket-style triggers
// whose filter text may contain spaces.
func detectTrigger(content string, cursorPos int, triggers []CompletionTrigger) (*CompletionTrigger, string, int) {
	token, tokenStart := extractTokenAtCursor(content, cursorPos)
	for i := range triggers {
		t := &triggers[i]
		if strings.HasPrefix(token, t.Prefix) {
			filter := token[len(t.Prefix):]
			return t, filter, tokenStart
		}
	}

	// Fallback: scan for unclosed [[ before cursor (allows spaces in filter)
	cp := cursorPos
	if cp > len(content) {
		cp = len(content)
	}
	if idx := strings.LastIndex(content[:cp], "[["); idx >= 0 {
		after := content[idx+2 : cp]
		if !strings.Contains(after, "]]") {
			for i := range triggers {
				if triggers[i].Prefix == "[[" {
					return &triggers[i], after, idx
				}
			}
		}
	}

	// Fallback: scan backward on the current line for > at a word boundary
	// (parent completion whose filter may contain spaces after drilling)
	for i := cp - 1; i >= 0; i-- {
		ch := content[i]
		if ch == '\n' {
			break
		}
		if ch == '>' && (i == 0 || content[i-1] == ' ' || content[i-1] == '\t' || content[i-1] == '\n') {
			after := content[i+1 : cp]
			// Exclude blockquotes: "> " with a space and no / is a quote, not a parent
			if len(after) > 0 && after[0] == ' ' {
				break
			}
			for j := range triggers {
				if triggers[j].Prefix == ">" {
					return &triggers[j], after, i
				}
			}
			break
		}
	}

	return nil, "", 0
}

// triggerHints builds overview CompletionItems for each trigger prefix,
// shown when the input is empty or cursor is at whitespace.
func triggerHints(triggers []CompletionTrigger) []CompletionItem {
	descriptions := map[string]string{
		"#":        "filter by tag",
		"created:": "creation date",
		"updated:": "update date",
		"before:":  "created before",
		"after:":   "created after",
		"between:": "date range",
		"title:":   "search title",
		"path:":    "search path",
		"parent:":  "parent filter",
		"sort:":    "sort results",
	}
	var items []CompletionItem
	for _, t := range triggers {
		if t.Prefix == "/" {
			continue // don't include the / trigger itself in its own hints
		}
		detail := descriptions[t.Prefix]
		if detail == "" {
			detail = "filter"
		}
		items = append(items, CompletionItem{
			Label:              t.Prefix,
			InsertText:         t.Prefix,
			Detail:             detail,
			ContinueCompleting: true,
		})
	}
	return items
}

// updateCompletion is called after every keystroke. It checks whether a trigger
// is active and updates the CompletionState accordingly.
func (gui *Gui) updateCompletion(v *gocui.View, triggers []CompletionTrigger, state *CompletionState) {
	content := v.TextArea.GetUnwrappedContent()
	cx, cy := v.TextArea.GetCursorXY()
	cursorPos := cursorBytePos(content, cx, cy)

	trigger, filter, tokenStart := detectTrigger(content, cursorPos, triggers)
	if trigger != nil {
		items := trigger.Candidates(filter)
		if len(items) > 0 {
			state.Active = true
			state.TriggerStart = tokenStart
			state.Items = items
			if state.SelectedIndex >= len(items) {
				state.SelectedIndex = 0
			}
			return
		}
	}

	state.Active = false
	state.Items = nil
	state.SelectedIndex = 0
}

// acceptCompletion replaces the current trigger token with the selected item's InsertText.
// When triggers is non-nil and the item has ContinueCompleting set, completion is
// re-run immediately so the inserted prefix can chain into its own trigger.
func (gui *Gui) acceptCompletion(v *gocui.View, state *CompletionState, triggers []CompletionTrigger) {
	if !state.Active || len(state.Items) == 0 {
		return
	}

	item := state.Items[state.SelectedIndex]
	content := v.TextArea.GetUnwrappedContent()
	cx, cy := v.TextArea.GetCursorXY()
	cursorPos := cursorBytePos(content, cx, cy)

	// Calculate how many chars to backspace (from cursorPos back to TriggerStart)
	charsToDelete := cursorPos - state.TriggerStart
	for range charsToDelete {
		v.TextArea.BackSpaceChar()
	}

	if item.ContinueCompleting {
		v.TextArea.TypeString(item.InsertText)
	} else {
		v.TextArea.TypeString(item.InsertText + " ")
	}

	// Clear completion state
	state.Active = false
	state.Items = nil
	state.SelectedIndex = 0

	v.RenderTextArea()

	// Re-run completion so a chained trigger can activate immediately
	if item.ContinueCompleting && triggers != nil {
		gui.updateCompletion(v, triggers, state)
	}
}

// completionDown moves the selection down in the completion list.
func completionDown(state *CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}
	if state.SelectedIndex < len(state.Items)-1 {
		state.SelectedIndex++
	}
}

// completionUp moves the selection up in the completion list.
func completionUp(state *CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}
	if state.SelectedIndex > 0 {
		state.SelectedIndex--
	}
}

// tagCandidates returns tag completion items filtered by the given prefix.
func (gui *Gui) tagCandidates(filter string) []CompletionItem {
	filter = strings.ToLower(filter)
	var items []CompletionItem
	for _, tag := range gui.state.Tags.Items {
		name := tag.Name
		if !strings.HasPrefix(name, "#") {
			name = "#" + name
		}
		// Filter: match if filter is empty or tag name contains the filter
		nameWithoutHash := strings.TrimPrefix(name, "#")
		if filter != "" && !strings.Contains(strings.ToLower(nameWithoutHash), filter) {
			continue
		}
		items = append(items, CompletionItem{
			Label:      name,
			InsertText: name,
			Detail:     fmt.Sprintf("(%d)", tag.Count),
		})
	}
	return items
}

// dateShortcuts are the common date values used by created:, updated:, before:, after:.
var dateShortcuts = []struct {
	value  string
	detail string
}{
	{"today", "today"},
	{"yesterday", "yesterday"},
	{"this-week", "current week"},
	{"last-week", "previous week"},
	{"this-month", "current month"},
	{"last-month", "previous month"},
	{"this-year", "current year"},
	{"last-year", "previous year"},
	{"1d", "1 day"},
	{"7d", "1 week"},
	{"2w", "2 weeks"},
	{"30d", "1 month"},
	{"90d", "3 months"},
	{"365d", "1 year"},
}

// dateCandidates builds completion items for a date-prefix filter (e.g. "created:", "updated:").
func dateCandidates(prefix, filter string) []CompletionItem {
	filter = strings.ToLower(filter)
	var items []CompletionItem
	for _, s := range dateShortcuts {
		if filter != "" &&
			!strings.Contains(s.value, filter) &&
			!strings.Contains(s.detail, filter) {
			continue
		}
		items = append(items, CompletionItem{
			Label:      prefix + s.value,
			InsertText: prefix + s.value,
			Detail:     s.detail,
		})
	}
	return items
}

func (gui *Gui) createdCandidates(filter string) []CompletionItem {
	return dateCandidates("created:", filter)
}

func (gui *Gui) updatedCandidates(filter string) []CompletionItem {
	return dateCandidates("updated:", filter)
}

func (gui *Gui) beforeCandidates(filter string) []CompletionItem {
	return dateCandidates("before:", filter)
}

func (gui *Gui) afterCandidates(filter string) []CompletionItem {
	return dateCandidates("after:", filter)
}

// betweenCandidates returns between: filter suggestions.
func (gui *Gui) betweenCandidates(filter string) []CompletionItem {
	shortcuts := []CompletionItem{
		{Label: "between:last-week,today", InsertText: "between:last-week,today", Detail: "last week to now"},
		{Label: "between:last-month,today", InsertText: "between:last-month,today", Detail: "last month to now"},
		{Label: "between:last-year,today", InsertText: "between:last-year,today", Detail: "last year to now"},
	}

	if filter == "" {
		return shortcuts
	}

	filter = strings.ToLower(filter)
	var items []CompletionItem
	for _, s := range shortcuts {
		suffix := strings.TrimPrefix(s.InsertText, "between:")
		if strings.Contains(suffix, filter) || strings.Contains(s.Detail, filter) {
			items = append(items, s)
		}
	}
	return items
}

// titleCandidates returns note titles as completion items.
func (gui *Gui) titleCandidates(filter string) []CompletionItem {
	filter = strings.ToLower(filter)
	seen := make(map[string]bool)
	var items []CompletionItem
	for _, note := range gui.state.Notes.Items {
		title := note.Title
		if title == "" || seen[title] {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(title), filter) {
			continue
		}
		seen[title] = true
		items = append(items, CompletionItem{
			Label:      "title:" + title,
			InsertText: "title:" + title,
			Detail:     note.ShortDate(),
		})
	}
	return items
}

// wikiLinkCandidates returns note titles for [[ wiki-style link completion.
// When the filter contains '#', it switches to header mode for the specified note.
func (gui *Gui) wikiLinkCandidates(filter string) []CompletionItem {
	// Header mode: filter contains '#'
	if noteTitle, after, ok := strings.Cut(filter, "#"); ok {
		headerFilter := strings.ToLower(after)
		return gui.headerCandidates(noteTitle, headerFilter)
	}

	filterLower := strings.ToLower(filter)
	seen := make(map[string]bool)
	var items []CompletionItem
	for _, note := range gui.state.Notes.Items {
		title := note.Title
		if title == "" || seen[title] {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(title), filterLower) {
			continue
		}
		seen[title] = true
		items = append(items, CompletionItem{
			Label:      title,
			InsertText: "[[" + title + "]]",
			Detail:     note.ShortDate(),
		})
	}
	return items
}

// headerInfo represents a markdown heading extracted from note content.
type headerInfo struct {
	Level int    // 1-6
	Text  string // heading text without # prefix
}

// extractHeaders parses markdown content and returns all headings.
// Skips headings inside fenced code blocks.
func extractHeaders(content string) []headerInfo {
	var headers []headerInfo
	inCodeBlock := false
	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Count heading level
		level := 0
		for _, ch := range trimmed {
			if ch == '#' {
				level++
			} else {
				break
			}
		}
		if level < 1 || level > 6 || level >= len(trimmed) {
			continue
		}
		// Markdown headings require a space after the # prefix
		if trimmed[level] != ' ' {
			continue
		}
		text := strings.TrimSpace(trimmed[level:])
		if text == "" {
			continue
		}
		headers = append(headers, headerInfo{Level: level, Text: text})
	}
	return headers
}

// headerCandidates returns completion items for headers within a specific note.
func (gui *Gui) headerCandidates(noteTitle, filter string) []CompletionItem {
	// Find the note by exact title match
	var content string
	for i, note := range gui.state.Notes.Items {
		if note.Title == noteTitle {
			if note.Content == "" {
				loaded, err := gui.loadNoteContent(note.Path)
				if err != nil {
					return nil
				}
				gui.state.Notes.Items[i].Content = loaded
				content = loaded
			} else {
				content = note.Content
			}
			break
		}
	}
	if content == "" {
		return nil
	}

	headers := extractHeaders(content)
	var items []CompletionItem
	for _, h := range headers {
		if filter != "" && !strings.Contains(strings.ToLower(h.Text), filter) {
			continue
		}
		items = append(items, CompletionItem{
			Label:      h.Text,
			InsertText: "[[" + noteTitle + "#" + h.Text + "]]",
			Detail:     fmt.Sprintf("h%d", h.Level),
		})
	}
	return items
}

// pathCandidates returns path: filter hint.
func (gui *Gui) pathCandidates(filter string) []CompletionItem {
	if filter != "" {
		return nil
	}
	return []CompletionItem{
		{Label: "path:", InsertText: "path:", Detail: "search by path"},
	}
}

// parentCandidates returns parent: filter suggestions.
func (gui *Gui) parentCandidates(filter string) []CompletionItem {
	shortcuts := []CompletionItem{
		{Label: "parent:none", InsertText: "parent:none", Detail: "root notes only"},
	}
	// Add known parent bookmarks
	for _, p := range gui.state.Parents.Items {
		item := CompletionItem{
			Label:      "parent:" + p.Name,
			InsertText: "parent:" + p.UUID,
			Detail:     p.Title,
		}
		shortcuts = append(shortcuts, item)
	}

	if filter == "" {
		return shortcuts
	}

	filter = strings.ToLower(filter)
	var items []CompletionItem
	for _, s := range shortcuts {
		suffix := strings.TrimPrefix(s.Label, "parent:")
		if strings.Contains(strings.ToLower(suffix), filter) ||
			strings.Contains(strings.ToLower(s.Detail), filter) {
			items = append(items, s)
		}
	}
	return items
}

// sortCandidates returns sort: completion items for the search popup.
func sortCandidates(filter string) []CompletionItem {
	items := []CompletionItem{
		{Label: "sort:created:desc", InsertText: "sort:created:desc", Detail: "newest first"},
		{Label: "sort:created:asc", InsertText: "sort:created:asc", Detail: "oldest first"},
		{Label: "sort:updated:desc", InsertText: "sort:updated:desc", Detail: "recently updated"},
		{Label: "sort:updated:asc", InsertText: "sort:updated:asc", Detail: "least updated"},
		{Label: "sort:title:asc", InsertText: "sort:title:asc", Detail: "A-Z"},
		{Label: "sort:title:desc", InsertText: "sort:title:desc", Detail: "Z-A"},
		{Label: "sort:order:asc", InsertText: "sort:order:asc", Detail: "manual order"},
		{Label: "sort:order:desc", InsertText: "sort:order:desc", Detail: "manual reverse"},
	}

	if filter == "" {
		return items
	}

	filter = strings.ToLower(filter)
	var filtered []CompletionItem
	for _, item := range items {
		suffix := strings.TrimPrefix(item.InsertText, "sort:")
		if strings.Contains(suffix, filter) || strings.Contains(strings.ToLower(item.Detail), filter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// extractSort removes any "sort:field:dir" token from the query, returning
// the cleaned query and the sort value (e.g. "created:desc") for the -s flag.
func extractSort(query string) (string, string) {
	var remaining []string
	var sortVal string
	for _, token := range strings.Fields(query) {
		if v, ok := strings.CutPrefix(token, "sort:"); ok {
			sortVal = v
		} else {
			remaining = append(remaining, token)
		}
	}
	return strings.Join(remaining, " "), sortVal
}

// searchTriggers returns the completion triggers for the search popup.
// The "/" trigger shows an overview of all available filter prefixes.
func (gui *Gui) searchTriggers() []CompletionTrigger {
	triggers := []CompletionTrigger{
		{Prefix: "#", Candidates: gui.tagCandidates},
		{Prefix: "created:", Candidates: gui.createdCandidates},
		{Prefix: "updated:", Candidates: gui.updatedCandidates},
		{Prefix: "before:", Candidates: gui.beforeCandidates},
		{Prefix: "after:", Candidates: gui.afterCandidates},
		{Prefix: "between:", Candidates: gui.betweenCandidates},
		{Prefix: "title:", Candidates: gui.titleCandidates},
		{Prefix: "path:", Candidates: gui.pathCandidates},
		{Prefix: "parent:", Candidates: gui.parentCandidates},
		{Prefix: "sort:", Candidates: sortCandidates},
	}
	// Capture triggers slice for the "/" hint candidate closure
	hintTriggers := triggers
	triggers = append(triggers, CompletionTrigger{
		Prefix: "/",
		Candidates: func(filter string) []CompletionItem {
			items := triggerHints(hintTriggers)
			if filter == "" {
				return items
			}
			filter = strings.ToLower(filter)
			var filtered []CompletionItem
			for _, item := range items {
				if strings.Contains(strings.ToLower(item.Label), filter) ||
					strings.Contains(strings.ToLower(item.Detail), filter) {
					filtered = append(filtered, item)
				}
			}
			return filtered
		},
	})
	return triggers
}

// markdownCandidates returns common Markdown syntax snippets.
func markdownCandidates(filter string) []CompletionItem {
	items := []CompletionItem{
		{Label: "# Heading 1", InsertText: "#", Detail: "h1"},
		{Label: "## Heading 2", InsertText: "##", Detail: "h2"},
		{Label: "### Heading 3", InsertText: "###", Detail: "h3"},
		{Label: "- List item", InsertText: "-", Detail: "bullet"},
		{Label: "1. Numbered", InsertText: "1.", Detail: "ordered"},
		{Label: "- [ ] Task", InsertText: "- [ ]", Detail: "checkbox"},
		{Label: "> Quote", InsertText: ">", Detail: "blockquote"},
		{Label: "--- Rule", InsertText: "---", Detail: "divider"},
		{Label: "``` Code block", InsertText: "```\n", Detail: "code", ContinueCompleting: true},
		{Label: "**bold**", InsertText: "**", Detail: "bold", ContinueCompleting: true},
		{Label: "*italic*", InsertText: "*", Detail: "italic", ContinueCompleting: true},
		{Label: "[link](url)", InsertText: "[]()", Detail: "link", ContinueCompleting: true},
		{Label: "[[wikilink]]", InsertText: "[[", Detail: "wikilink", ContinueCompleting: true},
	}

	if filter == "" {
		return items
	}

	filter = strings.ToLower(filter)
	var filtered []CompletionItem
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Label), filter) ||
			strings.Contains(strings.ToLower(item.Detail), filter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// parentCaptureCandidates returns parent note candidates for the > trigger in capture mode.
// At the top level it shows bookmarked parents; >> shows all notes; after drilling with / it shows children.
func (gui *Gui) parentCaptureCandidates(filter string) []CompletionItem {
	state := gui.state.CaptureCompletion

	// Detect >> mode (all notes) and strip the extra > for path parsing
	allNotesMode := strings.HasPrefix(filter, ">")
	workingFilter := filter
	if allNotesMode {
		workingFilter = filter[1:]
	}

	// Determine the typing filter (text after the last /)
	typingFilter := workingFilter
	if idx := strings.LastIndex(workingFilter, "/"); idx >= 0 {
		typingFilter = workingFilter[idx+1:]
	}

	// Sync drill stack: if user backspaced past a /, truncate the stack
	slashCount := strings.Count(workingFilter, "/")
	if slashCount < len(state.ParentDrill) {
		state.ParentDrill = state.ParentDrill[:slashCount]
	}

	typingFilter = strings.ToLower(typingFilter)

	if len(state.ParentDrill) == 0 {
		if allNotesMode {
			return gui.allNoteCandidates(typingFilter)
		}
		// Top level: show bookmarked parents
		var items []CompletionItem
		for _, p := range gui.state.Parents.Items {
			if typingFilter != "" && !strings.Contains(strings.ToLower(p.Name), typingFilter) &&
				!strings.Contains(strings.ToLower(p.Title), typingFilter) {
				continue
			}
			items = append(items, CompletionItem{
				Label:  p.Name,
				Detail: p.Title,
				Value:  p.UUID,
			})
		}
		return items
	}

	// Drilled: fetch children of the last drilled parent
	lastUUID := state.ParentDrill[len(state.ParentDrill)-1].UUID
	children, err := gui.ruinCmd.Search.Search("parent:"+lastUUID, commands.SearchOptions{
		Sort:  "created:desc",
		Limit: 50,
	})
	if err != nil {
		return nil
	}

	var items []CompletionItem
	for _, note := range children {
		if typingFilter != "" && !strings.Contains(strings.ToLower(note.Title), typingFilter) {
			continue
		}
		items = append(items, CompletionItem{
			Label:  note.Title,
			Detail: note.ShortDate(),
			Value:  note.UUID,
		})
	}
	return items
}

// allNoteCandidates returns all notes as parent candidates (for >> mode).
func (gui *Gui) allNoteCandidates(filter string) []CompletionItem {
	seen := make(map[string]bool)
	var items []CompletionItem
	for _, note := range gui.state.Notes.Items {
		if note.Title == "" || seen[note.Title] {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(note.Title), filter) {
			continue
		}
		seen[note.Title] = true
		items = append(items, CompletionItem{
			Label:  note.Title,
			Detail: note.ShortDate(),
			Value:  note.UUID,
		})
	}
	return items
}

// captureTriggers returns the completion triggers for the capture popup.
func (gui *Gui) captureTriggers() []CompletionTrigger {
	return []CompletionTrigger{
		{Prefix: "[[", Candidates: gui.wikiLinkCandidates},
		{Prefix: "#", Candidates: gui.tagCandidates},
		{Prefix: ">", Candidates: gui.parentCaptureCandidates},
		{Prefix: "/", Candidates: markdownCandidates},
	}
}

const maxSuggestionItems = 6

// renderSuggestionView creates or updates a suggestion dropdown view at the given position.
// It returns the view name so the caller can manage it.
func (gui *Gui) renderSuggestionView(g *gocui.Gui, viewName string, state *CompletionState, x0, y0, maxWidth int) error {
	if !state.Active || len(state.Items) == 0 {
		g.DeleteView(viewName)
		return nil
	}

	// Calculate dimensions
	itemCount := min(len(state.Items), maxSuggestionItems)

	// Find max label and detail widths for column alignment
	maxLabelW := 0
	maxDetailW := 0
	for _, item := range state.Items {
		if lw := len([]rune(item.Label)); lw > maxLabelW {
			maxLabelW = lw
		}
		if dw := len([]rune(item.Detail)); dw > maxDetailW {
			maxDetailW = dw
		}
	}

	// width = " " + label column + gap + detail column + " " + frame
	width := max(1+maxLabelW+2+maxDetailW+1+2, 20)
	if width > maxWidth {
		width = maxWidth
	}

	x1 := x0 + width
	y1 := y0 + itemCount + 1 // +1 for border

	v, err := g.SetView(viewName, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Frame = true
	v.FrameColor = gocui.ColorYellow
	v.TitleColor = gocui.ColorYellow
	setRoundedCorners(v)

	v.Clear()

	innerWidth := max(width-2, 10)

	// Determine visible window for scrolling
	startIdx := 0
	if state.SelectedIndex >= maxSuggestionItems {
		startIdx = state.SelectedIndex - maxSuggestionItems + 1
	}
	endIdx := startIdx + itemCount
	if endIdx > len(state.Items) {
		endIdx = len(state.Items)
		startIdx = max(endIdx-itemCount, 0)
	}

	// Detail column starts at a fixed position
	detailCol := innerWidth - maxDetailW - 1 // 1 for trailing space

	for i := startIdx; i < endIdx; i++ {
		item := state.Items[i]
		selected := i == state.SelectedIndex

		label := " " + item.Label
		detail := item.Detail

		// Pad label to reach detail column
		labelRunes := len([]rune(label))
		pad := max(detailCol-labelRunes, 1)

		line := label + strings.Repeat(" ", pad) + AnsiDim + detail + AnsiReset
		// Pad to full width for highlight (account for ANSI not taking visual space)
		visualLen := labelRunes + pad + len([]rune(detail))
		line = line + strings.Repeat(" ", max(innerWidth-visualLen, 0))

		if selected {
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line, AnsiReset)
		} else {
			fmt.Fprintln(v, line)
		}
	}

	g.SetViewOnTop(viewName)
	return nil
}
