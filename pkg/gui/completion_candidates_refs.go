package gui

import (
	"fmt"
	"strings"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

// titleCandidates returns note titles as completion items.
func (gui *Gui) titleCandidates(filter string) []types.CompletionItem {
	filter = strings.ToLower(filter)
	seen := make(map[string]bool)
	var items []types.CompletionItem
	for _, note := range gui.contexts.Notes.Items {
		title := note.Title
		if title == "" || seen[title] {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(title), filter) {
			continue
		}
		seen[title] = true
		items = append(items, types.CompletionItem{
			Label:      "title:" + title,
			InsertText: "title:" + title,
			Detail:     note.ShortDate(),
		})
	}
	return items
}

// wikiLinkCandidates returns note titles for [[ wiki-style link completion.
// When the filter contains '#', it switches to header mode for the specified note.
func (gui *Gui) wikiLinkCandidates(filter string) []types.CompletionItem {
	// Header mode: filter contains '#'
	if noteTitle, after, ok := strings.Cut(filter, "#"); ok {
		headerFilter := strings.ToLower(after)
		return gui.headerCandidates(noteTitle, headerFilter)
	}

	filterLower := strings.ToLower(filter)
	seen := make(map[string]bool)
	var items []types.CompletionItem
	for _, note := range gui.contexts.Notes.Items {
		title := note.Title
		if title == "" || seen[title] {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(title), filterLower) {
			continue
		}
		seen[title] = true
		items = append(items, types.CompletionItem{
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
func (gui *Gui) headerCandidates(noteTitle, filter string) []types.CompletionItem {
	// Find the note by exact title match
	var content string
	for i, note := range gui.contexts.Notes.Items {
		if note.Title == noteTitle {
			if note.Content == "" {
				loaded, err := gui.loadNoteContent(note.Path)
				if err != nil {
					return nil
				}
				gui.contexts.Notes.Items[i].Content = loaded
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
	var items []types.CompletionItem
	for _, h := range headers {
		if filter != "" && !strings.Contains(strings.ToLower(h.Text), filter) {
			continue
		}
		items = append(items, types.CompletionItem{
			Label:      h.Text,
			InsertText: "[[" + noteTitle + "#" + h.Text + "]]",
			Detail:     fmt.Sprintf("h%d", h.Level),
		})
	}
	return items
}

// pathCandidates returns path: filter hint.
func (gui *Gui) pathCandidates(filter string) []types.CompletionItem {
	if filter != "" {
		return nil
	}
	return []types.CompletionItem{
		{Label: "path:", InsertText: "path:", Detail: "search by path"},
	}
}

// parentCandidates returns parent: filter suggestions.
func (gui *Gui) parentCandidates(filter string) []types.CompletionItem {
	shortcuts := []types.CompletionItem{
		{Label: "parent:none", InsertText: "parent:none", Detail: "root notes only"},
	}
	// Add known parent bookmarks
	for _, p := range gui.contexts.Queries.Parents {
		item := types.CompletionItem{
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
	var items []types.CompletionItem
	for _, s := range shortcuts {
		suffix := strings.TrimPrefix(s.Label, "parent:")
		if strings.Contains(strings.ToLower(suffix), filter) ||
			strings.Contains(strings.ToLower(s.Detail), filter) {
			items = append(items, s)
		}
	}
	return items
}
