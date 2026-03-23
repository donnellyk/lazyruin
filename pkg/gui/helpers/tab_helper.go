package helpers

// CycleTab advances to the next tab in tabs (wrapping around) and calls onChange.
func CycleTab[T comparable](tabs []T, currentIdx int, setCurrent func(T), onChange func()) {
	next := (currentIdx + 1) % len(tabs)
	setCurrent(tabs[next])
	onChange()
}

// SwitchTab switches to a specific tab index and calls onChange.
// If idx is out of range, it does nothing.
func SwitchTab[T comparable](tabs []T, idx int, setCurrent func(T), onChange func()) {
	if idx < 0 || idx >= len(tabs) {
		return
	}
	setCurrent(tabs[idx])
	onChange()
}
