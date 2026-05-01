package gui

import (
	"strings"
	"testing"
)

// TestCapturePopup_WrapStaysWithinInnerWidth verifies that the New Note
// dialog's autowrap does not produce lines that overflow the visible area.
//
// Root cause being guarded against: gocui's TextArea autowrap algorithm
// (contentToCells) only triggers a wrap when a non-whitespace character
// pushes the line past AutoWrapWidth. A trailing space that itself exceeds
// AutoWrapWidth is silently included on the previous line. So when
// AutoWrapWidth == InnerWidth and the user types a word whose length is
// exactly InnerWidth followed by a space, the trailing space lands at
// column InnerWidth (one past the last visible column) — the off-by-one.
//
// The fix is to set AutoWrapWidth = InnerWidth - 1, so the trailing-space
// overflow still fits inside the visible area.
func TestCapturePopup_WrapStaysWithinInnerWidth(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if err := tg.gui.helpers.Capture().OpenCapture(); err != nil {
		t.Fatalf("OpenCapture: %v", err)
	}
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout: %v", err)
	}

	v := tg.gui.views.Capture
	if v == nil {
		t.Fatal("capture view is nil")
	}

	inner := v.InnerWidth()
	if inner < 4 {
		t.Fatalf("unexpected InnerWidth %d", inner)
	}

	// Build content that exposes the overflow:
	//   word_a + " " + word_b + " " + word_c
	// where len(word_a) + 1 + len(word_b) == inner — i.e. the second word
	// ends exactly at the last visible column. With AutoWrapWidth = InnerWidth,
	// gocui's wrap routine doesn't fire on the trailing space (spaces are
	// handled in a branch that skips the overflow check); it waits for the
	// next non-whitespace character. By then the line has accumulated
	// `inner + 1` cells and the trailing space sits one past the last
	// visible column — the off-by-one.
	wordALen := inner / 3
	wordBLen := inner - 1 - wordALen // so wordA + " " + wordB == inner chars
	wordA := strings.Repeat("a", wordALen)
	wordB := strings.Repeat("b", wordBLen)
	v.TextArea.TypeString(wordA + " " + wordB + " cc")

	wrapped := v.TextArea.GetContent()
	for i, line := range strings.Split(wrapped, "\n") {
		if w := len([]rune(line)); w > inner {
			t.Errorf("line %d %q has width %d, exceeds InnerWidth %d (off-by-one wrap)",
				i, line, w, inner)
		}
	}
}
