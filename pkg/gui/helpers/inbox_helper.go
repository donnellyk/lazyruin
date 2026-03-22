package helpers

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/inbox"
)

type InboxHelper struct {
	c        *HelperCommon
	store    *inbox.Store
	triggers func() []types.CompletionTrigger
}

func NewInboxHelper(c *HelperCommon) *InboxHelper {
	store := inbox.NewStore()
	if err := store.Load(); err != nil {
		// Non-fatal: inbox starts empty. Could log when logging is available.
	}
	return &InboxHelper{c: c, store: store}
}

// SetTriggers sets the completion trigger provider (called from gui package
// after initialization, since trigger functions live on *Gui).
func (self *InboxHelper) SetTriggers(fn func() []types.CompletionTrigger) {
	self.triggers = fn
}

func (self *InboxHelper) OpenInboxInput() error {
	self.c.Helpers().InputPopup().OpenInputPopup(&types.InputPopupConfig{
		Title:    "Jot to Inbox",
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

func (self *InboxHelper) OpenBrowser() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}
	if err := self.store.Load(); err != nil {
		gui.ShowError(err)
		return nil
	}
	ctx := gui.Contexts().InboxBrowser
	ctx.Items = self.store.Items()
	ctx.SelectedIdx = 0
	ctx.OnSelect = nil
	gui.PushContextByKey("inboxBrowser")
	return nil
}

// OpenBrowserForInsert opens the inbox browser on top of capture. When an item
// is selected, its text is inserted at the cursor in the capture view and the
// item is removed from the inbox.
func (self *InboxHelper) OpenBrowserForInsert(insertFn func(text string)) error {
	gui := self.c.GuiCommon()
	if err := self.store.Load(); err != nil {
		gui.ShowError(err)
		return nil
	}
	ctx := gui.Contexts().InboxBrowser
	ctx.Items = self.store.Items()
	ctx.SelectedIdx = 0
	ctx.OnSelect = func(item inbox.Item) error {
		self.DeleteItem(item.ID)
		insertFn(item.Text)
		return nil
	}
	gui.PushContextByKey("inboxBrowser")
	return nil
}

func (self *InboxHelper) DeleteItem(id string) {
	self.store.Delete(id)
	if err := self.store.Save(); err != nil {
		self.c.GuiCommon().ShowError(err)
	}
}

func (self *InboxHelper) RefreshBrowser() {
	gui := self.c.GuiCommon()
	if err := self.store.Load(); err != nil {
		gui.ShowError(err)
		return
	}
	ctx := gui.Contexts().InboxBrowser
	ctx.Items = self.store.Items()
	if ctx.SelectedIdx >= len(ctx.Items) {
		ctx.SelectedIdx = max(0, len(ctx.Items)-1)
	}
}

func (self *InboxHelper) HasItems() bool {
	return self.store.Len() > 0
}
