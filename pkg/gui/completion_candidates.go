package gui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"kvnd/lazyruin/pkg/gui/types"

	anytime "github.com/ijt/go-anytime"
)

// TagCandidates delegates to CompletionHelper.
func (gui *Gui) TagCandidates(filter string) []types.CompletionItem {
	return gui.helpers.Completion().TagCandidates(filter)
}

// CurrentCardTagCandidates delegates to CompletionHelper.
func (gui *Gui) CurrentCardTagCandidates(filter string) []types.CompletionItem {
	return gui.helpers.Completion().CurrentCardTagCandidates(filter)
}

// ScopedInlineTagCandidates delegates to CompletionHelper.
func (gui *Gui) ScopedInlineTagCandidates(filter string) []types.CompletionItem {
	return gui.helpers.Completion().ScopedInlineTagCandidates(filter)
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
func dateCandidates(prefix, filter string) []types.CompletionItem {
	return dateCandidatesAt(prefix, filter, time.Now())
}

// dateCandidatesAt is the testable core of dateCandidates, accepting an explicit "now" time.
func dateCandidatesAt(prefix, filter string, now time.Time) []types.CompletionItem {
	filterLower := strings.ToLower(filter)
	var items []types.CompletionItem

	// Always offer literal keywords that match the filter.
	for _, s := range dateLiterals {
		if filterLower != "" && !strings.HasPrefix(s.value, filterLower) {
			continue
		}
		items = append(items, types.CompletionItem{
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
			items = append(items, types.CompletionItem{
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
func ambientDateCandidates(token string) []types.CompletionItem {
	return ambientDateCandidatesAt(token, time.Now())
}

// ambientDateCandidatesAt is the testable core of ambientDateCandidates.
func ambientDateCandidatesAt(token string, now time.Time) []types.CompletionItem {
	parsed, err := anytime.Parse(token, now)
	if err != nil {
		return nil
	}
	iso := parsed.Format("2006-01-02")
	detail := parsed.Format("Mon, Jan 02, 2006")
	return []types.CompletionItem{
		{Label: "created:" + iso, InsertText: "created:" + iso, Detail: detail},
		{Label: "after:" + iso, InsertText: "after:" + iso, Detail: detail},
		{Label: "before:" + iso, InsertText: "before:" + iso, Detail: detail},
	}
}

// atDateCandidates returns completion items for the @ trigger, resolving natural
// language dates to ISO format. Used in both capture and search popups.
func atDateCandidates(filter string) []types.CompletionItem {
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
func atDateCandidatesAt(filter string, now time.Time) []types.CompletionItem {
	filterLower := strings.ToLower(filter)

	// "next week" / "last week" → expand to individual days
	if filterLower == "next week" || filterLower == "last week" {
		prefix := strings.SplitN(filterLower, " ", 2)[0] // "next" or "last"
		var items []types.CompletionItem
		for _, day := range daysOfNextWeek {
			s := prefix + day[4:] // "next" + " sunday", etc.
			items = append(items, atSuggestionItem(s, now))
		}
		return items
	}

	// Filter suggestions by prefix match
	var items []types.CompletionItem
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
			items = append(items, types.CompletionItem{
				Label:      "@" + iso,
				InsertText: "@" + iso,
				Detail:     parsed.Format("Mon, Jan 02"),
			})
		} else if parsed, err := anytime.Parse(filter, now); err == nil {
			iso := parsed.Format("2006-01-02")
			items = append(items, types.CompletionItem{
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

// atSuggestionItem builds a types.CompletionItem for a natural-language date string,
// resolving it via go-anytime for the detail preview. CLI-native keywords
// (today/yesterday/tomorrow) insert as-is; everything else inserts as @ISO date.
func atSuggestionItem(s string, now time.Time) types.CompletionItem {
	parsed, err := anytime.Parse(s, now)
	if err != nil {
		return types.CompletionItem{Label: "@" + s, InsertText: "@" + s}
	}
	insert := "@" + s
	if !cliNativeDates[s] {
		insert = "@" + parsed.Format("2006-01-02")
	}
	return types.CompletionItem{
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

func (gui *Gui) createdCandidates(filter string) []types.CompletionItem {
	return dateCandidates("created:", filter)
}

func (gui *Gui) updatedCandidates(filter string) []types.CompletionItem {
	return dateCandidates("updated:", filter)
}

func (gui *Gui) beforeCandidates(filter string) []types.CompletionItem {
	return dateCandidates("before:", filter)
}

func (gui *Gui) afterCandidates(filter string) []types.CompletionItem {
	return dateCandidates("after:", filter)
}

// betweenCandidates returns between: filter suggestions with computed ISO date ranges.
func (gui *Gui) betweenCandidates(filter string) []types.CompletionItem {
	return betweenCandidatesAt(filter, time.Now())
}

// betweenCandidatesAt is the testable core of betweenCandidates.
func betweenCandidatesAt(filter string, now time.Time) []types.CompletionItem {
	today := now.Format("2006-01-02")
	weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
	monthAgo := now.AddDate(0, -1, 0).Format("2006-01-02")
	yearAgo := now.AddDate(-1, 0, 0).Format("2006-01-02")

	shortcuts := []types.CompletionItem{
		{Label: "between:" + weekAgo + "," + today, InsertText: "between:" + weekAgo + "," + today, Detail: "last 7 days"},
		{Label: "between:" + monthAgo + "," + today, InsertText: "between:" + monthAgo + "," + today, Detail: "last 30 days"},
		{Label: "between:" + yearAgo + "," + today, InsertText: "between:" + yearAgo + "," + today, Detail: "last year"},
	}

	if filter == "" {
		return shortcuts
	}

	filterLower := strings.ToLower(filter)
	var items []types.CompletionItem
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
			items = append(items, types.CompletionItem{
				Label:      "between:" + iso + "," + today,
				InsertText: "between:" + iso + "," + today,
				Detail:     detail,
			})
		}
	}

	return items
}

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

// sortCandidates returns sort: completion items for the search popup.
func sortCandidates(filter string) []types.CompletionItem {
	items := []types.CompletionItem{
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
	var filtered []types.CompletionItem
	for _, item := range items {
		suffix := strings.TrimPrefix(item.InsertText, "sort:")
		if strings.Contains(suffix, filter) || strings.Contains(strings.ToLower(item.Detail), filter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// markdownCandidates returns common Markdown syntax snippets.
func markdownCandidates(filter string) []types.CompletionItem {
	items := []types.CompletionItem{
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
	var filtered []types.CompletionItem
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Label), filter) ||
			strings.Contains(strings.ToLower(item.Detail), filter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// ParentCandidatesFor delegates to CompletionHelper.
func (gui *Gui) ParentCandidatesFor(completionState *types.CompletionState) func(string) []types.CompletionItem {
	return gui.helpers.Completion().ParentCandidatesFor(completionState)
}

// abbreviationCandidates returns user-defined abbreviation snippets filtered by key.
func (gui *Gui) abbreviationCandidates(filter string) []types.CompletionItem {
	abbrevs := gui.config.VaultAbbreviations(gui.ruinCmd.VaultPath())
	if len(abbrevs) == 0 {
		return nil
	}
	filter = strings.ToLower(filter)
	keys := make([]string, 0, len(abbrevs))
	for k := range abbrevs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var items []types.CompletionItem
	for _, k := range keys {
		if filter != "" && !strings.Contains(strings.ToLower(k), filter) {
			continue
		}
		expansion := abbrevs[k]
		detail := expansion
		if len(detail) > 40 {
			detail = detail[:37] + "..."
		}
		items = append(items, types.CompletionItem{
			Label:      "!" + k,
			InsertText: expansion,
			Detail:     detail,
		})
	}
	return items
}

// flagCandidates returns --any and --todo flag suggestions for the pick popup.
func flagCandidates(filter string) []types.CompletionItem {
	items := []types.CompletionItem{
		{Label: "--any", InsertText: "--any", Detail: "match any tag"},
		{Label: "--todo", InsertText: "--todo", Detail: "only todo lines"},
		{Label: "--all-tags", InsertText: "--all-tags", Detail: "all scoped inline tags"},
	}
	if filter == "" {
		return items
	}
	filter = strings.ToLower(filter)
	var filtered []types.CompletionItem
	for _, item := range items {
		if strings.Contains(item.Label, filter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
