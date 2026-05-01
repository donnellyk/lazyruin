package controllers

import (
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/helpers"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/jesseduffield/gocui"
)

// NotesController handles all Notes panel keybindings and behavior.
type NotesController struct {
	baseController
	*ListControllerTrait[models.Note]
	NoteActionHandlersTrait
	c              *ControllerCommon
	getContext     func() *context.NotesContext
	getHomeContext func() *context.NotesHomeContext // nil when sections_mode is off

	// Callback for showInfo — delegates to PreviewController (not yet in helpers).
	onShowInfo func(note *models.Note) error
}

var _ types.IController = &NotesController{}

// NotesControllerOpts holds dependencies for construction.
type NotesControllerOpts struct {
	Common         *ControllerCommon
	GetContext     func() *context.NotesContext
	GetHomeContext func() *context.NotesHomeContext // optional — nil when sections_mode is off
	// Still uses callback — PreviewController's showInfoDialog is not yet in helpers.
	OnShowInfo func(note *models.Note) error
}

// NewNotesController creates a new NotesController.
func NewNotesController(opts NotesControllerOpts) *NotesController {
	ctrl := &NotesController{
		NoteActionHandlersTrait: NoteActionHandlersTrait{c: opts.Common},
		c:                       opts.Common,
		getContext:              opts.GetContext,
		getHomeContext:          opts.GetHomeContext,
		onShowInfo:              opts.OnShowInfo,
	}

	ctrl.ListControllerTrait = NewListControllerTrait[models.Note](
		opts.Common,
		func() types.IListContext { return opts.GetContext() },
		func() []models.Note { return opts.GetContext().Items },
		func() *context.ListContextTrait { return opts.GetContext().ListContextTrait },
	)

	return ctrl
}

// Context returns the context this controller is attached to.
func (self *NotesController) Context() types.Context {
	return self.getContext()
}

// homeTabActive reports whether the user is currently on the Home outer
// tab. NotesOuterTab returns "" when sections_mode is disabled, so this
// stays false in legacy mode without checking the config explicitly.
func (self *NotesController) homeTabActive() bool {
	return self.c.GuiCommon().NotesOuterTab() == "home"
}

// homeCtxOrNil returns the NotesHomeContext when sections_mode is on, or
// nil otherwise. Helpers in this controller defensively guard with this
// since they're only invoked from dispatch paths that already gated on
// homeTabActive.
func (self *NotesController) homeCtxOrNil() *context.NotesHomeContext {
	if self.getHomeContext == nil {
		return nil
	}
	return self.getHomeContext()
}

// disableOnHomeTab returns a DisabledReason fn that disables a binding
// whenever the Home outer tab is active. Used for note-action bindings
// that don't apply to section items.
func (self *NotesController) disableOnHomeTab() func() *types.DisabledReason {
	return func() *types.DisabledReason {
		if self.homeTabActive() {
			return &types.DisabledReason{Text: "Not available on Home tab"}
		}
		return nil
	}
}

// GetKeybindings returns the keybindings for notes. In sections_mode the
// Notes pane is shared between the flat-list outer tab and the Home outer
// tab; bindings for note-actions get a "disabled on Home" reason wrapper,
// and j/k/Enter dispatch by outer tab so the same keys steer rows on Home
// and notes on the flat list.
func (self *NotesController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	noteAction := self.disableOnHomeTab() // shared GetDisabledReason wrapper

	bindings := []*types.Binding{
		// Actions (have Description → shown in palette)
		{
			ID:                "notes.view",
			Key:               gocui.KeyEnter,
			Handler:           self.dispatchEnter,
			GetDisabledReason: self.requireForEnter(),
			Description:       "View in Preview",
			Category:          "Notes",
			DisplayOnScreen:   true,
			StatusBarLabel:    "View",
		},
		{
			ID:                "notes.edit",
			Key:               'E',
			Handler:           self.withItem(self.editNote),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Open in Editor",
			Category:          "Notes",
			DisplayOnScreen:   true,
			StatusBarLabel:    "Editor",
		},
		{
			ID:                "notes.edit_inline",
			Key:               'e',
			Handler:           self.withItem(self.editInline),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Edit in Popup",
			Category:          "Notes",
		},
		{
			ID:                "notes.delete",
			Key:               'd',
			Handler:           self.withItem(self.deleteNote),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Delete Note",
			Category:          "Notes",
			DisplayOnScreen:   true,
			StatusBarLabel:    "Delete",
		},
		{
			ID:                "notes.copy",
			Key:               'y',
			Handler:           self.withItem(self.copyPath),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Copy Note Path",
			Category:          "Notes",
		},
		{
			ID:                "notes.addTag",
			Key:               't',
			Handler:           self.withItem(func(_ models.Note) error { return self.addTag() }),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Add Tag",
			Category:          "Note Actions",
			DisplayOnScreen:   true,
			StatusBarLabel:    "Tag",
		},
		{
			ID:                "notes.removeTag",
			Key:               'T',
			Handler:           self.withItem(func(_ models.Note) error { return self.removeTag() }),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Remove Tag",
			Category:          "Note Actions",
		},
		{
			ID:                "notes.setParent",
			Key:               '>',
			Handler:           self.withItem(func(_ models.Note) error { return self.setParent() }),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Set Parent",
			Category:          "Note Actions",
		},
		{
			ID:                "notes.removeParent",
			Key:               'P',
			Handler:           self.withItem(func(_ models.Note) error { return self.removeParent() }),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Remove Parent",
			Category:          "Note Actions",
		},
		{
			ID:                "notes.bookmark",
			Key:               'b',
			Handler:           self.withItem(func(_ models.Note) error { return self.toggleBookmark() }),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Toggle Bookmark",
			Category:          "Note Actions",
			DisplayOnScreen:   true,
			StatusBarLabel:    "Bookmark",
		},
		{
			ID:                "notes.info",
			Key:               's',
			Handler:           self.withItem(self.showInfo),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected()),
			Description:       "Show Info",
			Category:          "Note Actions",
		},
		{
			ID:                "notes.openURL",
			Key:               'o',
			Handler:           self.withItem(self.openURL),
			GetDisabledReason: self.require(noteAction, self.singleItemSelected(), self.isLinkNote()),
			Description:       "Open URL",
			Category:          "Notes",
			DisplayOnScreen:   true,
			StatusBarLabel:    "Open",
		},
	}
	// Navigation bindings (no Description → excluded from palette).
	// In sections_mode, j/k/arrows dispatch on outer tab; legacy mode
	// uses the existing list-trait navigation directly.
	bindings = append(bindings, self.navBindings()...)
	return bindings
}

