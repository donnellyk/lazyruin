package context

import (
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/scratchpad"
)

type ScratchpadBrowserContext struct {
	BaseContext
	Items       []scratchpad.Item
	SelectedIdx int
	OnSelect    func(item scratchpad.Item) error // custom action on Enter; nil = promote to capture
}

func NewScratchpadBrowserContext() *ScratchpadBrowserContext {
	return &ScratchpadBrowserContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.TEMPORARY_POPUP,
			Key:       "scratchpadBrowser",
			ViewName:  "scratchpadBrowser",
			Focusable: true,
			Title:     "Scratchpad",
		}),
	}
}

var _ types.Context = &ScratchpadBrowserContext{}
