package types

import "github.com/jesseduffield/gocui"

// ContextKind classifies contexts for focus management.
type ContextKind int

const (
	SIDE_CONTEXT     ContextKind = iota // Notes, Tags, Queries panels
	MAIN_CONTEXT                        // Preview panel
	PERSISTENT_POPUP                    // Search, Capture, Calendar — can return to
	TEMPORARY_POPUP                     // Pick, Palette, Menus — cannot return to
	GLOBAL_CONTEXT                      // Global keybindings only
)

// ContextKey uniquely identifies a context.
type ContextKey string

// Context is the full interface for a rich context object.
type Context interface {
	IBaseContext

	HandleFocus(opts OnFocusOpts)
	HandleFocusLost(opts OnFocusLostOpts)
	HandleRender()
}

// IBaseContext defines the core identity and binding aggregation for a context.
type IBaseContext interface {
	GetKind() ContextKind
	GetKey() ContextKey
	IsFocusable() bool
	Title() string

	// View identity — contexts store view *names*, not *gocui.View pointers,
	// because views may be nil before layout or during resize. Views are
	// looked up at render time via the gui.
	//
	// Multi-view contexts (e.g., Calendar with grid/input/notes views)
	// return multiple names. Keybindings are registered for all of them.
	// The primary view receives focus when the context is activated.
	GetViewNames() []string
	GetPrimaryViewName() string

	// Aggregated keybindings (collected from attached controllers)
	GetKeybindings(opts KeybindingsOpts) []*Binding
	GetMouseKeybindings(opts KeybindingsOpts) []*gocui.ViewMouseBinding
	GetOnClick() func() error

	// Tab click bindings (for tabbed panels like Notes, Queries, Tags)
	GetTabClickBindingFn() func(int) error

	// Controller attachment points
	AddKeybindingsFn(KeybindingsFn)
	AddMouseKeybindingsFn(MouseKeybindingsFn)
	AddOnFocusFn(func(OnFocusOpts))
	AddOnFocusLostFn(func(OnFocusLostOpts))
	AddOnRenderToMainFn(func())
}

// IListContext extends Context with list-specific behavior.
type IListContext interface {
	Context

	GetList() IList
	GetSelectedItemId() string
}
