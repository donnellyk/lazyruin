package context

// TabIndexOf returns the index of current in tabs, or 0 if not found.
func TabIndexOf[T comparable](tabs []T, current T) int {
	for i, tab := range tabs {
		if tab == current {
			return i
		}
	}
	return 0
}
