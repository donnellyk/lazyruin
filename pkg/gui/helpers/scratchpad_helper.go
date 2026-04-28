package helpers

import (
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/scratchpad"
)

type ScratchpadHelper struct {
	c        *HelperCommon
	store    *scratchpad.Store
	triggers func() []types.CompletionTrigger
}

func NewScratchpadHelper(c *HelperCommon) *ScratchpadHelper {
	store := scratchpad.NewStoreForVault(c.RuinCmd().VaultPath())
	if err := store.Load(); err != nil {
		// Non-fatal: scratchpad starts empty. Could log when logging is available.
	}
	return &ScratchpadHelper{c: c, store: store}
}

// SetTriggers sets the completion trigger provider (called from gui package
// after initialization, since trigger functions live on *Gui).
func (self *ScratchpadHelper) SetTriggers(fn func() []types.CompletionTrigger) {
	self.triggers = fn
}

func (self *ScratchpadHelper) OpenScratchpadInput() error {
	self.c.Helpers().InputPopup().OpenInputPopup(&types.InputPopupConfig{
		Title:    "Jot to Scratchpad",
		Footer:   " # for tags | [[ wiki-links | @ dates | Tab: complete ",
		Triggers: self.triggers,
		OnAccept: func(raw string, _ *types.CompletionItem) error {
			if raw == "" {
				return nil
			}
			self.store.Add(raw)
			return self.store.Save()
		},
	})
	return nil
}

func (self *ScratchpadHelper) OpenBrowser() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}
	if err := self.store.Load(); err != nil {
		gui.ShowError(err)
		return nil
	}
	ctx := gui.Contexts().ScratchpadBrowser
	ctx.Items = self.store.Items()
	ctx.SelectedIdx = 0
	ctx.OnSelect = nil
	gui.PushContextByKey("scratchpadBrowser")
	return nil
}

// OpenBrowserForInsert opens the scratchpad browser on top of capture. When an
// item is selected, its text is inserted at the cursor in the capture view and
// the item is removed from the scratchpad.
func (self *ScratchpadHelper) OpenBrowserForInsert(insertFn func(text string)) error {
	gui := self.c.GuiCommon()
	if err := self.store.Load(); err != nil {
		gui.ShowError(err)
		return nil
	}
	ctx := gui.Contexts().ScratchpadBrowser
	ctx.Items = self.store.Items()
	ctx.SelectedIdx = 0
	ctx.OnSelect = func(item scratchpad.Item) error {
		self.DeleteItem(item.ID)
		insertFn(item.Text)
		return nil
	}
	gui.PushContextByKey("scratchpadBrowser")
	return nil
}

func (self *ScratchpadHelper) DeleteItem(id string) {
	self.store.Delete(id)
	if err := self.store.Save(); err != nil {
		self.c.GuiCommon().ShowError(err)
	}
}

func (self *ScratchpadHelper) RefreshBrowser() {
	gui := self.c.GuiCommon()
	if err := self.store.Load(); err != nil {
		gui.ShowError(err)
		return
	}
	ctx := gui.Contexts().ScratchpadBrowser
	ctx.Items = self.store.Items()
	if ctx.SelectedIdx >= len(ctx.Items) {
		ctx.SelectedIdx = max(0, len(ctx.Items)-1)
	}
}

func (self *ScratchpadHelper) HasItems() bool {
	return self.store.Len() > 0
}
