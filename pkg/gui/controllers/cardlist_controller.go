package controllers

import (
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/helpers"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/jesseduffield/gocui"
)

// CardListController handles keybindings for the card-list preview mode.
// Includes full navigation + mutations + line-ops + note-actions.
type CardListController struct {
	baseController
	PreviewNavTrait
	NoteActionHandlersTrait
	FilterablePreviewTrait
	c          *ControllerCommon
	getContext func() *context.CardListContext
}

var _ types.IController = &CardListController{}

func NewCardListController(c *ControllerCommon, getContext func() *context.CardListContext) *CardListController {
	return &CardListController{
		PreviewNavTrait:         PreviewNavTrait{c: c},
		NoteActionHandlersTrait: NoteActionHandlersTrait{c: c},
		FilterablePreviewTrait:  FilterablePreviewTrait{c: c},
		c:                       c,
		getContext:              getContext,
	}
}

func (self *CardListController) Context() types.Context { return self.getContext() }

func (self *CardListController) mutations() *helpers.PreviewMutationsHelper {
	return self.c.Helpers().PreviewMutations()
}

func (self *CardListController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return self.BuildPreviewBindings("cardList",
		&types.Binding{
			ID: "cardList.delete", Key: 'd',
			Handler: self.mutations().DeleteCard, Description: "Delete Card", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "Del",
		},
		&types.Binding{
			ID: "cardList.open_editor", Key: 'E',
			Handler: self.nav().OpenInEditor, Description: "Open in Editor", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "Editor",
		},
		&types.Binding{
			ID: "cardList.edit_inline", Key: 'e',
			Handler: self.editInline, Description: "Edit in Popup", Category: "Preview",
		},
		&types.Binding{
			ID: "cardList.move_card", Key: 'm',
			Handler: self.mutations().MoveCardDialog, Description: "Move Card", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "Move",
		},
		&types.Binding{
			ID: "cardList.merge_card", Key: 'M',
			Handler: self.mutations().MergeCardDialog, Description: "Merge Notes", Category: "Preview",
		},
		&types.Binding{
			ID: "cardList.add_tag", Key: 't',
			Handler: self.addTag, Description: "Add Tag", Category: "Note Actions",
			DisplayOnScreen: true, StatusBarLabel: "Tag",
		},
		&types.Binding{
			ID: "cardList.remove_tag", Key: 'T',
			Handler: self.removeTag, Description: "Remove Tag", Category: "Note Actions",
		},
		&types.Binding{
			ID: "cardList.set_parent", Key: '>',
			Handler: self.setParent, Description: "Set Parent", Category: "Note Actions",
		},
		&types.Binding{
			ID: "cardList.remove_parent", Key: 'P',
			Handler: self.removeParent, Description: "Remove Parent", Category: "Note Actions",
		},
		&types.Binding{
			ID: "cardList.toggle_bookmark", Key: 'b',
			Handler: self.toggleBookmark, Description: "Toggle Bookmark", Category: "Note Actions",
		},
		&types.Binding{
			ID:      "cardList.order_cards",
			Handler: self.mutations().OrderCards, Description: "Order Cards", Category: "Preview",
		},
		&types.Binding{
			ID: "cardList.open_url", Key: 'o',
			Handler: self.openURL, Description: "Open URL", Category: "Preview",
		},
		&types.Binding{
			ID: "cardList.re_resolve_link", Key: 'R',
			Handler:           self.reResolveLink,
			GetDisabledReason: requireLinkNote(func() *models.Note { return self.c.Helpers().Preview().CurrentPreviewCard() }),
			Description:       "Re-resolve Link", Category: "Preview",
		},
		&types.Binding{
			ID: "cardList.filter", Key: 'F',
			Handler: self.openFilter, Description: "Filter Cards", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "Filter",
		},
		&types.Binding{
			ID: "cardList.clear_filter", Key: 'X',
			Handler:           self.clearFilter,
			GetDisabledReason: self.filterNotActive,
			Description:       "Clear Filter", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "Clear",
		},
	)
}

func (self *CardListController) openURL() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	return self.c.Helpers().Link().OpenLinkURL(card)
}

func (self *CardListController) editInline() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	return self.c.Helpers().Capture().OpenCaptureForEdit(card)
}

func (self *CardListController) reResolveLink() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	return self.c.Helpers().Link().ReResolveLink(card)
}

func (self *CardListController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return self.NavMouseBindings()
}
