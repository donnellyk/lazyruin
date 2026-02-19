package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"
)

// PreviewController handles all Preview panel keybindings.
// Action methods delegate to specialized preview helpers.
type PreviewController struct {
	baseController
	c          *ControllerCommon
	getContext func() *context.PreviewContext
}

var _ types.IController = &PreviewController{}

// NewPreviewController creates a new PreviewController.
func NewPreviewController(c *ControllerCommon, getContext func() *context.PreviewContext) *PreviewController {
	return &PreviewController{
		c:          c,
		getContext: getContext,
	}
}

// Context returns the context this controller is attached to.
func (self *PreviewController) Context() types.Context {
	return self.getContext()
}

// Helper accessors.
func (self *PreviewController) preview() *helpers.PreviewHelper { return self.c.Helpers().Preview() }
func (self *PreviewController) nav() *helpers.PreviewNavHelper  { return self.c.Helpers().PreviewNav() }
func (self *PreviewController) links() *helpers.PreviewLinksHelper {
	return self.c.Helpers().PreviewLinks()
}
func (self *PreviewController) mutations() *helpers.PreviewMutationsHelper {
	return self.c.Helpers().PreviewMutations()
}
func (self *PreviewController) lineOps() *helpers.PreviewLineOpsHelper {
	return self.c.Helpers().PreviewLineOps()
}
func (self *PreviewController) info() *helpers.PreviewInfoHelper {
	return self.c.Helpers().PreviewInfo()
}

// Note action wrappers — call helpers directly.
func (self *PreviewController) addTag() error    { return self.c.Helpers().NoteActions().AddGlobalTag() }
func (self *PreviewController) removeTag() error { return self.c.Helpers().NoteActions().RemoveTag() }
func (self *PreviewController) setParent() error {
	return self.c.Helpers().NoteActions().SetParentDialog()
}
func (self *PreviewController) removeParent() error {
	return self.c.Helpers().NoteActions().RemoveParent()
}
func (self *PreviewController) toggleBookmark() error {
	return self.c.Helpers().NoteActions().ToggleBookmark()
}

