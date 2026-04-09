package context

import (
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/inbox"
)

type InboxBrowserContext struct {
	BaseContext
	Items       []inbox.Item
	SelectedIdx int
	OnSelect    func(item inbox.Item) error // custom action on Enter; nil = promote to capture
}

func NewInboxBrowserContext() *InboxBrowserContext {
	return &InboxBrowserContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.TEMPORARY_POPUP,
			Key:       "inboxBrowser",
			ViewName:  "inboxBrowser",
			Focusable: true,
			Title:     "Inbox",
		}),
	}
}

var _ types.Context = &InboxBrowserContext{}
