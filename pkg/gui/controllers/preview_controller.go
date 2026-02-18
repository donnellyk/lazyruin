package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"
)

// PreviewController handles all Preview panel keybindings.
// During the hybrid migration period, action callbacks delegate to existing
// gui.preview.* methods. State remains in GuiState.Preview.
type PreviewController struct {
	baseController
	c          *ControllerCommon
	getContext func() *context.PreviewContext

	// Navigation
	onMoveDown   func() error
	onMoveUp     func() error
	onCardDown   func() error
	onCardUp     func() error
	onNextHeader func() error
	onPrevHeader func() error
	onNextLink   func() error
	onPrevLink   func() error
	onClick      func() error

	// Card actions
	onDeleteCard        func() error
	onOpenInEditor      func() error
	onAppendDone        func() error
	onMoveCard          func() error
	onMergeCard         func() error
	onToggleFrontmatter func() error
	onViewOptions       func() error
	onToggleInlineTag   func() error
	onToggleInlineDate  func() error
	onOpenLink          func() error
	onToggleTodo        func() error
	onFocusNote         func() error
	onBack              func() error
	onNavBack           func() error
	onNavForward        func() error

	// Note actions — showInfo still delegates to gui.preview (not yet in helpers)
	onShowInfo func() error

	// Palette-only (Key: nil = no keybinding, just palette entry)
	onToggleTitle      func() error
	onToggleGlobalTags func() error
	onToggleMarkdown   func() error
	onOrderCards       func() error
	onShowNavHistory   func() error
}

var _ types.IController = &PreviewController{}

// PreviewControllerOpts holds the callbacks injected during wiring.
type PreviewControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.PreviewContext

	// Navigation
	OnMoveDown   func() error
	OnMoveUp     func() error
	OnCardDown   func() error
	OnCardUp     func() error
	OnNextHeader func() error
	OnPrevHeader func() error
	OnNextLink   func() error
	OnPrevLink   func() error
	OnClick      func() error

	// Card actions
	OnDeleteCard        func() error
	OnOpenInEditor      func() error
	OnAppendDone        func() error
	OnMoveCard          func() error
	OnMergeCard         func() error
	OnToggleFrontmatter func() error
	OnViewOptions       func() error
	OnToggleInlineTag   func() error
	OnToggleInlineDate  func() error
	OnOpenLink          func() error
	OnToggleTodo        func() error
	OnFocusNote         func() error
	OnBack              func() error
	OnNavBack           func() error
	OnNavForward        func() error

	// Note actions — showInfo still delegates to gui.preview (not yet in helpers)
	OnShowInfo func() error

	// Palette-only
	OnToggleTitle      func() error
	OnToggleGlobalTags func() error
	OnToggleMarkdown   func() error
	OnOrderCards       func() error
	OnShowNavHistory   func() error
}

// NewPreviewController creates a new PreviewController.
func NewPreviewController(opts PreviewControllerOpts) *PreviewController {
	return &PreviewController{
		c:                   opts.Common,
		getContext:          opts.GetContext,
		onMoveDown:          opts.OnMoveDown,
		onMoveUp:            opts.OnMoveUp,
		onCardDown:          opts.OnCardDown,
		onCardUp:            opts.OnCardUp,
		onNextHeader:        opts.OnNextHeader,
		onPrevHeader:        opts.OnPrevHeader,
		onNextLink:          opts.OnNextLink,
		onPrevLink:          opts.OnPrevLink,
		onClick:             opts.OnClick,
		onDeleteCard:        opts.OnDeleteCard,
		onOpenInEditor:      opts.OnOpenInEditor,
		onAppendDone:        opts.OnAppendDone,
		onMoveCard:          opts.OnMoveCard,
		onMergeCard:         opts.OnMergeCard,
		onToggleFrontmatter: opts.OnToggleFrontmatter,
		onViewOptions:       opts.OnViewOptions,
		onToggleInlineTag:   opts.OnToggleInlineTag,
		onToggleInlineDate:  opts.OnToggleInlineDate,
		onOpenLink:          opts.OnOpenLink,
		onToggleTodo:        opts.OnToggleTodo,
		onFocusNote:         opts.OnFocusNote,
		onBack:              opts.OnBack,
		onNavBack:           opts.OnNavBack,
		onNavForward:        opts.OnNavForward,
		onShowInfo:          opts.OnShowInfo,
		onToggleTitle:       opts.OnToggleTitle,
		onToggleGlobalTags:  opts.OnToggleGlobalTags,
		onToggleMarkdown:    opts.OnToggleMarkdown,
		onOrderCards:        opts.OnOrderCards,
		onShowNavHistory:    opts.OnShowNavHistory,
	}
}

// Context returns the context this controller is attached to.
func (self *PreviewController) Context() types.Context {
	return self.getContext()
}

func (self *PreviewController) h() *helpers.Helpers {
	return self.c.Helpers().(*helpers.Helpers)
}

