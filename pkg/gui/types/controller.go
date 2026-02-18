package types

// IController defines the contract for a controller that supplies
// behavior to a context. Controllers provide producer functions
// that the context aggregates.
type IController interface {
	// Context returns the context this controller is attached to.
	Context() Context

	// Binding producers (return nil if not applicable)
	GetKeybindingsFn() KeybindingsFn
	GetMouseKeybindingsFn() MouseKeybindingsFn

	// Lifecycle hooks (return nil if not applicable)
	GetOnRenderToMain() func()
	GetOnFocus() func(OnFocusOpts)
	GetOnFocusLost() func(OnFocusLostOpts)
}
