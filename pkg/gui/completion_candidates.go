package gui

import (
	"sort"
	"strings"

	"kvnd/lazyruin/pkg/gui/types"
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
		{Label: "inbox", InsertText: "", Detail: "insert from inbox", Value: "action:inbox"},
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
