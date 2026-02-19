package helpers

import (
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/types"
)

// CompletionHelper provides completion candidate functions that were previously
// on *Gui. Helpers and controllers access them via Helpers().Completion().
type CompletionHelper struct {
	c *HelperCommon
}

// NewCompletionHelper creates a new CompletionHelper.
func NewCompletionHelper(c *HelperCommon) *CompletionHelper {
	return &CompletionHelper{c: c}
}

// TagCandidates returns tag completion items filtered by the given prefix.
func (self *CompletionHelper) TagCandidates(filter string) []types.CompletionItem {
	filter = strings.ToLower(filter)
	var items []types.CompletionItem
	for _, tag := range self.c.GuiCommon().Contexts().Tags.Items {
		name := tag.Name
		if !strings.HasPrefix(name, "#") {
			name = "#" + name
		}
		nameWithoutHash := strings.TrimPrefix(name, "#")
		if filter != "" && !strings.Contains(strings.ToLower(nameWithoutHash), filter) {
			continue
		}
		items = append(items, types.CompletionItem{
			Label:      name,
			InsertText: name,
			Detail:     fmt.Sprintf("(%d)", tag.Count),
		})
	}
	return items
}

// CurrentCardTagCandidates returns tag completion items limited to tags on the current preview card.
func (self *CompletionHelper) CurrentCardTagCandidates(filter string) []types.CompletionItem {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	filter = strings.ToLower(filter)
	var items []types.CompletionItem
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
		items = append(items, types.CompletionItem{
			Label:      name,
			InsertText: name,
		})
	}
	return items
}

// ParentCandidatesFor returns a candidate function for the > trigger that uses the given
// CompletionState for drill-down tracking.
func (self *CompletionHelper) ParentCandidatesFor(completionState *types.CompletionState) func(string) []types.CompletionItem {
	return func(filter string) []types.CompletionItem {
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
				return self.allNoteCandidates(typingFilter)
			}
			// Top level: show bookmarked parents
			var items []types.CompletionItem
			for _, p := range self.c.GuiCommon().Contexts().Queries.Parents {
				if typingFilter != "" && !strings.Contains(strings.ToLower(p.Name), typingFilter) &&
					!strings.Contains(strings.ToLower(p.Title), typingFilter) {
					continue
				}
				items = append(items, types.CompletionItem{
					Label:  p.Name,
					Detail: p.Title,
					Value:  p.UUID,
				})
			}
			return items
		}

		// Drilled: fetch children of the last drilled parent
		lastUUID := state.ParentDrill[len(state.ParentDrill)-1].UUID
		children, err := self.c.RuinCmd().Search.Search("parent:"+lastUUID, commands.SearchOptions{
			Sort:  "created:desc",
			Limit: 50,
		})
		if err != nil {
			return nil
		}

		var items []types.CompletionItem
		for _, note := range children {
			if typingFilter != "" && !strings.Contains(strings.ToLower(note.Title), typingFilter) {
				continue
			}
			items = append(items, types.CompletionItem{
				Label:  note.Title,
				Detail: note.ShortDate(),
				Value:  note.UUID,
			})
		}
		return items
	}
}

// allNoteCandidates returns all notes as parent candidates (for >> mode).
func (self *CompletionHelper) allNoteCandidates(filter string) []types.CompletionItem {
	seen := make(map[string]bool)
	var items []types.CompletionItem
	for _, note := range self.c.GuiCommon().Contexts().Notes.Items {
		if note.Title == "" || seen[note.Title] {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(note.Title), filter) {
			continue
		}
		seen[note.Title] = true
		items = append(items, types.CompletionItem{
			Label:  note.Title,
			Detail: note.ShortDate(),
			Value:  note.UUID,
		})
	}
	return items
}
