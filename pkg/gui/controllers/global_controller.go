package controllers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// GlobalController handles application-wide keybindings that fire in any view.
type GlobalController struct {
	baseController
	getContext func() *context.GlobalContext

	onQuit              func() error
	onSearch            func() error
	onPick              func() error
	onNewNote           func() error
	onRefresh           func() error
	onHelp              func() error
	onPalette           func() error
	onCalendar          func() error
	onContrib           func() error
	onFocusNotes        func() error
	onFocusQueries      func() error
	onFocusTags         func() error
	onFocusPreview      func() error
	onFocusSearchFilter func() error
	onNextPanel         func() error
	onPrevPanel         func() error
}

var _ types.IController = &GlobalController{}

// GlobalControllerOpts holds callbacks injected during wiring.
type GlobalControllerOpts struct {
	GetContext          func() *context.GlobalContext
	OnQuit              func() error
	OnSearch            func() error
	OnPick              func() error
	OnNewNote           func() error
	OnRefresh           func() error
	OnHelp              func() error
	OnPalette           func() error
	OnCalendar          func() error
	OnContrib           func() error
	OnFocusNotes        func() error
	OnFocusQueries      func() error
	OnFocusTags         func() error
	OnFocusPreview      func() error
	OnFocusSearchFilter func() error
	OnNextPanel         func() error
	OnPrevPanel         func() error
}

// NewGlobalController creates a GlobalController.
func NewGlobalController(opts GlobalControllerOpts) *GlobalController {
	return &GlobalController{
		getContext:          opts.GetContext,
		onQuit:              opts.OnQuit,
		onSearch:            opts.OnSearch,
		onPick:              opts.OnPick,
		onNewNote:           opts.OnNewNote,
		onRefresh:           opts.OnRefresh,
		onHelp:              opts.OnHelp,
		onPalette:           opts.OnPalette,
		onCalendar:          opts.OnCalendar,
		onContrib:           opts.OnContrib,
		onFocusNotes:        opts.OnFocusNotes,
		onFocusQueries:      opts.OnFocusQueries,
		onFocusTags:         opts.OnFocusTags,
		onFocusPreview:      opts.OnFocusPreview,
		onFocusSearchFilter: opts.OnFocusSearchFilter,
		onNextPanel:         opts.OnNextPanel,
		onPrevPanel:         opts.OnPrevPanel,
	}
}

// Context returns the context this controller is attached to.
func (self *GlobalController) Context() types.Context {
	return self.getContext()
}

// GetKeybindingsFn returns all global keybinding producers.
func (self *GlobalController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			// Quit (two keys, only first shown in palette)
			{ID: "global.quit", Key: 'q', Handler: self.onQuit, Description: "Quit", Category: "Global"},
			{Key: gocui.KeyCtrlC, Handler: self.onQuit},

			// Core actions
			{ID: "global.search", Key: '/', Handler: self.onSearch, Description: "Search", Category: "Global"},
			{ID: "global.pick", Key: 'p', Handler: self.onPick, Description: "Pick", Category: "Global"},
			{Key: '\\', Handler: self.onPick},
			{ID: "global.new_note", Key: 'n', Handler: self.onNewNote, Description: "New Note", Category: "Global"},
			{ID: "global.refresh", Key: gocui.KeyCtrlR, Handler: self.onRefresh, Description: "Refresh", Category: "Global"},
			{ID: "global.help", Key: '?', Handler: self.onHelp, Description: "Keybindings", Category: "Global"},
			{ID: "global.palette", Key: ':', Handler: self.onPalette}, // no Description = not in palette
			{ID: "global.calendar", Key: 'c', Handler: self.onCalendar, Description: "Calendar", Category: "Global"},
			{ID: "global.contrib", Key: 'C', Handler: self.onContrib, Description: "Contributions", Category: "Global"},

			// Focus shortcuts
			{ID: "global.focus_notes", Key: '1', Handler: self.onFocusNotes, Description: "Focus Notes", Category: "Focus"},
			{ID: "global.focus_queries", Key: '2', Handler: self.onFocusQueries, Description: "Focus Queries", Category: "Focus"},
			{ID: "global.focus_tags", Key: '3', Handler: self.onFocusTags, Description: "Focus Tags", Category: "Focus"},
			{ID: "global.focus_preview", Handler: self.onFocusPreview, Description: "Focus Preview", Category: "Focus"},
			{ID: "global.focus_search_filter", Key: '0', Handler: self.onFocusSearchFilter, Description: "Focus Search Filter", Category: "Focus"},

			// Panel navigation (no Description = not in palette)
			{Key: gocui.KeyTab, Handler: self.onNextPanel},
			{Key: gocui.KeyBacktab, Handler: self.onPrevPanel},
		}
	}
}
