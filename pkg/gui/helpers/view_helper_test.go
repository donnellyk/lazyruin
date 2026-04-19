package helpers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

// newScrollTestView builds a headless gocui view with the given inner height
// (visible rows) and a buffer of bufferLines total lines.
func newScrollTestView(t *testing.T, innerHeight, bufferLines int) (*gocui.Gui, *gocui.View) {
	t.Helper()
	g, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputNormal,
		Headless:   true,
		Width:      40,
		Height:     innerHeight + 2, // +2 for the view frame top/bottom borders
	})
	if err != nil {
		t.Fatalf("NewGui: %v", err)
	}
	// Lay the view out once so InnerSize is populated. SetView returns
	// ErrUnknownView on first creation — that's expected, not an error.
	g.SetManager(gocui.ManagerFunc(func(gg *gocui.Gui) error {
		_, err := gg.SetView("list", 0, 0, 39, innerHeight+1, 0)
		if err != nil && err.Error() == "unknown view" {
			return nil
		}
		return err
	}))
	if err := g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout: %v", err)
	}

	v, err := g.View("list")
	if err != nil {
		t.Fatalf("View: %v", err)
	}
	v.Clear()
	var b strings.Builder
	for i := 0; i < bufferLines; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	fmt.Fprint(v, b.String())
	if err := g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("redraw: %v", err)
	}
	return g, v
}

func TestScrollViewport_ClampsAtMax(t *testing.T) {
	// 10-line buffer in a 5-row viewport: max origin is 5 (so the last 5
	// lines fill the viewport). Scrolling past that should clamp, not drift.
	g, v := newScrollTestView(t, 5, 10)
	defer g.Close()

	// Scroll way past the end.
	for i := 0; i < 20; i++ {
		ScrollViewport(v, 3)
	}

	_, oy := v.Origin()
	if oy > 5 {
		t.Errorf("Origin y = %d after overscroll, want <= 5 (bufferLines - innerHeight)", oy)
	}
	if oy < 5 {
		t.Errorf("Origin y = %d, want 5 (last page fully visible)", oy)
	}
}

func TestScrollViewport_ClampsAtMin(t *testing.T) {
	// Existing behavior: scrolling past the top clamps to 0.
	g, v := newScrollTestView(t, 5, 10)
	defer g.Close()

	ScrollViewport(v, 7)
	ScrollViewport(v, -20)

	_, oy := v.Origin()
	if oy != 0 {
		t.Errorf("Origin y = %d, want 0 after overscroll up", oy)
	}
}

func TestScrollViewport_ShortBufferStaysAtZero(t *testing.T) {
	// Buffer shorter than the viewport: origin should stay at 0 — there's
	// nothing to scroll to.
	g, v := newScrollTestView(t, 10, 3)
	defer g.Close()

	ScrollViewport(v, 5)

	_, oy := v.Origin()
	if oy != 0 {
		t.Errorf("Origin y = %d for short buffer, want 0", oy)
	}
}
