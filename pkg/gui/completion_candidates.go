package gui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	anytime "github.com/ijt/go-anytime"
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
		items = append(items, CompletionItem{
			Label:      name,
			InsertText: name,
			Detail:     fmt.Sprintf("(%d)", tag.Count),
		})
	}
	return items
}

// currentCardTagCandidates returns tag completion items limited to tags on the current preview card.
func (gui *Gui) currentCardTagCandidates(filter string) []CompletionItem {
	card := gui.currentPreviewCard()
	if card == nil {
		return nil
	}
	filter = strings.ToLower(filter)
	var items []CompletionItem
	allTags := append(card.Tags, card.InlineTags...)
	for _, tag := range allTags {
		name := tag
		if !strings.HasPrefix(name, "#") {
			name = "#" + name
		}
		nameWithoutHash := strings.TrimPrefix(name, "#")
		if filter != "" && !strings.Contains(strings.ToLower(nameWithoutHash), filter) {
			continue
		}
		items = append(items, CompletionItem{
			Label:      name,
			InsertText: name,
		})
	}
	return items
}

// dateLiterals are the date keywords supported natively by ruin's date filters.
var dateLiterals = []struct {
	value  string
	detail string
}{
	{"today", "today"},
	{"yesterday", "yesterday"},
	{"tomorrow", "tomorrow"},
	{"this-week", "current week"},
	{"last-week", "previous week"},
	{"next-week", "next week"},
	{"this-month", "current month"},
	{"last-month", "previous month"},
	{"next-month", "next month"},
}

// dateCandidates builds completion items for a date-prefix filter (e.g. "created:", "updated:").
// It offers literal keywords (today/yesterday/tomorrow) plus dynamic natural-language parsing
// via go-anytime, showing resolved ISO dates with a human-readable detail.
func dateCandidates(prefix, filter string) []CompletionItem {
	return dateCandidatesAt(prefix, filter, time.Now())
}

// dateCandidatesAt is the testable core of dateCandidates, accepting an explicit "now" time.
func dateCandidatesAt(prefix, filter string, now time.Time) []CompletionItem {
	filterLower := strings.ToLower(filter)
	var items []CompletionItem

	// Always offer literal keywords that match the filter.
	for _, s := range dateLiterals {
		if filterLower != "" && !strings.HasPrefix(s.value, filterLower) {
			continue
		}
		items = append(items, CompletionItem{
			Label:      prefix + s.value,
			InsertText: prefix + s.value,
			Detail:     s.detail,
		})
	}

	// If the filter is non-empty and not an exact literal match, try anytime parsing.
	if filter != "" && !isDateLiteral(filterLower) {
		if parsed, err := anytime.Parse(filter, now); err == nil {
			iso := parsed.Format("2006-01-02")
			detail := parsed.Format("Mon, Jan 02, 2006")
			items = append(items, CompletionItem{
				Label:      prefix + iso,
				InsertText: prefix + iso,
				Detail:     detail,
			})
		}
	}

	return items
}

// ambientDateCandidates parses a bare token (no trigger prefix) with go-anytime
// and returns a created: suggestion if it resolves to a valid date.
// Used as a FallbackCandidates function for the search popup.
func ambientDateCandidates(token string) []CompletionItem {
	return ambientDateCandidatesAt(token, time.Now())
}

// ambientDateCandidatesAt is the testable core of ambientDateCandidates.
func ambientDateCandidatesAt(token string, now time.Time) []CompletionItem {
	parsed, err := anytime.Parse(token, now)
	if err != nil {
		return nil
	}
	iso := parsed.Format("2006-01-02")
	detail := parsed.Format("Mon, Jan 02, 2006")
	return []CompletionItem{
		{Label: "created:" + iso, InsertText: "created:" + iso, Detail: detail},
		{Label: "after:" + iso, InsertText: "after:" + iso, Detail: detail},
		{Label: "before:" + iso, InsertText: "before:" + iso, Detail: detail},
	}
}

