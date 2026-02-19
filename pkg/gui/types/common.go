package types

// OnFocusOpts is passed when a context gains focus.
type OnFocusOpts struct {
	ClickedViewLineIdx      int
	ScrollSelectionIntoView bool
}

// OnFocusLostOpts is passed when a context loses focus.
type OnFocusLostOpts struct {
	NewContextKey ContextKey
}
