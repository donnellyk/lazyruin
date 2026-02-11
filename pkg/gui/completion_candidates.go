package gui

import (
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/commands"
)

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
		detail := fmt.Sprintf("(%d)", tag.Count)
		if len(tag.Scope) > 0 {
			detail += " [" + strings.Join(tag.Scope, ", ") + "]"
		}
		items = append(items, CompletionItem{
			Label:      name,
			InsertText: name,
			Detail:     detail,
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