// atDateCandidates returns completion items for the @ trigger, resolving natural
// language dates to ISO format. Used in both capture and search popups.
func atDateCandidates(filter string) []CompletionItem {
	return atDateCandidatesAt(filter, time.Now())
}

// atDateSuggestions are natural-language date strings that go-anytime can parse.
// Grouped by prefix for compact filtering.
var atDateSuggestions = []string{
	"today", "yesterday", "tomorrow",
	"this-week", "last-week", "next-week",
	"this-month", "last-month", "next-month",
	"monday", "tuesday", "wednesday", "thursday",
	"friday", "saturday", "sunday",
	"next monday", "next tuesday", "next wednesday", "next thursday",
	"next friday", "next saturday", "next sunday",
	"last monday", "last tuesday", "last wednesday", "last thursday",
	"last friday", "last saturday", "last sunday",
	"next week", "last week",
	"next month", "last month",
	"january", "february", "march", "april", "may", "june",
	"july", "august", "september", "october", "november", "december",
	"next year", "last year",
}

// daysOfNextWeek returns "next sunday" through "next saturday" for expanding "next week".
var daysOfNextWeek = []string{
	"next sunday", "next monday", "next tuesday", "next wednesday",
	"next thursday", "next friday", "next saturday",
}

// atDateCandidatesAt is the testable core of atDateCandidates.
func atDateCandidatesAt(filter string, now time.Time) []CompletionItem {
	filterLower := strings.ToLower(filter)

	// "next week" / "last week" → expand to individual days
	if filterLower == "next week" || filterLower == "last week" {
		prefix := strings.SplitN(filterLower, " ", 2)[0] // "next" or "last"
		var items []CompletionItem
		for _, day := range daysOfNextWeek {
			s := prefix + day[4:] // "next" + " sunday", etc.
			items = append(items, atSuggestionItem(s, now))
		}
		return items
	}

	// Filter suggestions by prefix match
	var items []CompletionItem
	for _, s := range atDateSuggestions {
		if filterLower != "" && !strings.HasPrefix(s, filterLower) {
			continue
		}
		items = append(items, atSuggestionItem(s, now))
	}

	// If nothing matched from suggestions, try US date formats then anytime
	if len(items) == 0 && filter != "" {
		if parsed, ok := parseUSDate(filter); ok {
			iso := parsed.Format("2006-01-02")
			items = append(items, CompletionItem{
				Label:      "@" + iso,
				InsertText: "@" + iso,
				Detail:     parsed.Format("Mon, Jan 02"),
			})
		} else if parsed, err := anytime.Parse(filter, now); err == nil {
			iso := parsed.Format("2006-01-02")
			items = append(items, CompletionItem{
				Label:      "@" + iso,
				InsertText: "@" + iso,
				Detail:     parsed.Format("Mon, Jan 02"),
			})
		}
	}

	return items
}

