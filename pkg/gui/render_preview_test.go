package gui

import (
	"strings"
	"testing"

	"github.com/donnellyk/lazyruin/pkg/config"
)

// TestWrapText_HyphenOverflow ensures the hyphen-counting bug in
// muesli/reflow/wordwrap (which emits lines longer than the requested
// limit when breakpoint runes appear on the line) does not leak through.
// Each output line must respect the limit so the preview pane's right
// padding column is not eaten by overflowing content.
func TestWrapText_HyphenOverflow(t *testing.T) {
	tests := []struct {
		text  string
		limit int
	}{
		{"abcd-efgh", 8},
		{"abcd-efgh-ijkl", 8},
		{"foo-bar baz qux 12345678901234567890123456-abc", 44},
	}
	for _, tt := range tests {
		got := wrapText(tt.text, tt.limit)
		for _, line := range strings.Split(got, "\n") {
			if width := visibleWidth(line); width > tt.limit {
				t.Errorf("wrapText(%q, %d) line %q exceeds limit (%d > %d)",
					tt.text, tt.limit, line, width, tt.limit)
			}
		}
	}
}

func TestWrapText_PreservesAnsi(t *testing.T) {
	// chroma-style ANSI escapes wrapping a single visible word should not
	// be split (visible width 5 fits in limit 10).
	in := "\x1b[31mhello\x1b[0m world"
	got := wrapText(in, 10)
	if !strings.Contains(got, "\x1b[31mhello\x1b[0m") {
		t.Errorf("ANSI escape was split or stripped: got %q", got)
	}
}

func TestApplyPreviewPadding(t *testing.T) {
	tests := []struct {
		name             string
		pad              int
		width            int
		contentWidth     int
		wantPrefix       string
		wantWidth        int
		wantContentWidth int
	}{
		{"zero pad is a no-op", 0, 80, 78, "", 80, 78},
		{"pad shrinks both widths and returns prefix", 2, 80, 78, "  ", 76, 74},
		{"floor caps shrunk widths at 10", 5, 18, 15, "     ", 10, 10},
		{"negative pad treated as zero", -3, 80, 78, "", 80, 78},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gui := &Gui{config: &config.Config{PreviewPadding: tt.pad}}
			prefix, w, cw := gui.applyPreviewPadding(tt.width, tt.contentWidth)
			if prefix != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", prefix, tt.wantPrefix)
			}
			if w != tt.wantWidth {
				t.Errorf("width = %d, want %d", w, tt.wantWidth)
			}
			if cw != tt.wantContentWidth {
				t.Errorf("contentWidth = %d, want %d", cw, tt.wantContentWidth)
			}
		})
	}
}

func TestApplyPreviewPadding_NilConfig(t *testing.T) {
	gui := &Gui{}
	prefix, w, cw := gui.applyPreviewPadding(80, 78)
	if prefix != "" || w != 80 || cw != 78 {
		t.Errorf("nil config should be no-op; got prefix=%q w=%d cw=%d", prefix, w, cw)
	}
}