// navBindings returns the j/k/g/G/arrow nav bindings, dispatching to the
// Home tab cursor when sections_mode is on and Home is active.
func (self *NotesController) navBindings() []*types.Binding {
	return []*types.Binding{
		{Key: 'j', Handler: self.dispatchNext, KeyDisplay: "j/k", Description: "Move down/up", Category: "Navigation"},
		{Key: 'k', Handler: self.dispatchPrev},
		{Key: 'g', Handler: self.dispatchTop, KeyDisplay: "g/G", Description: "Go to top/bottom", Category: "Navigation"},
		{Key: 'G', Handler: self.dispatchBottom},
		{Key: gocui.KeyArrowDown, Handler: self.dispatchNext},
		{Key: gocui.KeyArrowUp, Handler: self.dispatchPrev},
	}
}

// dispatchEnter routes Enter to the Home item activator when the Home
// outer tab is active, and to the flat-list "view in preview" handler
// otherwise.
func (self *NotesController) dispatchEnter() error {
	if self.homeTabActive() {
		homeCtx := self.homeCtxOrNil()
		if homeCtx == nil {
			return nil
		}
		row := homeCtx.Selected()
		if row == nil {
			return nil
		}
		return self.c.Helpers().NotesHome().Activate(*row)
	}
	return self.withItem(self.viewInPreview)()
}

// requireForEnter returns the GetDisabledReason for Enter. On Home, only
// disabled when no item is selectable. On Notes, requires a single item.
func (self *NotesController) requireForEnter() func() *types.DisabledReason {
	flatRequire := self.require(self.singleItemSelected())
	return func() *types.DisabledReason {
		if self.homeTabActive() {
			homeCtx := self.homeCtxOrNil()
			if homeCtx == nil || homeCtx.Selected() == nil {
				return &types.DisabledReason{Text: "No item selected"}
			}
			return nil
		}
		return flatRequire()
	}
}

func (self *NotesController) dispatchNext() error {
	if self.homeTabActive() {
		if ctx := self.homeCtxOrNil(); ctx != nil {
			ctx.SelectedIdx = ctx.NextSelectable(ctx.SelectedIdx)
			self.c.GuiCommon().RenderNotes()
			self.hoverSelected()
		}
		return nil
	}
	return self.ListControllerTrait.nextItem()
}

func (self *NotesController) dispatchPrev() error {
	if self.homeTabActive() {
		if ctx := self.homeCtxOrNil(); ctx != nil {
			ctx.SelectedIdx = ctx.PrevSelectable(ctx.SelectedIdx)
			self.c.GuiCommon().RenderNotes()
			self.hoverSelected()
		}
		return nil
	}
	return self.ListControllerTrait.prevItem()
}

func (self *NotesController) dispatchTop() error {
	if self.homeTabActive() {
		if ctx := self.homeCtxOrNil(); ctx != nil {
			if first := ctx.FirstSelectableIdx(); first >= 0 {
				ctx.SelectedIdx = first
			}
			self.c.GuiCommon().RenderNotes()
			self.hoverSelected()
		}
		return nil
	}
	return self.ListControllerTrait.goTop()
}

