package gui

import (
	"kvnd/lazyruin/pkg/gui/types"
	"strings"
)

// triggerHints builds overview CompletionItems for each trigger prefix,
// shown when the input is empty or cursor is at whitespace.
func triggerHints(triggers []types.CompletionTrigger) []types.CompletionItem {
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
		"@":        "insert date",
		"!":        "abbreviation",
	}
	var items []types.CompletionItem
	for _, t := range triggers {
		if t.Prefix == "/" {
			continue // don't include the / trigger itself in its own hints
		}
		detail := descriptions[t.Prefix]
		if detail == "" {
			detail = "filter"
		}
		items = append(items, types.CompletionItem{
			Label:              t.Prefix,
			InsertText:         t.Prefix,
			Detail:             detail,
			ContinueCompleting: true,
		})
	}
	return items
}

// searchTriggers returns the completion triggers for the search popup.
// The "/" trigger shows an overview of all available filter prefixes.
func (gui *Gui) searchTriggers() []types.CompletionTrigger {
	triggers := []types.CompletionTrigger{
		{Prefix: "!", Candidates: gui.abbreviationCandidates},
		{Prefix: "#", Candidates: gui.TagCandidates},
		{Prefix: "@", Candidates: atDateCandidates},
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
	triggers = append(triggers, types.CompletionTrigger{
		Prefix: "/",
		Candidates: func(filter string) []types.CompletionItem {
			items := triggerHints(hintTriggers)
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
		},
	})
	return triggers
}

// captureTriggers returns the completion triggers for the capture popup.
func (gui *Gui) captureTriggers() []types.CompletionTrigger {
	return []types.CompletionTrigger{
		{Prefix: "!", Candidates: gui.abbreviationCandidates},
		{Prefix: "[[", Candidates: gui.wikiLinkCandidates},
		{Prefix: "#", Candidates: gui.TagCandidates},
		{Prefix: "@", Candidates: atDateCandidates},
		{Prefix: ">", Candidates: gui.ParentCandidatesFor(gui.contexts.Capture.Completion)},
		{Prefix: "/", Candidates: markdownCandidates},
	}
}

// pickTriggers returns the completion triggers for the pick popup.
func (gui *Gui) pickTriggers() []types.CompletionTrigger {
	return []types.CompletionTrigger{
		{Prefix: "#", Candidates: gui.TagCandidates},
		{Prefix: "@", Candidates: atDateCandidates},
	}
}

// snippetExpansionTriggers returns the completion triggers for the snippet
// expansion field. It merges search and capture triggers, excluding ! (to
// avoid recursion) and rebinding > to use SnippetEditor's completion state.
func (gui *Gui) snippetExpansionTriggers() []types.CompletionTrigger {
	seen := make(map[string]bool)
	var merged []types.CompletionTrigger

	for _, t := range gui.captureTriggers() {
		if t.Prefix == "!" {
			continue
		}
		if t.Prefix == ">" {
			t.Candidates = gui.ParentCandidatesFor(gui.contexts.SnippetEditor.Completion)
		}
		seen[t.Prefix] = true
		merged = append(merged, t)
	}

	for _, t := range gui.searchTriggers() {
		if t.Prefix == "!" || seen[t.Prefix] {
			continue
		}
		seen[t.Prefix] = true
		merged = append(merged, t)
	}

	return merged
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