// GetKeybindings returns the keybindings for preview.
// Bindings with Key == nil are palette-only (no keybinding registered).
func (self *PreviewController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		// Navigation (no Description -> excluded from palette)
		{Key: 'j', Handler: self.nav().MoveDown},
		{Key: gocui.KeyArrowDown, Handler: self.nav().MoveDown},
		{Key: 'k', Handler: self.nav().MoveUp},
		{Key: gocui.KeyArrowUp, Handler: self.nav().MoveUp},
		{Key: 'J', Handler: self.nav().CardDown},
		{Key: 'K', Handler: self.nav().CardUp},
		{Key: '}', Handler: self.nav().NextHeader},
		{Key: '{', Handler: self.nav().PrevHeader},
		{Key: 'l', Handler: self.links().HighlightNextLink},
		{Key: 'L', Handler: self.links().HighlightPrevLink},

		// Card actions (Description -> shown in palette)
		{
			ID:          "preview.delete",
			Key:         'd',
			Handler:     self.mutations().DeleteCard,
			Description: "Delete Card",
			Category:    "Preview",
		},
		{
			ID:              "preview.open_editor",
			Key:             'E',
			Handler:         self.nav().OpenInEditor,
			Description:     "Open in Editor",
			Category:        "Preview",
			DisplayOnScreen: true,
		},
		{
			ID:          "preview.append_done",
			Key:         'D',
			Handler:     self.lineOps().AppendDone,
			Description: "Toggle #done",
			Category:    "Preview",
		},
		{
			ID:          "preview.move_card",
			Key:         'm',
			Handler:     self.mutations().MoveCardDialog,
			Description: "Move Card",
			Category:    "Preview",
		},
		{
			ID:          "preview.merge_card",
			Key:         'M',
			Handler:     self.mutations().MergeCardDialog,
			Description: "Merge Notes",
			Category:    "Preview",
		},
		{
			ID:          "preview.toggle_frontmatter",
			Key:         'f',
			Handler:     self.preview().ToggleFrontmatter,
			Description: "Toggle Frontmatter",
			Category:    "Preview",
		},
		{
			ID:          "preview.view_options",
			Key:         'v',
			Handler:     self.preview().ViewOptionsDialog,
			Description: "View Options",
			Category:    "Preview",
		},
		{
			ID:          "preview.toggle_inline_tag",
			Key:         gocui.KeyCtrlT,
			Handler:     self.lineOps().ToggleInlineTag,
			Description: "Toggle Inline Tag",
			Category:    "Preview",
		},
		{
			ID:          "preview.toggle_inline_date",
			Key:         gocui.KeyCtrlD,
			Handler:     self.lineOps().ToggleInlineDate,
			Description: "Toggle Inline Date",
			Category:    "Preview",
		},
		{
			ID:          "preview.open_link",
			Key:         'o',
			Handler:     self.links().OpenLink,
			Description: "Open Link",
			Category:    "Preview",
		},
		{
			ID:          "preview.toggle_todo",
			Key:         'x',
			Handler:     self.lineOps().ToggleTodo,
			Description: "Toggle Todo",
			Category:    "Preview",
		},
		{
			ID:              "preview.focus_note",
			Key:             gocui.KeyEnter,
			Handler:         self.nav().FocusNote,
			Description:     "Focus Note from Preview",
			Category:        "Preview",
			DisplayOnScreen: true,
		},
		// Back (no Description -> excluded from palette)
		{
			ID:      "preview.back",
			Key:     gocui.KeyEsc,
			Handler: self.nav().Back,
		},
		{
			ID:          "preview.nav_back",
			Key:         '[',
			Handler:     self.nav().NavBack,
			Description: "Go Back",
			Category:    "Preview",
		},
		{
			ID:          "preview.nav_forward",
			Key:         ']',
			Handler:     self.nav().NavForward,
			Description: "Go Forward",
			Category:    "Preview",
		},

		// Note actions (shared with Notes panel) -- call helpers directly
		{
			ID:              "preview.add_tag",
			Key:             't',
			Handler:         self.addTag,
			Description:     "Add Tag",
			Category:        "Note Actions",
			DisplayOnScreen: true,
		},
		{
			ID:          "preview.remove_tag",
			Key:         'T',
			Handler:     self.removeTag,
			Description: "Remove Tag",
			Category:    "Note Actions",
		},
		{
			ID:          "preview.set_parent",
			Key:         '>',
			Handler:     self.setParent,
			Description: "Set Parent",
			Category:    "Note Actions",
		},
		{
			ID:          "preview.remove_parent",
			Key:         'P',
			Handler:     self.removeParent,
			Description: "Remove Parent",
			Category:    "Note Actions",
		},
		{
			ID:          "preview.toggle_bookmark",
			Key:         'b',
			Handler:     self.toggleBookmark,
			Description: "Toggle Bookmark",
			Category:    "Note Actions",
		},
		{
			ID:          "preview.show_info",
			Key:         's',
			Handler:     self.info().ShowInfoDialog,
			Description: "Show Info",
			Category:    "Note Actions",
		},

		// Palette-only commands (Key == nil -> no keybinding, palette entry only)
		{
			ID:          "preview.toggle_title",
			Handler:     self.preview().ToggleTitle,
			Description: "Toggle Title",
			Category:    "Preview",
		},
		{
			ID:          "preview.toggle_global_tags",
			Handler:     self.preview().ToggleGlobalTags,
			Description: "Toggle Global Tags",
			Category:    "Preview",
		},
		{
			ID:          "preview.toggle_markdown",
			Handler:     self.preview().ToggleMarkdown,
			Description: "Toggle Markdown",
			Category:    "Preview",
		},
		{
			ID:          "preview.order_cards",
			Handler:     self.mutations().OrderCards,
			Description: "Order Cards",
			Category:    "Preview",
		},
		{
			ID:          "preview.show_nav_history",
			Handler:     self.nav().ShowNavHistory,
			Description: "View History",
			Category:    "Preview",
		},
	}
}

// GetMouseKeybindings returns mouse bindings for the preview panel.
func (self *PreviewController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: "preview",
			Key:      gocui.MouseLeft,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return self.nav().Click()
			},
		},
		{
			ViewName: "",
			Key:      gocui.MouseWheelDown,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return self.nav().ScrollDown()
			},
		},
		{
			ViewName: "",
			Key:      gocui.MouseWheelUp,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return self.nav().ScrollUp()
			},
		},
	}
}