func (self *NotesController) dispatchBottom() error {
	if self.homeTabActive() {
		if ctx := self.homeCtxOrNil(); ctx != nil {
			for i := len(ctx.Rows) - 1; i >= 0; i-- {
				r := ctx.Rows[i]
				if !r.IsHeader && !r.Blank {
					ctx.SelectedIdx = i
					break
				}
			}
			self.c.GuiCommon().RenderNotes()
			self.hoverSelected()
		}
		return nil
	}
	return self.ListControllerTrait.goBottom()
}

// hoverSelected runs the currently selected Home row through the hover
// preview path (no nav-history entry). Caller has already moved the
// cursor and re-rendered.
func (self *NotesController) hoverSelected() {
	ctx := self.homeCtxOrNil()
	if ctx == nil {
		return
	}
	row := ctx.Selected()
	if row == nil {
		return
	}
	self.c.Helpers().NotesHome().Hover(*row)
}

// GetMouseKeybindings returns mouse bindings for the notes panel.
//
// We build these manually rather than reusing ListMouseBindings because
// gocui dispatches the FIRST matching binding in g.keybindings (see
// gui.go execKeybindings), so layering an override on top of
// ListMouseBindings would never fire — the layered binding sits later
// in the slice. Building the bindings up front lets the click and
// wheel handlers dispatch correctly in both outer-tab modes:
//
//   - flat-list (Notes): click maps screen-row → note-index using
//     itemHeight=3 (each note is 3 rendered lines); wheel free-scrolls
//     the view origin.
//   - Home: click maps screen-row → row-index using itemHeight=1
//     (each row is 1 rendered line); wheel steps the cursor (with
//     re-hover) because renderNotesHome's scrollListView would snap a
//     free-scroll back to the cursor on the next render.
func (self *NotesController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	gc := func() IGuiCommon { return self.c.GuiCommon() }
	view := func() *gocui.View { return gc().GetView("notes") }

	return []*gocui.ViewMouseBinding{
		{
			ViewName: "notes",
			Key:      gocui.MouseLeft,
			Handler: func(_ gocui.ViewMouseBindingOpts) error {
				v := view()
				if v == nil {
					return nil
				}
				if self.homeTabActive() {
					ctx := self.homeCtxOrNil()
					if ctx == nil {
						return nil
					}
					// Any click inside the Home pane focuses it; only
					// row-level effects (cursor move + hover) require
					// hitting a selectable row.
					gc().PushContext(ctx, types.OnFocusOpts{})
					idx := helpers.ListClickIndex(v, 1)
					if idx < 0 || idx >= len(ctx.Rows) {
						return nil
					}
					r := ctx.Rows[idx]
					if r.IsHeader || r.Blank {
						return nil
					}
					ctx.SelectedIdx = idx
					gc().RenderNotes()
					self.hoverSelected()
					return nil
				}
				idx := helpers.ListClickIndex(v, 3)
				notesCtx := self.getContext()
				if idx >= 0 && idx < len(notesCtx.Items) {
					notesCtx.SetSelectedLineIdx(idx)
				}
				gc().PushContext(notesCtx, types.OnFocusOpts{})
				return nil
			},
		},
		{
			ViewName: "notes",
			Key:      gocui.MouseWheelDown,
			Handler: func(_ gocui.ViewMouseBindingOpts) error {
				if self.homeTabActive() {
					return self.dispatchNext()
				}
				if v := view(); v != nil {
					helpers.ScrollViewport(v, 3)
				}
				return nil
			},
		},
		{
			ViewName: "notes",
			Key:      gocui.MouseWheelUp,
			Handler: func(_ gocui.ViewMouseBindingOpts) error {
				if self.homeTabActive() {
					return self.dispatchPrev()
				}
				if v := view(); v != nil {
					helpers.ScrollViewport(v, -3)
				}
				return nil
			},
		},
	}
}

// Action handlers — call helpers directly.

func (self *NotesController) viewInPreview(note models.Note) error {
	if len(self.getContext().Items) == 0 {
		return nil
	}
	return self.c.Helpers().Navigator().NavigateTo("cardList", note.Title, func() error {
		source := self.c.Helpers().Preview().NewSingleNoteSource(note.UUID)
		self.c.Helpers().Preview().ShowCardList(note.Title, []models.Note{note}, source)
		return nil
	})
}

func (self *NotesController) editNote(note models.Note) error {
	return self.c.Helpers().Editor().OpenInEditor(note.Path)
}

func (self *NotesController) editInline(note models.Note) error {
	return self.c.Helpers().Capture().OpenCaptureForEdit(&note)
}

func (self *NotesController) deleteNote(note models.Note) error {
	return self.c.Helpers().Notes().DeleteNote(&note)
}

func (self *NotesController) copyPath(note models.Note) error {
	return self.c.Helpers().Clipboard().CopyToClipboard(note.Path)
}

func (self *NotesController) showInfo(note models.Note) error {
	return self.onShowInfo(&note)
}

func (self *NotesController) openURL(note models.Note) error {
	return self.c.Helpers().Link().OpenLinkURL(&note)
}

func (self *NotesController) isLinkNote() func() *types.DisabledReason {
	return requireLinkNote(func() *models.Note { return self.getContext().Selected() })
}
