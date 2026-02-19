package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// CalendarController handles keybindings for the calendar dialog popup.
type CalendarController struct {
	baseController
	c          *ControllerCommon
	getContext func() *context.CalendarContext
}

var _ types.IController = &CalendarController{}

// CalendarControllerOpts holds the dependencies for CalendarController.
type CalendarControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.CalendarContext
}

// NewCalendarController creates a CalendarController.
func NewCalendarController(opts CalendarControllerOpts) *CalendarController {
	return &CalendarController{
		c:          opts.Common,
		getContext: opts.GetContext,
	}
}

// Context returns the context this controller is attached to.
func (self *CalendarController) Context() types.Context {
	return self.getContext()
}

func (self *CalendarController) gridLeft() error  { self.c.Helpers().Calendar().MoveDay(-1); return nil }
func (self *CalendarController) gridRight() error { self.c.Helpers().Calendar().MoveDay(1); return nil }
func (self *CalendarController) gridUp() error    { self.c.Helpers().Calendar().MoveDay(-7); return nil }
func (self *CalendarController) gridDown() error  { self.c.Helpers().Calendar().MoveDay(7); return nil }

func (self *CalendarController) close() error {
	self.c.Helpers().Calendar().Close()
	return nil
}

// GetKeybindings returns keybindings for the calendar dialog.
func (self *CalendarController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	cal := func() *CalendarController { return self }
	gv := "calendarGrid"
	iv := "calendarInput"
	nv := "calendarNotes"
	return []*types.Binding{
		// Grid navigation
		{ViewName: gv, Key: 'h', Handler: cal().gridLeft},
		{ViewName: gv, Key: 'l', Handler: cal().gridRight},
		{ViewName: gv, Key: 'k', Handler: cal().gridUp},
		{ViewName: gv, Key: 'j', Handler: cal().gridDown},
		{ViewName: gv, Key: gocui.KeyArrowLeft, Handler: cal().gridLeft},
		{ViewName: gv, Key: gocui.KeyArrowRight, Handler: cal().gridRight},
		{ViewName: gv, Key: gocui.KeyArrowUp, Handler: cal().gridUp},
		{ViewName: gv, Key: gocui.KeyArrowDown, Handler: cal().gridDown},
		{ViewName: gv, Key: gocui.KeyEnter, Handler: self.c.Helpers().Calendar().GridEnter},
		{ViewName: gv, Key: gocui.KeyEsc, Handler: cal().close},
		{ViewName: gv, Key: gocui.KeyTab, Handler: self.c.Helpers().Calendar().Tab},
		{ViewName: gv, Key: gocui.KeyBacktab, Handler: self.c.Helpers().Calendar().Backtab},
		{ViewName: gv, Key: '/', Handler: self.c.Helpers().Calendar().FocusInput},
		// Input view
		{ViewName: iv, Key: gocui.KeyEnter, Handler: self.c.Helpers().Calendar().InputEnter},
		{ViewName: iv, Key: gocui.KeyEsc, Handler: self.c.Helpers().Calendar().InputEsc},
		{ViewName: iv, Key: gocui.KeyTab, Handler: self.c.Helpers().Calendar().Tab},
		{ViewName: iv, Key: gocui.KeyBacktab, Handler: self.c.Helpers().Calendar().Backtab},
		// Note list navigation
		{ViewName: nv, Key: 'j', Handler: self.c.Helpers().Calendar().NoteDown},
		{ViewName: nv, Key: 'k', Handler: self.c.Helpers().Calendar().NoteUp},
		{ViewName: nv, Key: gocui.KeyArrowDown, Handler: self.c.Helpers().Calendar().NoteDown},
		{ViewName: nv, Key: gocui.KeyArrowUp, Handler: self.c.Helpers().Calendar().NoteUp},
		{ViewName: nv, Key: gocui.KeyEnter, Handler: self.c.Helpers().Calendar().NoteEnter},
		{ViewName: nv, Key: gocui.KeyEsc, Handler: cal().close},
		{ViewName: nv, Key: gocui.KeyTab, Handler: self.c.Helpers().Calendar().Tab},
		{ViewName: nv, Key: gocui.KeyBacktab, Handler: self.c.Helpers().Calendar().Backtab},
		{ViewName: nv, Key: '/', Handler: self.c.Helpers().Calendar().FocusInput},
	}
}

// GetMouseKeybindings returns mouse bindings for the calendar views.
func (self *CalendarController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: "calendarGrid",
			Key:      gocui.MouseLeft,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return self.c.Helpers().Calendar().GridClick()
			},
		},
		{
			ViewName: "calendarInput",
			Key:      gocui.MouseLeft,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return self.c.Helpers().Calendar().InputClick()
			},
		},
	}
}
