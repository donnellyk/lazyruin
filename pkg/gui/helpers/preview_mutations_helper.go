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

func (self *PreviewMutationsHelper) ctx() *context.PreviewContext {
	return self.c.GuiCommon().Contexts().Preview
}

// DeleteCard deletes the currently selected card.
func (self *PreviewMutationsHelper) DeleteCard() error {
	pc := self.ctx()
	if len(pc.Cards) == 0 {
		return nil
	}

	card := pc.Cards[pc.SelectedCardIndex]
	displayName := card.Title
	if displayName == "" {
		displayName = card.Path
	}

	gui := self.c.GuiCommon()
	self.c.Helpers().Confirmation().ConfirmDelete("Note", displayName,
		func() error { return self.c.RuinCmd().Note.Delete(card.UUID) },
		func() {
			idx := pc.SelectedCardIndex
			pc.Cards = append(pc.Cards[:idx], pc.Cards[idx+1:]...)
			if pc.SelectedCardIndex >= len(pc.Cards) && pc.SelectedCardIndex > 0 {
				pc.SelectedCardIndex--
			}
			self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
			gui.RenderPreview()
		},
	)
	return nil
}

// MoveCardDialog shows the move direction menu.
func (self *PreviewMutationsHelper) MoveCardDialog() error {
	pc := self.ctx()
	if len(pc.Cards) <= 1 {
		return nil
	}
	self.c.GuiCommon().ShowMenuDialog("Move", []types.MenuItem{
		{Label: "Move card up", Key: "u", OnRun: func() error { return self.moveCard("up") }},
		{Label: "Move card down", Key: "d", OnRun: func() error { return self.moveCard("down") }},
	})
	return nil
}

func (self *PreviewMutationsHelper) moveCard(direction string) error {
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	if direction == "up" {
		if idx <= 0 {
			return nil
		}
		pc.Cards[idx], pc.Cards[idx-1] = pc.Cards[idx-1], pc.Cards[idx]
		pc.SelectedCardIndex--
	} else {
		if idx >= len(pc.Cards)-1 {
			return nil
		}
		pc.Cards[idx], pc.Cards[idx+1] = pc.Cards[idx+1], pc.Cards[idx]
		pc.SelectedCardIndex++
	}

	if pc.TemporarilyMoved == nil {
		pc.TemporarilyMoved = make(map[int]bool)
	}
	pc.TemporarilyMoved[pc.SelectedCardIndex] = true

	gui := self.c.GuiCommon()
	gui.RenderPreview()
	newIdx := pc.SelectedCardIndex
	if newIdx < len(pc.CardLineRanges) {
		pc.CursorLine = pc.CardLineRanges[newIdx][0] + 1
	}
	gui.RenderPreview()
	return nil
}

// MergeCardDialog shows the merge direction menu.
func (self *PreviewMutationsHelper) MergeCardDialog() error {
	pc := self.ctx()
	if len(pc.Cards) <= 1 {
		return nil
	}
	self.c.GuiCommon().ShowMenuDialog("Merge", []types.MenuItem{
		{Label: "Merge card below into this one", Key: "d", OnRun: func() error { return self.executeMerge("down") }},
		{Label: "Merge card above into this one", Key: "u", OnRun: func() error { return self.executeMerge("up") }},
	})
	return nil
}

func (self *PreviewMutationsHelper) executeMerge(direction string) error {
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	var targetIdx, sourceIdx int
	if direction == "down" {
		if idx >= len(pc.Cards)-1 {
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

	target := pc.Cards[targetIdx]
	source := pc.Cards[sourceIdx]

	result, err := self.c.RuinCmd().Note.Merge(target.UUID, source.UUID, true, false)
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	pc.Cards[targetIdx].Content = ""
	if len(result.TagsMerged) > 0 {
		pc.Cards[targetIdx].Tags = result.TagsMerged
	}

	pc.Cards = append(pc.Cards[:sourceIdx], pc.Cards[sourceIdx+1:]...)
	if pc.SelectedCardIndex >= len(pc.Cards) {
		pc.SelectedCardIndex = len(pc.Cards) - 1
	}
	if pc.SelectedCardIndex < 0 {
		pc.SelectedCardIndex = 0
	}

	self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
	self.c.GuiCommon().RenderPreview()
	return nil
}

// OrderCards persists the current card order to frontmatter order fields.
func (self *PreviewMutationsHelper) OrderCards() error {
	pc := self.ctx()
	for i, card := range pc.Cards {
		if err := self.c.RuinCmd().Note.SetOrder(card.UUID, i+1); err != nil {
			self.c.GuiCommon().ShowError(err)
			return nil
		}
	}
	pc.TemporarilyMoved = nil
	self.c.Helpers().Preview().ReloadContent()
	return nil
}
