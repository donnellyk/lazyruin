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

// OpenSearch opens the search popup with completion.
func (self *SearchHelper) OpenSearch() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}
	cs := types.NewCompletionState()
	cs.FallbackCandidates = gui.AmbientDateCandidates()
	gui.SetSearchCompletion(cs)
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

	query, sort := extractSort(raw)
	opts := gui.BuildSearchOptions()
	opts.Sort = sort
	notes, err := self.c.RuinCmd().Search.Search(query, opts)
	if err != nil {
		gui.ShowError(err)
		return true
	}

	gui.SetSearchQuery(raw)
	gui.SetSearchCompletion(types.NewCompletionState())
	gui.SetCursorEnabled(false)

	gui.PreviewPushNavHistory()
	gui.SetPreviewCards(notes, 0, " Search: "+query+" ")
	gui.ReplaceContextByKey("preview")
	return true
}

// CancelSearch dismisses the search popup.
func (self *SearchHelper) CancelSearch() {
	gui := self.c.GuiCommon()
	gui.SetSearchCompletion(types.NewCompletionState())
	gui.SetCursorEnabled(false)
	gui.PopContext()
}

// ClearSearch clears the active search and returns to the notes panel.
func (self *SearchHelper) ClearSearch() {
	gui := self.c.GuiCommon()
	gui.SetSearchQuery("")
	notesCtx := gui.Contexts().Notes
	notesCtx.CurrentTab = context.NotesTabAll
	self.c.Helpers().Notes().LoadNotesForCurrentTab()
	gui.PushContextByKey("notes")
}

// FocusSearchFilter re-runs the current search and focuses the filter pane.
func (self *SearchHelper) FocusSearchFilter() error {
	gui := self.c.GuiCommon()
	sq := gui.GetSearchQuery()
	if sq != "" {
		query, sort := extractSort(sq)
		opts := gui.BuildSearchOptions()
		opts.Sort = sort
		notes, err := self.c.RuinCmd().Search.Search(query, opts)
		if err == nil {
			gui.SetPreviewCards(notes, 0, " Search: "+sq+" ")
		}
		gui.PushContextByKey("searchFilter")
	}
	return nil
}

// extractSort splits a "sort:value" token out of a query string.
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
