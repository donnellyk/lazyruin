package helpers

import "github.com/jesseduffield/gocui"

// ListClickIndex returns the item index for a mouse click in a list view
// with the given item height (lines per item).
func ListClickIndex(v *gocui.View, itemHeight int) int {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	return (cy + oy) / itemHeight
}

// ScrollViewport scrolls a list view's origin by delta lines, clamped to
// [0, bufferLines - innerHeight]. Without the upper clamp, mouse-wheel
// scrolling past the last line silently drifts the origin, forcing the user
// to press up/k many times before the viewport begins scrolling back.
func ScrollViewport(v *gocui.View, delta int) {
	_, oy := v.Origin()
	_, innerH := v.InnerSize()
	maxOy := max(len(v.ViewBufferLines())-innerH, 0)
	newOy := min(max(oy+delta, 0), maxOy)
	v.SetOrigin(0, newOy)
}
