package gui

import "strings"

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

// captureTriggers returns the completion triggers for the capture popup.
func (gui *Gui) captureTriggers() []CompletionTrigger {
	return []CompletionTrigger{
		{Prefix: "[[", Candidates: gui.wikiLinkCandidates},
		{Prefix: "#", Candidates: gui.tagCandidates},
		{Prefix: ">", Candidates: gui.parentCaptureCandidates},
		{Prefix: "/", Candidates: markdownCandidates},
	}
}

// pickTriggers returns the completion triggers for the pick popup.
// Only tag completion is supported since pick only accepts inline tags.
func (gui *Gui) pickTriggers() []CompletionTrigger {
	return []CompletionTrigger{
		{Prefix: "#", Candidates: gui.tagCandidates},
	}
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
