package controllers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// GlobalController handles application-wide keybindings that fire in any view.
type GlobalController struct {
	baseController
	c          *ControllerCommon
	getContext func() *context.GlobalContext

	// Callbacks for actions not yet migrated to helpers.
	onQuit     func() error
	onPick     func() error
	onNewNote  func() error
	onHelp     func() error
	onPalette  func() error
	onCalendar func() error
	onContrib  func() error
}

var _ types.IController = &GlobalController{}

// GlobalControllerOpts holds callbacks injected during wiring.
type GlobalControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.GlobalContext
	// Callbacks for actions not yet migrated to helpers.
	OnQuit     func() error
	OnPick     func() error
	OnNewNote  func() error
	OnHelp     func() error
	OnPalette  func() error
	OnCalendar func() error
	OnContrib  func() error
}

// NewGlobalController creates a GlobalController.
func NewGlobalController(opts GlobalControllerOpts) *GlobalController {
	return &GlobalController{
		c:          opts.Common,
		getContext: opts.GetContext,
		onQuit:     opts.OnQuit,
		onPick:     opts.OnPick,
		onNewNote:  opts.OnNewNote,
		onHelp:     opts.OnHelp,
		onPalette:  opts.OnPalette,
		onCalendar: opts.OnCalendar,
		onContrib:  opts.OnContrib,
	}
}

// Context returns the context this controller is attached to.
func (self *GlobalController) Context() types.Context {
	return self.getContext()
}

// NextPanel implements Tab panel cycling.
func (self *GlobalController) NextPanel() error {
	gc := self.c.GuiCommon()
	order := []types.ContextKey{"notes", "queries", "tags"}
	if gc.SearchQueryActive() {
		order = []types.ContextKey{"searchFilter", "notes", "queries", "tags"}
	}
	currentKey := gc.CurrentContextKey()
	for i, key := range order {
		if key == currentKey {
			nextKey := order[(i+1)%len(order)]
			if ctx := gc.ContextByKey(nextKey); ctx != nil {
				gc.PushContext(ctx, types.OnFocusOpts{})
			} else {
				gc.PushContextByKey(nextKey)
			}
			return nil
		}
	}
	if ctx := gc.ContextByKey("notes"); ctx != nil {
		gc.PushContext(ctx, types.OnFocusOpts{})
	}
	return nil
}

// PrevPanel implements Shift+Tab panel cycling.
func (self *GlobalController) PrevPanel() error {
	gc := self.c.GuiCommon()
	order := []types.ContextKey{"notes", "queries", "tags"}
	if gc.SearchQueryActive() {
		order = []types.ContextKey{"searchFilter", "notes", "queries", "tags"}
	}
	currentKey := gc.CurrentContextKey()
	for i, key := range order {
		if key == currentKey {
			prev := order[(i-1+len(order))%len(order)]
			if ctx := gc.ContextByKey(prev); ctx != nil {
				gc.PushContext(ctx, types.OnFocusOpts{})
			} else {
				gc.PushContextByKey(prev)
			}
			return nil
		}
	}
	if ctx := gc.ContextByKey("notes"); ctx != nil {
		gc.PushContext(ctx, types.OnFocusOpts{})
	}
	return nil
}

// FocusNotes focuses the Notes panel, cycling tabs if already focused.
func (self *GlobalController) FocusNotes() error {
	gc := self.c.GuiCommon()
	if gc.CurrentContextKey() == "notes" {
		self.c.Helpers().Notes().CycleNotesTab()
		return nil
	}
	if ctx := gc.ContextByKey("notes"); ctx != nil {
		gc.PushContext(ctx, types.OnFocusOpts{})
	}
	return nil
}

// FocusQueries focuses the Queries panel, cycling tabs if already focused.
func (self *GlobalController) FocusQueries() error {
	gc := self.c.GuiCommon()
	if gc.CurrentContextKey() == "queries" {
		self.c.Helpers().Queries().CycleQueriesTab()
		return nil
	}
	if ctx := gc.ContextByKey("queries"); ctx != nil {
		gc.PushContext(ctx, types.OnFocusOpts{})
	}
	return nil
}

// FocusTags focuses the Tags panel, cycling tabs if already focused.
func (self *GlobalController) FocusTags() error {
	gc := self.c.GuiCommon()
	if gc.CurrentContextKey() == "tags" {
		self.c.Helpers().Tags().CycleTagsTab()
		return nil
	}
	if ctx := gc.ContextByKey("tags"); ctx != nil {
		gc.PushContext(ctx, types.OnFocusOpts{})
	}
	return nil
}

// FocusPreview focuses the Preview panel.
func (self *GlobalController) FocusPreview() error {
	gc := self.c.GuiCommon()
	if ctx := gc.ContextByKey("preview"); ctx != nil {
		gc.PushContext(ctx, types.OnFocusOpts{})
	}
	return nil
}

func (self *GlobalController) openSearch() error {
	return self.c.Helpers().Search().OpenSearch()
}

func (self *GlobalController) refresh() error {
	self.c.GuiCommon().RefreshAll()
	return nil
}

func (self *GlobalController) focusSearchFilter() error {
	return self.c.Helpers().Search().FocusSearchFilter()
}

// GetKeybindings returns all global keybindings.
func (self *GlobalController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		// Quit (two keys, only first shown in palette)
		{ID: "global.quit", Key: 'q', Handler: self.onQuit, Description: "Quit", Category: "Global"},
		{Key: gocui.KeyCtrlC, Handler: self.onQuit},

		// Core actions
		{ID: "global.search", Key: '/', Handler: self.openSearch, Description: "Search", Category: "Global"},
		{ID: "global.pick", Key: 'p', Handler: self.onPick, Description: "Pick", Category: "Global"},
		{Key: '\\', Handler: self.onPick},
		{ID: "global.new_note", Key: 'n', Handler: self.onNewNote, Description: "New Note", Category: "Global"},
		{ID: "global.refresh", Key: gocui.KeyCtrlR, Handler: self.refresh, Description: "Refresh", Category: "Global"},
		{ID: "global.help", Key: '?', Handler: self.onHelp, Description: "Keybindings", Category: "Global"},
		{ID: "global.palette", Key: ':', Handler: self.onPalette}, // no Description = not in palette
		{ID: "global.calendar", Key: 'c', Handler: self.onCalendar, Description: "Calendar", Category: "Global"},
		{ID: "global.contrib", Key: 'C', Handler: self.onContrib, Description: "Contributions", Category: "Global"},

		// Focus shortcuts
		{ID: "global.focus_notes", Key: '1', Handler: self.FocusNotes, Description: "Focus Notes", Category: "Focus"},
		{ID: "global.focus_queries", Key: '2', Handler: self.FocusQueries, Description: "Focus Queries", Category: "Focus"},
		{ID: "global.focus_tags", Key: '3', Handler: self.FocusTags, Description: "Focus Tags", Category: "Focus"},
		{ID: "global.focus_preview", Handler: self.FocusPreview, Description: "Focus Preview", Category: "Focus"},
		{ID: "global.focus_search_filter", Key: '0', Handler: self.focusSearchFilter, Description: "Focus Search Filter", Category: "Focus"},

		// Panel navigation (no Description = not in palette)
		{Key: gocui.KeyTab, Handler: self.NextPanel},
		{Key: gocui.KeyBacktab, Handler: self.PrevPanel},
	}
}
