package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// NotesController handles all Notes panel keybindings and behavior.
type NotesController struct {
	baseController
	*ListControllerTrait[models.Note]
	c          *ControllerCommon
	getContext func() *context.NotesContext

	// Action callbacks injected by gui wiring during the hybrid migration period.
	onViewInPreview  func(note *models.Note) error
	onEditNote       func(note *models.Note) error
	onDeleteNote     func(note *models.Note) error
	onCopyPath       func(note *models.Note) error
	onAddTag         func(note *models.Note) error
	onRemoveTag      func(note *models.Note) error
	onSetParent      func(note *models.Note) error
	onRemoveParent   func(note *models.Note) error
	onToggleBookmark func(note *models.Note) error
	onShowInfo       func(note *models.Note) error
	onClickFn        func(g *gocui.Gui, v *gocui.View) error
	onWheelDown      func(g *gocui.Gui, v *gocui.View) error
	onWheelUp        func(g *gocui.Gui, v *gocui.View) error
}

var _ types.IController = &NotesController{}

// NotesControllerOpts holds the callbacks injected during wiring.
type NotesControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.NotesContext
	// Action callbacks — delegate back to gui methods during hybrid period
	OnViewInPreview  func(note *models.Note) error
	OnEditNote       func(note *models.Note) error
	OnDeleteNote     func(note *models.Note) error
	OnCopyPath       func(note *models.Note) error
	OnAddTag         func(note *models.Note) error
	OnRemoveTag      func(note *models.Note) error
	OnSetParent      func(note *models.Note) error
	OnRemoveParent   func(note *models.Note) error
	OnToggleBookmark func(note *models.Note) error
	OnShowInfo       func(note *models.Note) error
	OnClick          func(g *gocui.Gui, v *gocui.View) error
	OnWheelDown      func(g *gocui.Gui, v *gocui.View) error
	OnWheelUp        func(g *gocui.Gui, v *gocui.View) error
}

// NewNotesController creates a new NotesController.
func NewNotesController(opts NotesControllerOpts) *NotesController {
	ctrl := &NotesController{
		c:                opts.Common,
		getContext:       opts.GetContext,
		onViewInPreview:  opts.OnViewInPreview,
		onEditNote:       opts.OnEditNote,
		onDeleteNote:     opts.OnDeleteNote,
		onCopyPath:       opts.OnCopyPath,
		onAddTag:         opts.OnAddTag,
		onRemoveTag:      opts.OnRemoveTag,
		onSetParent:      opts.OnSetParent,
		onRemoveParent:   opts.OnRemoveParent,
		onToggleBookmark: opts.OnToggleBookmark,
		onShowInfo:       opts.OnShowInfo,
		onClickFn:        opts.OnClick,
		onWheelDown:      opts.OnWheelDown,
		onWheelUp:        opts.OnWheelUp,
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
					return self.onClickFn(nil, nil)
				},
			},
			{
				ViewName: "notes",
				Key:      gocui.MouseWheelDown,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					return self.onWheelDown(nil, nil)
				},
			},
			{
				ViewName: "notes",
				Key:      gocui.MouseWheelUp,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					return self.onWheelUp(nil, nil)
				},
			},
		}
	}
}

// Action handlers — delegate to injected callbacks.

func (self *NotesController) viewInPreview(note models.Note) error {
	return self.onViewInPreview(&note)
}

func (self *NotesController) editNote(note models.Note) error {
	return self.onEditNote(&note)
}

func (self *NotesController) deleteNote(note models.Note) error {
	return self.onDeleteNote(&note)
}

func (self *NotesController) copyPath(note models.Note) error {
	return self.onCopyPath(&note)
}

func (self *NotesController) addTag(note models.Note) error {
	return self.onAddTag(&note)
}

func (self *NotesController) removeTag(note models.Note) error {
	return self.onRemoveTag(&note)
}

func (self *NotesController) setParent(note models.Note) error {
	return self.onSetParent(&note)
}

func (self *NotesController) removeParent(note models.Note) error {
	return self.onRemoveParent(&note)
}

func (self *NotesController) toggleBookmark(note models.Note) error {
	return self.onToggleBookmark(&note)
}

func (self *NotesController) showInfo(note models.Note) error {
	return self.onShowInfo(&note)
}
