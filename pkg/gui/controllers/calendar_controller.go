package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// CalendarController handles keybindings for the calendar dialog popup.
type CalendarController struct {
	baseController
	getContext   func() *context.CalendarContext
	onGridLeft   func() error
	onGridRight  func() error
	onGridUp     func() error
	onGridDown   func() error
	onGridEnter  func() error
	onEsc        func() error // calendarGrid and calendarNotes
	onTab        func() error // all views
	onBacktab    func() error // calendarGrid and calendarInput
	onFocusInput func() error // calendarGrid and calendarNotes '/'
	onGridClick  func() error
	onInputEnter func() error
	onInputEsc   func() error
	onInputClick func() error
	onNoteDown   func() error
	onNoteUp     func() error
	onNoteEnter  func() error
}

var _ types.IController = &CalendarController{}

// CalendarControllerOpts holds the callbacks injected during wiring.
type CalendarControllerOpts struct {
	GetContext   func() *context.CalendarContext
	OnGridLeft   func() error
	OnGridRight  func() error
	OnGridUp     func() error
	OnGridDown   func() error
	OnGridEnter  func() error
	OnEsc        func() error
	OnTab        func() error
	OnBacktab    func() error
	OnFocusInput func() error
	OnGridClick  func() error
	OnInputEnter func() error
	OnInputEsc   func() error
	OnInputClick func() error
	OnNoteDown   func() error
	OnNoteUp     func() error
	OnNoteEnter  func() error
}

// NewCalendarController creates a CalendarController.
func NewCalendarController(opts CalendarControllerOpts) *CalendarController {
	return &CalendarController{
		getContext:   opts.GetContext,
		onGridLeft:   opts.OnGridLeft,
		onGridRight:  opts.OnGridRight,
		onGridUp:     opts.OnGridUp,
		onGridDown:   opts.OnGridDown,
		onGridEnter:  opts.OnGridEnter,
		onEsc:        opts.OnEsc,
		onTab:        opts.OnTab,
		onBacktab:    opts.OnBacktab,
		onFocusInput: opts.OnFocusInput,
		onGridClick:  opts.OnGridClick,
		onInputEnter: opts.OnInputEnter,
		onInputEsc:   opts.OnInputEsc,
		onInputClick: opts.OnInputClick,
		onNoteDown:   opts.OnNoteDown,
		onNoteUp:     opts.OnNoteUp,
		onNoteEnter:  opts.OnNoteEnter,
	}
}

// Context returns the context this controller is attached to.
func (self *CalendarController) Context() types.Context {
	return self.getContext()
}

// GetKeybindingsFn returns keybindings for the calendar dialog.
func (self *CalendarController) GetKeybindingsFn() types.KeybindingsFn {
	gv := "calendarGrid"
	iv := "calendarInput"
	nv := "calendarNotes"
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			// Grid navigation
			{ViewName: gv, Key: 'h', Handler: self.onGridLeft},
			{ViewName: gv, Key: 'l', Handler: self.onGridRight},
			{ViewName: gv, Key: 'k', Handler: self.onGridUp},
			{ViewName: gv, Key: 'j', Handler: self.onGridDown},
			{ViewName: gv, Key: gocui.KeyArrowLeft, Handler: self.onGridLeft},
			{ViewName: gv, Key: gocui.KeyArrowRight, Handler: self.onGridRight},
			{ViewName: gv, Key: gocui.KeyArrowUp, Handler: self.onGridUp},
			{ViewName: gv, Key: gocui.KeyArrowDown, Handler: self.onGridDown},
			{ViewName: gv, Key: gocui.KeyEnter, Handler: self.onGridEnter},
			{ViewName: gv, Key: gocui.KeyEsc, Handler: self.onEsc},
			{ViewName: gv, Key: gocui.KeyTab, Handler: self.onTab},
			{ViewName: gv, Key: gocui.KeyBacktab, Handler: self.onBacktab},
			{ViewName: gv, Key: '/', Handler: self.onFocusInput},
			// Input view
			{ViewName: iv, Key: gocui.KeyEnter, Handler: self.onInputEnter},
			{ViewName: iv, Key: gocui.KeyEsc, Handler: self.onInputEsc},
			{ViewName: iv, Key: gocui.KeyTab, Handler: self.onTab},
			{ViewName: iv, Key: gocui.KeyBacktab, Handler: self.onBacktab},
			// Note list navigation
			{ViewName: nv, Key: 'j', Handler: self.onNoteDown},
			{ViewName: nv, Key: 'k', Handler: self.onNoteUp},
			{ViewName: nv, Key: gocui.KeyArrowDown, Handler: self.onNoteDown},
			{ViewName: nv, Key: gocui.KeyArrowUp, Handler: self.onNoteUp},
			{ViewName: nv, Key: gocui.KeyEnter, Handler: self.onNoteEnter},
			{ViewName: nv, Key: gocui.KeyEsc, Handler: self.onEsc},
			{ViewName: nv, Key: gocui.KeyTab, Handler: self.onTab},
			{ViewName: nv, Key: gocui.KeyBacktab, Handler: self.onBacktab},
			{ViewName: nv, Key: '/', Handler: self.onFocusInput},
		}
	}
}

// GetMouseKeybindingsFn returns mouse bindings for the calendar views.
func (self *CalendarController) GetMouseKeybindingsFn() types.MouseKeybindingsFn {
	return func(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
		return []*gocui.ViewMouseBinding{
			{
				ViewName: "calendarGrid",
				Key:      gocui.MouseLeft,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					return self.onGridClick()
				},
			},
			{
				ViewName: "calendarInput",
				Key:      gocui.MouseLeft,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					return self.onInputClick()
				},
			},
		}
	}
}
