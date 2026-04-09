package gui

import (
	"github.com/donnellyk/lazyruin/pkg/gui/helpers"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
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

// tagTrigger returns the # tag completion trigger.
func (gui *Gui) tagTrigger() types.CompletionTrigger {
	return types.CompletionTrigger{Prefix: "#", Candidates: gui.TagCandidates}
}

// dateTrigger returns the @ date completion trigger.
func (gui *Gui) dateTrigger() types.CompletionTrigger {
	return types.CompletionTrigger{Prefix: "@", Candidates: atDateCandidates}
}

// abbreviationTrigger returns the ! abbreviation completion trigger.
func (gui *Gui) abbreviationTrigger() types.CompletionTrigger {
	return types.CompletionTrigger{Prefix: "!", Candidates: gui.abbreviationCandidates}
}

// wikiLinkTrigger returns the [[ wiki-link completion trigger.
func (gui *Gui) wikiLinkTrigger() types.CompletionTrigger {
	return types.CompletionTrigger{Prefix: "[[", Candidates: gui.wikiLinkCandidates}
}

// searchTriggers returns the completion triggers for the search popup.
// The "/" trigger shows an overview of all available filter prefixes.
func (gui *Gui) searchTriggers() []types.CompletionTrigger {
	triggers := []types.CompletionTrigger{
		gui.abbreviationTrigger(),
		gui.tagTrigger(),
		gui.dateTrigger(),
		{Prefix: "created:", Candidates: func(f string) []types.CompletionItem { return dateCandidates("created:", f) }},
		{Prefix: "updated:", Candidates: func(f string) []types.CompletionItem { return dateCandidates("updated:", f) }},
		{Prefix: "before:", Candidates: func(f string) []types.CompletionItem { return dateCandidates("before:", f) }},
		{Prefix: "after:", Candidates: func(f string) []types.CompletionItem { return dateCandidates("after:", f) }},
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

// searchOrFilterTriggers returns source-specific triggers in filter mode,
// falling back to searchTriggers otherwise.
func (gui *Gui) searchOrFilterTriggers() []types.CompletionTrigger {
	if gui.contexts.Search.FilterTriggers != nil {
		return gui.contexts.Search.FilterTriggers()
	}
	return gui.searchTriggers()
}

// captureTriggers returns the completion triggers for the capture popup.
func (gui *Gui) captureTriggers() []types.CompletionTrigger {
	return []types.CompletionTrigger{
		gui.abbreviationTrigger(),
		gui.wikiLinkTrigger(),
		gui.tagTrigger(),
		gui.dateTrigger(),
		{Prefix: ">", Candidates: gui.ParentCandidatesFor(gui.contexts.Capture.Completion)},
		{Prefix: "/", Candidates: markdownCandidates},
	}
}

// inboxTriggers returns the completion triggers for the inbox input popup.
// A subset of capture triggers: tags, wiki-links, dates.
func (gui *Gui) inboxTriggers() []types.CompletionTrigger {
	return []types.CompletionTrigger{
		gui.tagTrigger(),
		gui.wikiLinkTrigger(),
		gui.dateTrigger(),
	}
}

// pickTriggers returns the completion triggers for the pick popup.
// In dialog mode, tag suggestions are scoped to inline tags in the underlying document.
func (gui *Gui) pickTriggers() []types.CompletionTrigger {
	tagTrigger := gui.tagTrigger()
	if gui.contexts.Pick.DialogMode {
		tagTrigger.Candidates = gui.ScopedInlineTagCandidates
	}
	return []types.CompletionTrigger{
		tagTrigger,
		gui.dateTrigger(),
		{Prefix: "--", Candidates: flagCandidates},
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

// extractSort delegates to helpers.ExtractSort.
func extractSort(query string) (string, string) {
	return helpers.ExtractSort(query)
}
