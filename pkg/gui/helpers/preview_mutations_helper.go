package helpers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// PreviewMutationsHelper handles card mutations: delete, move, merge, order.
type PreviewMutationsHelper struct {
	c *HelperCommon
}

// NewPreviewMutationsHelper creates a new PreviewMutationsHelper.
func NewPreviewMutationsHelper(c *HelperCommon) *PreviewMutationsHelper {
	return &PreviewMutationsHelper{c: c}
}

func (self *PreviewMutationsHelper) ctx() *context.CardListContext {
	return self.c.GuiCommon().Contexts().CardList
}

// DeleteCard deletes the currently selected card.
func (self *PreviewMutationsHelper) DeleteCard() error {
	cl := self.ctx()
	if len(cl.Cards) == 0 {
		return nil
	}

	card := cl.Cards[cl.SelectedCardIdx]
	displayName := card.Title
	if displayName == "" {
		displayName = card.Path
	}

	gui := self.c.GuiCommon()
	self.c.Helpers().Confirmation().ConfirmDelete("Note", displayName,
		func() error { return self.c.RuinCmd().Note.Delete(card.UUID) },
		func() {
			idx := cl.SelectedCardIdx
			cl.Cards = append(cl.Cards[:idx], cl.Cards[idx+1:]...)
			if cl.SelectedCardIdx >= len(cl.Cards) && cl.SelectedCardIdx > 0 {
				cl.SelectedCardIdx--
			}
			self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
			gui.RenderPreview()
		},
	)
	return nil
}

// MoveCardDialog shows the move direction menu.
func (self *PreviewMutationsHelper) MoveCardDialog() error {
	cl := self.ctx()
	if len(cl.Cards) <= 1 {
		return nil
	}
	self.c.GuiCommon().ShowMenuDialog("Move", []types.MenuItem{
		{Label: "Move card up", Key: "u", OnRun: func() error { return self.moveCard("up") }},
		{Label: "Move card down", Key: "d", OnRun: func() error { return self.moveCard("down") }},
	})
	return nil
}

func (self *PreviewMutationsHelper) moveCard(direction string) error {
	cl := self.ctx()
	idx := cl.SelectedCardIdx
	if direction == "up" {
		if idx <= 0 {
			return nil
		}
		cl.Cards[idx], cl.Cards[idx-1] = cl.Cards[idx-1], cl.Cards[idx]
		cl.SelectedCardIdx--
	} else {
		if idx >= len(cl.Cards)-1 {
			return nil
		}
		cl.Cards[idx], cl.Cards[idx+1] = cl.Cards[idx+1], cl.Cards[idx]
		cl.SelectedCardIdx++
	}

	if cl.TemporarilyMoved == nil {
		cl.TemporarilyMoved = make(map[int]bool)
	}
	cl.TemporarilyMoved[cl.SelectedCardIdx] = true

	gui := self.c.GuiCommon()
	gui.RenderPreview()
	ns := cl.NavState()
	newIdx := cl.SelectedCardIdx
	if newIdx < len(ns.CardLineRanges) {
		ns.CursorLine = ns.CardLineRanges[newIdx][0] + 1
	}
	gui.RenderPreview()
	return nil
}

// MergeCardDialog shows the merge direction menu.
func (self *PreviewMutationsHelper) MergeCardDialog() error {
	cl := self.ctx()
	if len(cl.Cards) <= 1 {
		return nil
	}
	self.c.GuiCommon().ShowMenuDialog("Merge", []types.MenuItem{
		{Label: "Merge card below into this one", Key: "d", OnRun: func() error { return self.executeMerge("down") }},
		{Label: "Merge card above into this one", Key: "u", OnRun: func() error { return self.executeMerge("up") }},
	})
	return nil
}

func (self *PreviewMutationsHelper) executeMerge(direction string) error {
	cl := self.ctx()
	idx := cl.SelectedCardIdx
	var targetIdx, sourceIdx int
	if direction == "down" {
		if idx >= len(cl.Cards)-1 {
			return nil
		}
		targetIdx = idx
		sourceIdx = idx + 1
	} else {
		if idx <= 0 {
			return nil
		}
		targetIdx = idx
		sourceIdx = idx - 1
	}

	target := cl.Cards[targetIdx]
	source := cl.Cards[sourceIdx]

	result, err := self.c.RuinCmd().Note.Merge(target.UUID, source.UUID, true, false)
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	cl.Cards[targetIdx].Content = ""
	if len(result.TagsMerged) > 0 {
		cl.Cards[targetIdx].Tags = result.TagsMerged
	}

	cl.Cards = append(cl.Cards[:sourceIdx], cl.Cards[sourceIdx+1:]...)
	if cl.SelectedCardIdx >= len(cl.Cards) {
		cl.SelectedCardIdx = len(cl.Cards) - 1
	}
	if cl.SelectedCardIdx < 0 {
		cl.SelectedCardIdx = 0
	}

	self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
	self.c.GuiCommon().RenderPreview()
	return nil
}

// OrderCards persists the current card order to frontmatter order fields.
func (self *PreviewMutationsHelper) OrderCards() error {
	cl := self.ctx()
	for i, card := range cl.Cards {
		if err := self.c.RuinCmd().Note.SetOrder(card.UUID, i+1); err != nil {
			self.c.GuiCommon().ShowError(err)
			return nil
		}
	}
	cl.TemporarilyMoved = nil
	self.c.Helpers().Preview().ReloadContent()
	return nil
}
