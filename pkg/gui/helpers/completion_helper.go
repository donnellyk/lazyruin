package helpers

import (
	"fmt"
	"strings"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/ruin-note-cli/pkg/notetext"
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

// ScopedInlineTags returns the unique inline tags found in the underlying
// preview document (compose note or cardList card). Each tag includes the
// leading # prefix. Returns nil when no scoped content is available.
//
// Uses the shared notetext extractor so tags inside code spans, fenced code
// blocks, and markdown links are excluded — same rules `ruin log` applies.
func (self *CompletionHelper) ScopedInlineTags() []string {
	gui := self.c.GuiCommon()
	var content string
	switch gui.Contexts().ActivePreviewKey {
	case "compose":
		content = gui.Contexts().Compose.Note.Content
	case "cardList":
		cl := gui.Contexts().CardList
		if cl.SelectedCardIdx < len(cl.Cards) {
			content = cl.Cards[cl.SelectedCardIdx].Content
		}
	}
	if content == "" {
		return nil
	}
	return notetext.ExtractTags(content)
}

// ScopedInlineTagCandidates returns tag completion items limited to inline tags
// found in the underlying preview document (compose note or cardList card).
func (self *CompletionHelper) ScopedInlineTagCandidates(filter string) []types.CompletionItem {
	tags := self.ScopedInlineTags()
	if tags == nil {
		return self.TagCandidates(filter)
	}

	filter = strings.ToLower(filter)
	var items []types.CompletionItem
	for _, name := range tags {
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

		// Determine the typing filter (text after the last /)
		typingFilter := filter
		if idx := strings.LastIndex(filter, "/"); idx >= 0 {
			typingFilter = filter[idx+1:]
		}

		// Sync drill stack: if user backspaced past a /, truncate the stack
		slashCount := strings.Count(filter, "/")
		if slashCount < len(state.ParentDrill) {
			state.ParentDrill = state.ParentDrill[:slashCount]
		}

		typingFilter = strings.ToLower(typingFilter)

		if len(state.ParentDrill) == 0 {
			return self.topLevelParentCandidates(typingFilter)
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

// topLevelParentCandidates returns two sections — bookmarks and all notes —
// filtered by the given typing filter. Empty sections are omitted; notes
// whose title matches a bookmark in the first section are skipped in the
// second so they aren't listed twice.
func (self *CompletionHelper) topLevelParentCandidates(filter string) []types.CompletionItem {
	gui := self.c.GuiCommon()

	var bookmarks []types.CompletionItem
	bookmarkedUUIDs := make(map[string]bool)
	for _, p := range gui.Contexts().Queries.Parents {
		bookmarkedUUIDs[p.UUID] = true
		if filter != "" && !strings.Contains(strings.ToLower(p.Name), filter) &&
			!strings.Contains(strings.ToLower(p.Title), filter) {
			continue
		}
		bookmarks = append(bookmarks, types.CompletionItem{
			Label:  p.Name,
			Detail: p.Title,
			Value:  p.UUID,
		})
	}

	var notes []types.CompletionItem
	seen := make(map[string]bool)
	for _, note := range gui.Contexts().Notes.Items {
		if note.Title == "" || seen[note.Title] || bookmarkedUUIDs[note.UUID] {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(note.Title), filter) {
			continue
		}
		seen[note.Title] = true
		notes = append(notes, types.CompletionItem{
			Label:  note.Title,
			Detail: note.ShortDate(),
			Value:  note.UUID,
		})
	}

	var items []types.CompletionItem
	if len(bookmarks) > 0 {
		items = append(items, types.CompletionItem{Label: "Bookmarks", IsHeader: true})
		items = append(items, bookmarks...)
	}
	if len(notes) > 0 {
		items = append(items, types.CompletionItem{Label: "Notes", IsHeader: true})
		items = append(items, notes...)
	}
	return items
}
