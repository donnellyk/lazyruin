package helpers

import (
	"strings"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// SearchHelper manages search execution and query management.
type SearchHelper struct {
	c *HelperCommon
}

// NewSearchHelper creates a new SearchHelper.
func NewSearchHelper(c *HelperCommon) *SearchHelper {
	return &SearchHelper{c: c}
}

func (self *SearchHelper) searchCtx() *context.SearchContext {
	return self.c.GuiCommon().Contexts().Search
}

// OpenSearch opens the search popup with completion.
func (self *SearchHelper) OpenSearch() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}
	cs := types.NewCompletionState()
	cs.FallbackCandidates = AmbientDateCandidates()
	self.searchCtx().Completion = cs
	gui.PushContextByKey("search")
	return nil
}

// ExecuteSearch runs the search and displays results in the preview pane.
// Returns true if the search was executed, false if the input was empty (caller should cancel).
func (self *SearchHelper) ExecuteSearch(raw string) (executed bool) {
	gui := self.c.GuiCommon()
	if raw == "" {
		return false
	}

	query, sort := ExtractSort(raw)
	opts := self.c.Helpers().Preview().BuildSearchOptions()
	opts.Sort = sort
	notes, err := self.c.RuinCmd().Search.Search(query, opts)
	if err != nil {
		gui.ShowError(err)
		return true
	}

	sc := self.searchCtx()
	sc.Query = raw
	sc.Completion = types.NewCompletionState()
	gui.SetCursorEnabled(false)

	source := self.c.Helpers().Preview().NewSearchSourceWithExtractSort(raw)

	self.c.Helpers().PreviewNav().PushNavHistory()
	self.c.Helpers().Preview().ShowCardList("Search: "+query, notes, source)
	gui.ReplaceContextByKey("cardList")
	return true
}

// OpenSearchAsFilter opens the search dialog in filter mode with a custom
// title, pre-filled seed text, source-specific completion triggers, and a
// submit callback that replaces the normal search behavior.
func (self *SearchHelper) OpenSearchAsFilter(title, seed string, triggers func() []types.CompletionTrigger, onSubmit func(string) error) {
	ctx := self.searchCtx()
	ctx.FilterTitle = title
	ctx.FilterSeed = seed
	ctx.FilterTriggers = triggers
	ctx.OnFilterSubmit = onSubmit
	cs := types.NewCompletionState()
	cs.FallbackCandidates = AmbientDateCandidates()
	ctx.Completion = cs
	self.c.GuiCommon().PushContextByKey("search")
}

// CancelSearch dismisses the search popup.
func (self *SearchHelper) CancelSearch() {
	gui := self.c.GuiCommon()
	self.searchCtx().ClearFilterMode()
	self.searchCtx().Completion = types.NewCompletionState()
	gui.SetCursorEnabled(false)
	gui.PopContext()
}

// ClearSearch clears the active search and returns to the notes panel.
func (self *SearchHelper) ClearSearch() {
	gui := self.c.GuiCommon()
	self.searchCtx().Query = ""
	notesCtx := gui.Contexts().Notes
	notesCtx.CurrentTab = context.NotesTabAll
	self.c.Helpers().Notes().LoadNotesForCurrentTab()
	gui.PushContextByKey("notes")
}

// FocusSearchFilter re-runs the current search and focuses the filter pane.
func (self *SearchHelper) FocusSearchFilter() error {
	gui := self.c.GuiCommon()
	sq := self.searchCtx().Query
	if sq != "" {
		query, sort := ExtractSort(sq)
		opts := self.c.Helpers().Preview().BuildSearchOptions()
		opts.Sort = sort
		notes, err := self.c.RuinCmd().Search.Search(query, opts)
		if err == nil {
			source := self.c.Helpers().Preview().NewSearchSourceWithExtractSort(sq)
			self.c.Helpers().Preview().ShowCardList("Search: "+sq, notes, source)
		}
		gui.PushContextByKey("searchFilter")
	}
	return nil
}

// ExtractSort splits a "sort:value" token out of a query string.
func ExtractSort(query string) (string, string) {
	var remaining []string
	var sortVal string
	for token := range strings.FieldsSeq(query) {
		if v, ok := strings.CutPrefix(token, "sort:"); ok {
			sortVal = v
		} else {
			remaining = append(remaining, token)
		}
	}
	return strings.Join(remaining, " "), sortVal
}
