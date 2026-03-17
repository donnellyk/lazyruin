package gui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

// TestCaptureWrapOriginResets verifies that after a line wraps, the X origin
// resets to 0 so the first character of the wrapped continuation is visible.
//
// Scenario:
//  1. Type text that fills the view width (cursor reaches end of visible area)
//  2. Type a space (cursor moves past InnerWidth, origin would shift with old code)
//  3. Type another character → wrapping triggers, cursor moves to line 2
//  4. The X origin must be 0 so the first column is not clipped
func TestCaptureWrapOriginResets(t *testing.T) {
	innerWidth := 20

	ta := &gocui.TextArea{}
	ta.AutoWrap = true
	ta.AutoWrapWidth = innerWidth // fixed: was innerWidth-1

	// Type chars to fill the line up to AutoWrapWidth
	for i := 0; i < innerWidth-1; i++ {
		ta.TypeCharacter(string(rune('a' + (i % 26))))
	}

	// Type a space then 'x' — triggers wrapping, cursor lands at (1,1)
	ta.TypeCharacter(" ")
	ta.TypeCharacter("x")

	cursorX, cursorY := ta.GetCursorXY()
	if cursorY != 1 {
		t.Fatalf("expected wrapping to line 1, got cursorY=%d", cursorY)
	}

	// Simulate the state after the space was typed: origin was 1 at that point.
	// With the fixed captureXCursorAndOrigin, origin always resets to 0 when
	// cursorX is within view width.
	prevOriginX := 1 // this was set by the (now-gone) stale carry-over
	_ = prevOriginX  // not used by captureXCursorAndOrigin — that's the fix

	viewCursorX, newOriginX := captureXCursorAndOrigin(innerWidth, cursorX)
	if newOriginX != 0 {
		t.Errorf("origin X should be 0 after wrap, got %d (cursorX=%d, viewCursorX=%d)",
			newOriginX, cursorX, viewCursorX)
	}
	if viewCursorX != cursorX {
		t.Errorf("viewCursorX should equal cursorX (%d) when origin is 0, got %d",
			cursorX, viewCursorX)
	}
}

// TestCaptureXCursorAndOrigin tests captureXCursorAndOrigin directly.
func TestCaptureXCursorAndOrigin(t *testing.T) {
	width := 20

	tests := []struct {
		name        string
		cursorX     int
		wantViewCur int
		wantOrigin  int
	}{
		{"cursor at start", 0, 0, 0},
		{"cursor mid-line", 10, 10, 0},
		{"cursor at last visible col", 19, 19, 0},
		{"cursor one past visible", 20, 19, 1},
		{"cursor well past visible", 30, 19, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewCur, origin := captureXCursorAndOrigin(width, tt.cursorX)
			if viewCur != tt.wantViewCur || origin != tt.wantOrigin {
				t.Errorf("captureXCursorAndOrigin(%d, %d) = (%d, %d), want (%d, %d)",
					width, tt.cursorX, viewCur, origin, tt.wantViewCur, tt.wantOrigin)
			}
		})
	}
}

// TestCaptureAutoWrapWidth verifies that AutoWrapWidth = InnerWidth uses the
// full available width, not InnerWidth-1 which wastes one column.
func TestCaptureAutoWrapWidth(t *testing.T) {
	innerWidth := 20

	// With AutoWrapWidth = innerWidth (fixed), the first word group "abcdefghijklmnopq rs"
	// should stay on one line since "rs " + "x" wraps at the second space (position 20).
	ta := &gocui.TextArea{}
	ta.AutoWrap = true
	ta.AutoWrapWidth = innerWidth
	ta.TypeString("abcdefghijklmnopq rs x")

	content := ta.GetContent()
	lines := strings.Split(content, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), content)
	}

	// First line should include "abcdefghijklmnopq rs " — wrapping at the full width.
	// Trim trailing space before checking the word content.
	firstLineTrimmed := strings.TrimRight(lines[0], " ")
	wantFirstLine := "abcdefghijklmnopq rs"
	if firstLineTrimmed != wantFirstLine {
		t.Errorf("first line = %q, want %q\n  full content: %q", firstLineTrimmed, wantFirstLine, content)
	}

	wantSecondLine := "x"
	if lines[1] != wantSecondLine {
		t.Errorf("second line = %q, want %q", lines[1], wantSecondLine)
	}
}