// parseUSDate tries MM/DD/YYYY and MM/DD/YY formats, returning the parsed time.
func parseUSDate(s string) (time.Time, bool) {
	for _, layout := range []string{"1/2/2006", "1/2/06"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// cliNativeDates are the keywords the ruin CLI handles directly; all other
// suggestions must be resolved to @YYYY-MM-DD before insertion.
var cliNativeDates = map[string]bool{
	"today": true, "yesterday": true, "tomorrow": true,
	"this-week": true, "last-week": true, "next-week": true,
	"this-month": true, "last-month": true, "next-month": true,
}

// atSuggestionItem builds a CompletionItem for a natural-language date string,
// resolving it via go-anytime for the detail preview. CLI-native keywords
// (today/yesterday/tomorrow) insert as-is; everything else inserts as @ISO date.
func atSuggestionItem(s string, now time.Time) CompletionItem {
	parsed, err := anytime.Parse(s, now)
	if err != nil {
		return CompletionItem{Label: "@" + s, InsertText: "@" + s}
	}
	insert := "@" + s
	if !cliNativeDates[s] {
		insert = "@" + parsed.Format("2006-01-02")
	}
	return CompletionItem{
		Label:      "@" + s,
		InsertText: insert,
		Detail:     parsed.Format("Mon, Jan 02"),
	}
}

// isDateLiteral returns true if s exactly matches a supported literal keyword.
func isDateLiteral(s string) bool {
	for _, lit := range dateLiterals {
		if s == lit.value {
			return true
		}
	}
	return false
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

// betweenCandidates returns between: filter suggestions with computed ISO date ranges.
func (gui *Gui) betweenCandidates(filter string) []CompletionItem {
	return betweenCandidatesAt(filter, time.Now())
}

// betweenCandidatesAt is the testable core of betweenCandidates.
func betweenCandidatesAt(filter string, now time.Time) []CompletionItem {
	today := now.Format("2006-01-02")
	weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
	monthAgo := now.AddDate(0, -1, 0).Format("2006-01-02")
	yearAgo := now.AddDate(-1, 0, 0).Format("2006-01-02")

	shortcuts := []CompletionItem{
		{Label: "between:" + weekAgo + "," + today, InsertText: "between:" + weekAgo + "," + today, Detail: "last 7 days"},
		{Label: "between:" + monthAgo + "," + today, InsertText: "between:" + monthAgo + "," + today, Detail: "last 30 days"},
		{Label: "between:" + yearAgo + "," + today, InsertText: "between:" + yearAgo + "," + today, Detail: "last year"},
	}

	if filter == "" {
		return shortcuts
	}

	filterLower := strings.ToLower(filter)
	var items []CompletionItem
	for _, s := range shortcuts {
		suffix := strings.TrimPrefix(s.InsertText, "between:")
		if strings.Contains(suffix, filterLower) || strings.Contains(s.Detail, filterLower) {
			items = append(items, s)
		}
	}

	// Try anytime parsing: offer between:<parsed>,<today>
	if len(items) == 0 {
		if parsed, err := anytime.Parse(filter, now); err == nil {
			iso := parsed.Format("2006-01-02")
			detail := parsed.Format("Mon, Jan 02, 2006") + " to today"
			items = append(items, CompletionItem{
				Label:      "between:" + iso + "," + today,
				InsertText: "between:" + iso + "," + today,
				Detail:     detail,
			})
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
		{Label: "# Heading 1", InsertText: "#", Detail: "h1", PrependToLine: true},
		{Label: "## Heading 2", InsertText: "##", Detail: "h2", PrependToLine: true},
		{Label: "### Heading 3", InsertText: "###", Detail: "h3", PrependToLine: true},
		{Label: "- List item", InsertText: "-", Detail: "bullet", PrependToLine: true},
		{Label: "1. Numbered", InsertText: "1.", Detail: "ordered", PrependToLine: true},
		{Label: "- [ ] Task", InsertText: "- [ ]", Detail: "checkbox", PrependToLine: true},
		{Label: "> Quote", InsertText: ">", Detail: "blockquote", PrependToLine: true},
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

// parentCandidatesFor returns a candidate function for the > trigger that uses the given
// CompletionState for drill-down tracking. This allows both the capture editor and
// parent input popup to share the same parent-selection logic.
func (gui *Gui) parentCandidatesFor(completionState *CompletionState) func(string) []CompletionItem {
	return func(filter string) []CompletionItem {
		state := completionState

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
}

// abbreviationCandidates returns user-defined abbreviation snippets filtered by key.
func (gui *Gui) abbreviationCandidates(filter string) []CompletionItem {
	if len(gui.config.Abbreviations) == 0 {
		return nil
	}
	filter = strings.ToLower(filter)
	keys := make([]string, 0, len(gui.config.Abbreviations))
	for k := range gui.config.Abbreviations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var items []CompletionItem
	for _, k := range keys {
		if filter != "" && !strings.Contains(strings.ToLower(k), filter) {
			continue
		}
		expansion := gui.config.Abbreviations[k]
		detail := expansion
		if len(detail) > 40 {
			detail = detail[:37] + "..."
		}
		items = append(items, CompletionItem{
			Label:      "!" + k,
			InsertText: expansion,
			Detail:     detail,
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
