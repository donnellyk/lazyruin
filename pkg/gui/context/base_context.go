package context

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/types"
)

// BaseContext provides the common infrastructure for all contexts.
// It aggregates keybinding producer functions and lifecycle hooks
// from attached controllers.
type BaseContext struct {
	kind            types.ContextKind
	key             types.ContextKey
	viewNames       []string
	primaryViewName string
	focusable       bool
	title           string

	keybindingsFns      []types.KeybindingsFn
	mouseKeybindingsFns []types.MouseKeybindingsFn
	onFocusFns          []func(types.OnFocusOpts)
	onFocusLostFns      []func(types.OnFocusLostOpts)
	onRenderToMainFns   []func()
	onClickFn           func() error
	tabClickBindingFn   func(int) error
}

// NewBaseContext creates a BaseContext with a single view name.
func NewBaseContext(opts NewBaseContextOpts) BaseContext {
	viewNames := opts.ViewNames
	if len(viewNames) == 0 && opts.ViewName != "" {
		viewNames = []string{opts.ViewName}
	}
	primaryViewName := opts.PrimaryViewName
	if primaryViewName == "" && len(viewNames) > 0 {
		primaryViewName = viewNames[0]
	}
	return BaseContext{
		kind:            opts.Kind,
		key:             opts.Key,
		viewNames:       viewNames,
		primaryViewName: primaryViewName,
		focusable:       opts.Focusable,
		title:           opts.Title,
	}
}

// NewBaseContextOpts holds initialization parameters for BaseContext.
type NewBaseContextOpts struct {
	Kind            types.ContextKind
	Key             types.ContextKey
	ViewName        string   // convenience for single-view contexts
	ViewNames       []string // for multi-view contexts (overrides ViewName)
	PrimaryViewName string   // receives focus; defaults to first ViewName
	Focusable       bool
	Title           string
}

func (self *BaseContext) GetKind() types.ContextKind { return self.kind }
func (self *BaseContext) GetKey() types.ContextKey   { return self.key }
func (self *BaseContext) IsFocusable() bool          { return self.focusable }
func (self *BaseContext) Title() string              { return self.title }
func (self *BaseContext) SetTitle(t string)          { self.title = t }
func (self *BaseContext) GetViewNames() []string     { return self.viewNames }
func (self *BaseContext) GetPrimaryViewName() string { return self.primaryViewName }

func (self *BaseContext) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	var bindings []*types.Binding
	for _, fn := range self.keybindingsFns {
		bindings = append(bindings, fn(opts)...)
	}
	return bindings
}

func (self *BaseContext) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	var bindings []*gocui.ViewMouseBinding
	for _, fn := range self.mouseKeybindingsFns {
		bindings = append(bindings, fn(opts)...)
	}
	return bindings
}

func (self *BaseContext) GetOnClick() func() error {
	return self.onClickFn
}

func (self *BaseContext) SetOnClick(fn func() error) {
	self.onClickFn = fn
}

func (self *BaseContext) GetTabClickBindingFn() func(int) error {
	return self.tabClickBindingFn
}

func (self *BaseContext) SetTabClickBindingFn(fn func(int) error) {
	self.tabClickBindingFn = fn
}

func (self *BaseContext) AddKeybindingsFn(fn types.KeybindingsFn) {
	self.keybindingsFns = append(self.keybindingsFns, fn)
}

func (self *BaseContext) AddMouseKeybindingsFn(fn types.MouseKeybindingsFn) {
	self.mouseKeybindingsFns = append(self.mouseKeybindingsFns, fn)
}

func (self *BaseContext) AddOnFocusFn(fn func(types.OnFocusOpts)) {
	self.onFocusFns = append(self.onFocusFns, fn)
}

func (self *BaseContext) AddOnFocusLostFn(fn func(types.OnFocusLostOpts)) {
	self.onFocusLostFns = append(self.onFocusLostFns, fn)
}

func (self *BaseContext) AddOnRenderToMainFn(fn func()) {
	self.onRenderToMainFns = append(self.onRenderToMainFns, fn)
}

// HandleFocus calls all registered focus hooks.
func (self *BaseContext) HandleFocus(opts types.OnFocusOpts) {
	for _, fn := range self.onFocusFns {
		fn(opts)
	}
}

// HandleFocusLost calls all registered focus-lost hooks.
func (self *BaseContext) HandleFocusLost(opts types.OnFocusLostOpts) {
	for _, fn := range self.onFocusLostFns {
		fn(opts)
	}
}

// HandleRender calls all registered render-to-main hooks.
func (self *BaseContext) HandleRender() {
	for _, fn := range self.onRenderToMainFns {
		fn()
	}
}
