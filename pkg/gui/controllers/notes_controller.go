package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	helpers "kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// NotesController handles all Notes panel keybindings and behavior.
type NotesController struct {
	baseController
	*ListControllerTrait[models.Note]
	c          *ControllerCommon
	getContext func() *context.NotesContext

	// Callback for showInfo — delegates to PreviewController (not yet in helpers).
	onShowInfo func(note *models.Note) error
}

var _ types.IController = &NotesController{}

// NotesControllerOpts holds dependencies for construction.
type NotesControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.NotesContext
	// Still uses callback — PreviewController's showInfoDialog is not yet in helpers.
	OnShowInfo func(note *models.Note) error
}

// NewNotesController creates a new NotesController.
func NewNotesController(opts NotesControllerOpts) *NotesController {
	ctrl := &NotesController{
		c:          opts.Common,
		getContext: opts.GetContext,
		onShowInfo: opts.OnShowInfo,
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

func (self *NotesController) h() *helpers.Helpers {
	return self.c.Helpers().(*helpers.Helpers)
}

// GetKeybindingsFn returns the keybinding producer for notes.
func (self *NotesController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		bindings := []*types.Binding{
			// Actions (have Description → shown in palette)
			{
				ID:                "notes.view",
				Key:               gocui.KeyEnter,
				Handler:           self.withItem(self.viewInPreview),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "View in Preview",
				Category:          "Notes",
				DisplayOnScreen:   true,
			},
			{
				ID:                "notes.edit",
				Key:               'E',
				Handler:           self.withItem(self.editNote),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Open in Editor",
				Category:          "Notes",
				DisplayOnScreen:   true,
			},
			{
				ID:                "notes.delete",
				Key:               'd',
				Handler:           self.withItem(self.deleteNote),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Delete Note",
				Category:          "Notes",
			},
			{
				ID:                "notes.copy",
				Key:               'y',
				Handler:           self.withItem(self.copyPath),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Copy Note Path",
				Category:          "Notes",
			},
			// Note actions (also bound on PreviewView)
			{
				ID:                "notes.addTag",
				Key:               't',
				Handler:           self.withItem(self.addTag),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Add Tag",
				Category:          "Note Actions",
			},
			{
				ID:                "notes.removeTag",
				Key:               'T',
				Handler:           self.withItem(self.removeTag),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Remove Tag",
				Category:          "Note Actions",
			},
			{
				ID:                "notes.setParent",
				Key:               '>',
				Handler:           self.withItem(self.setParent),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Set Parent",
				Category:          "Note Actions",
			},
			{
				ID:                "notes.removeParent",
				Key:               'P',
				Handler:           self.withItem(self.removeParent),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Remove Parent",
				Category:          "Note Actions",
			},
			{
				ID:                "notes.bookmark",
				Key:               'b',
				Handler:           self.withItem(self.toggleBookmark),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Toggle Bookmark",
				Category:          "Note Actions",
			},
			{
				ID:                "notes.info",
				Key:               's',
				Handler:           self.withItem(self.showInfo),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Show Info",
				Category:          "Note Actions",
			},
		}
		// Navigation bindings (no Description → excluded from palette)
		bindings = append(bindings, self.NavBindings()...)
		return bindings
	}
}

// GetMouseKeybindingsFn returns mouse bindings for the notes panel.
func (self *NotesController) GetMouseKeybindingsFn() types.MouseKeybindingsFn {
	return func(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
		return []*gocui.ViewMouseBinding{
			{
				ViewName: "notes",
				Key:      gocui.MouseLeft,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					v := self.c.GuiCommon().GetView("notes")
					if v == nil {
						return nil
					}
					ctx := self.getContext()
					idx := helpers.ListClickIndex(v, 3)
					if idx >= 0 && idx < len(ctx.Items) {
						ctx.SetSelectedLineIdx(idx)
					}
					self.c.GuiCommon().PushContext(ctx, types.OnFocusOpts{})
					return nil
				},
			},
			{
				ViewName: "notes",
				Key:      gocui.MouseWheelDown,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					v := self.c.GuiCommon().GetView("notes")
					if v != nil {
						helpers.ScrollViewport(v, 3)
					}
					return nil
				},
			},
			{
				ViewName: "notes",
				Key:      gocui.MouseWheelUp,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					v := self.c.GuiCommon().GetView("notes")
					if v != nil {
						helpers.ScrollViewport(v, -3)
					}
					return nil
				},
			},
		}
	}
}

// Action handlers — call helpers directly.

func (self *NotesController) viewInPreview(note models.Note) error {
	if len(self.getContext().Items) == 0 {
		return nil
	}
	self.c.GuiCommon().PushContextByKey("preview")
	return nil
}

func (self *NotesController) editNote(note models.Note) error {
	return self.h().Editor().OpenInEditor(note.Path)
}

func (self *NotesController) deleteNote(note models.Note) error {
	return self.h().Notes().DeleteNote(&note)
}

func (self *NotesController) copyPath(note models.Note) error {
	return self.h().Clipboard().CopyToClipboard(note.Path)
}

func (self *NotesController) addTag(_ models.Note) error {
	return self.h().NoteActions().AddGlobalTag()
}

func (self *NotesController) removeTag(_ models.Note) error {
	return self.h().NoteActions().RemoveTag()
}

func (self *NotesController) setParent(_ models.Note) error {
	return self.h().NoteActions().SetParentDialog()
}

func (self *NotesController) removeParent(_ models.Note) error {
	return self.h().NoteActions().RemoveParent()
}

func (self *NotesController) toggleBookmark(_ models.Note) error {
	return self.h().NoteActions().ToggleBookmark()
}

func (self *NotesController) showInfo(note models.Note) error {
	return self.onShowInfo(&note)
}
