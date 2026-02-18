package gui

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/types"
)

// IGuiCommon defines the interface that controllers and helpers use
// to interact with the GUI without depending on the full Gui struct.
type IGuiCommon interface {
	PushContext(ctx types.Context, opts types.OnFocusOpts)
	PopContext()
	ReplaceContext(ctx types.Context)
	CurrentContext() types.Context
	PopupActive() bool

	GetView(name string) *gocui.View
	Render()
	Update(func() error)
	Suspend() error
	Resume() error
}