// Note action wrappers — call helpers directly.
func (self *PreviewController) addTag() error         { return self.h().NoteActions().AddGlobalTag() }
func (self *PreviewController) removeTag() error      { return self.h().NoteActions().RemoveTag() }
func (self *PreviewController) setParent() error      { return self.h().NoteActions().SetParentDialog() }
func (self *PreviewController) removeParent() error   { return self.h().NoteActions().RemoveParent() }
func (self *PreviewController) toggleBookmark() error { return self.h().NoteActions().ToggleBookmark() }

// GetKeybindingsFn returns the keybinding producer for preview.
// Bindings with Key == nil are palette-only (no keybinding registered).
func (self *PreviewController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			// Navigation (no Description → excluded from palette)
			{Key: 'j', Handler: self.onMoveDown},
			{Key: gocui.KeyArrowDown, Handler: self.onMoveDown},
			{Key: 'k', Handler: self.onMoveUp},
			{Key: gocui.KeyArrowUp, Handler: self.onMoveUp},
			{Key: 'J', Handler: self.onCardDown},
			{Key: 'K', Handler: self.onCardUp},
			{Key: '}', Handler: self.onNextHeader},
			{Key: '{', Handler: self.onPrevHeader},
			{Key: 'l', Handler: self.onNextLink},
			{Key: 'L', Handler: self.onPrevLink},

			// Card actions (Description → shown in palette)
			{
				ID:          "preview.delete",
				Key:         'd',
				Handler:     self.onDeleteCard,
				Description: "Delete Card",
				Category:    "Preview",
			},
			{
				ID:              "preview.open_editor",
				Key:             'E',
				Handler:         self.onOpenInEditor,
				Description:     "Open in Editor",
				Category:        "Preview",
				DisplayOnScreen: true,
			},
			{
				ID:          "preview.append_done",
				Key:         'D',
				Handler:     self.onAppendDone,
				Description: "Toggle #done",
				Category:    "Preview",
			},
			{
				ID:          "preview.move_card",
				Key:         'm',
				Handler:     self.onMoveCard,
				Description: "Move Card",
				Category:    "Preview",
			},
			{
				ID:          "preview.merge_card",
				Key:         'M',
				Handler:     self.onMergeCard,
				Description: "Merge Notes",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_frontmatter",
				Key:         'f',
				Handler:     self.onToggleFrontmatter,
				Description: "Toggle Frontmatter",
				Category:    "Preview",
			},
			{
				ID:          "preview.view_options",
				Key:         'v',
				Handler:     self.onViewOptions,
				Description: "View Options",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_inline_tag",
				Key:         gocui.KeyCtrlT,
				Handler:     self.onToggleInlineTag,
				Description: "Toggle Inline Tag",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_inline_date",
				Key:         gocui.KeyCtrlD,
				Handler:     self.onToggleInlineDate,
				Description: "Toggle Inline Date",
				Category:    "Preview",
			},
			{
				ID:          "preview.open_link",
				Key:         'o',
				Handler:     self.onOpenLink,
				Description: "Open Link",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_todo",
				Key:         'x',
				Handler:     self.onToggleTodo,
				Description: "Toggle Todo",
				Category:    "Preview",
			},
			{
				ID:              "preview.focus_note",
				Key:             gocui.KeyEnter,
				Handler:         self.onFocusNote,
				Description:     "Focus Note from Preview",
				Category:        "Preview",
				DisplayOnScreen: true,
			},
			// Back (no Description → excluded from palette)
			{
				ID:      "preview.back",
				Key:     gocui.KeyEsc,
				Handler: self.onBack,
			},
			{
				ID:          "preview.nav_back",
				Key:         '[',
				Handler:     self.onNavBack,
				Description: "Go Back",
				Category:    "Preview",
			},
			{
				ID:          "preview.nav_forward",
				Key:         ']',
				Handler:     self.onNavForward,
				Description: "Go Forward",
				Category:    "Preview",
			},

			// Note actions (shared with Notes panel) — call helpers directly
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
				Handler:     self.onShowInfo,
				Description: "Show Info",
				Category:    "Note Actions",
			},

			// Palette-only commands (Key == nil → no keybinding, palette entry only)
			{
				ID:          "preview.toggle_title",
				Handler:     self.onToggleTitle,
				Description: "Toggle Title",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_global_tags",
				Handler:     self.onToggleGlobalTags,
				Description: "Toggle Global Tags",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_markdown",
				Handler:     self.onToggleMarkdown,
				Description: "Toggle Markdown",
				Category:    "Preview",
			},
			{
				ID:          "preview.order_cards",
				Handler:     self.onOrderCards,
				Description: "Order Cards",
				Category:    "Preview",
			},
			{
				ID:          "preview.show_nav_history",
				Handler:     self.onShowNavHistory,
				Description: "View History",
				Category:    "Preview",
			},
		}
	}
}

// GetMouseKeybindingsFn returns mouse bindings for the preview panel.
func (self *PreviewController) GetMouseKeybindingsFn() types.MouseKeybindingsFn {
	return func(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
		return []*gocui.ViewMouseBinding{
			{
				ViewName: "preview",
				Key:      gocui.MouseLeft,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					return self.onClick()
				},
			},
		}
	}
}
