package helpers

import "github.com/jesseduffield/gocui"

// ListClickIndex returns the item index for a mouse click in a list view
// with the given item height (lines per item).
func ListClickIndex(v *gocui.View, itemHeight int) int {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	return (cy + oy) / itemHeight
}

// ScrollViewport scrolls a list view's origin by delta lines without
// constraining the selection to stay visible.
func ScrollViewport(v *gocui.View, delta int) {
	_, oy := v.Origin()
	newOy := oy + delta
	if newOy < 0 {
		newOy = 0
	}
	v.SetOrigin(0, newOy)
}
