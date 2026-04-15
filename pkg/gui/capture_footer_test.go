package gui

import (
	"strings"
	"testing"
)

// TestCaptureFooter_MatchesCLIExtraction verifies that the tag list shown in
// the capture popup footer is produced by the shared notetext extractor, not
// the TUI's old naïve regex. In particular:
//
//   - tags inside inline code spans are ignored (CLI wouldn't record them)
//   - tags inside markdown links are ignored
//   - tags with slashes (#date/2026) are captured in full, not truncated
//
// Regression: the old InlineTagRe-based extractor drifted from what `ruin log`
// actually stored, so the footer gave misleading feedback while drafting.
func TestCaptureFooter_MatchesCLIExtraction(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	v := tg.gui.views.Capture
	if v == nil {
		t.Fatal("Capture view should exist")
	}

	v.TextArea.TypeString("real #project note; skip `#codetag`; link [txt #notag](https://x) and #date/2026")
	tg.gui.updateCaptureFooter()

	footer := tg.gui.views.Capture.Footer

	for _, want := range []string{"#project", "#date/2026"} {
		if !strings.Contains(footer, want) {
			t.Errorf("footer missing %q; got %q", want, footer)
		}
	}
	for _, unwanted := range []string{"#codetag", "#notag"} {
		if strings.Contains(footer, unwanted) {
			t.Errorf("footer should not contain %q (code/link scope); got %q", unwanted, footer)
		}
	}
}
