package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
)

// writeNote creates a note file at tmpDir/name with the given body and
// returns the absolute path.
func writeNote(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestSubmitCapture_BareURL_RoutesToLinkResolve: when a New Note's entire
// content is a valid URL, saving should route through the link-resolution
// flow (`ruin link resolve` → `ruin link new`) instead of `ruin log`, so
// the user gets the resolved title/description + #link tag without
// retyping in the New Link popup.
func TestSubmitCapture_BareURL_RoutesToLinkResolve(t *testing.T) {
	mock := defaultMock().WithLinkJSON([]byte(
		`{"title":"Example Title","summary":"An example page."}`))
	tg := newTestGui(t, mock)
	defer tg.Close()

	err := tg.gui.helpers.Capture().SubmitCapture("https://example.com/foo", false)
	if err != nil {
		t.Fatalf("SubmitCapture: %v", err)
	}

	// No `ruin log` call (the New Note path) should have fired.
	for _, call := range mock.Calls {
		if len(call) > 0 && call[0] == "log" {
			t.Errorf("unexpected `ruin log` call on bare-URL submit; want the link flow instead. calls=%v", mock.Calls)
			break
		}
	}

	// A link resolve call should have fired (async), so drive layout a
	// few times to let the goroutine settle.
	for i := 0; i < 10; i++ {
		_ = tg.g.ForceLayoutAndRedraw()
		time.Sleep(10 * time.Millisecond)
	}

	var sawResolve bool
	for _, call := range mock.Calls {
		if len(call) >= 3 && call[0] == "link" && call[1] == "resolve" {
			sawResolve = true
			break
		}
	}
	if !sawResolve {
		t.Errorf("expected a `ruin link resolve` call after bare-URL submit; calls=%v", mock.Calls)
	}
}

// TestSubmitCapture_BareURL_DisabledByConfig: when
// DisableBareURLAsLink is set, submitting a URL-only body takes the
// plain `ruin log` path instead of the link-resolve flow.
func TestSubmitCapture_BareURL_DisabledByConfig(t *testing.T) {
	mock := defaultMock()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.config.DisableBareURLAsLink = true

	if err := tg.gui.helpers.Capture().SubmitCapture("https://example.com/foo", false); err != nil {
		t.Fatalf("SubmitCapture: %v", err)
	}

	var sawLog, sawResolve bool
	for _, call := range mock.Calls {
		switch {
		case len(call) >= 2 && call[0] == "log" && call[1] == "https://example.com/foo":
			sawLog = true
		case len(call) >= 3 && call[0] == "link" && call[1] == "resolve":
			sawResolve = true
		}
	}
	if !sawLog {
		t.Errorf("expected `ruin log` call when bare-URL routing is disabled; calls=%v", mock.Calls)
	}
	if sawResolve {
		t.Errorf("did not expect a `ruin link resolve` call when disabled; calls=%v", mock.Calls)
	}
}

func TestOpenCaptureForEdit_DismissesCompletionAfterPrefill(t *testing.T) {
	// Regression: when the prefilled content ends with a trigger character
	// (e.g. a note whose body ends in `#followup`), the layout's TypeString
	// must leave completion state dismissed so the user's first Esc closes
	// the popup rather than just dismissing a stale dropdown.
	dir := t.TempDir()
	notePath := writeNote(t, dir, "tagged.md", "---\nuuid: tag-1\n---\n\nsome content #followup\n")

	mock := defaultMock().WithNotes(
		models.Note{UUID: "tag-1", Title: "Tagged Note", Path: notePath, Created: time.Now()},
	)
	tg := newTestGui(t, mock)
	defer tg.Close()

	note := tg.gui.contexts.Notes.Items[0]
	if err := tg.gui.helpers.Capture().OpenCaptureForEdit(&note); err != nil {
		t.Fatalf("OpenCaptureForEdit: %v", err)
	}
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout: %v", err)
	}

	// PrefillContent should have been consumed by layout and the completion
	// state should be dismissed, even though the prefilled body ends with `#followup`.
	if tg.gui.contexts.Capture.Completion.Active {
		t.Errorf("completion should be dismissed after prefill; items: %v",
			labelsFromCompletionItems(tg.gui.contexts.Capture.Completion.Items))
	}
}

// labelsFromCompletionItems mirrors `labels` in the embed integration test —
// duplicated here to keep this file self-contained.
func labelsFromCompletionItems(items []types.CompletionItem) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Label)
	}
	return out
}

func TestOpenCaptureForEdit_PopulatesContent(t *testing.T) {
	dir := t.TempDir()
	notePath := writeNote(t, dir, "my-note.md", "---\nuuid: edit-1\n---\n\n# My Note\n\nOriginal body.\n")

	mock := defaultMock().WithNotes(
		models.Note{UUID: "edit-1", Title: "My Note", Path: notePath, Created: time.Now()},
	)
	tg := newTestGui(t, mock)
	defer tg.Close()

	note := tg.gui.contexts.Notes.Items[0]
	if err := tg.gui.helpers.Capture().OpenCaptureForEdit(&note); err != nil {
		t.Fatalf("OpenCaptureForEdit: %v", err)
	}

	if tg.gui.contextMgr.Current() != "capture" {
		t.Errorf("context = %v, want capture", tg.gui.contextMgr.Current())
	}
	ctx := tg.gui.contexts.Capture
	if ctx.EditingPath != notePath {
		t.Errorf("EditingPath = %q, want %q", ctx.EditingPath, notePath)
	}
	if ctx.EditingTitle != "My Note" {
		t.Errorf("EditingTitle = %q, want %q", ctx.EditingTitle, "My Note")
	}
	if ctx.EditingMtime.IsZero() {
		t.Error("EditingMtime should be captured at open time")
	}
	want := "# My Note\n\nOriginal body.\n"
	if ctx.PrefillContent != want {
		t.Errorf("PrefillContent = %q, want %q", ctx.PrefillContent, want)
	}
}

func TestCloseCapture_ResetsEditingState(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	ctx := tg.gui.contexts.Capture
	ctx.EditingPath = "/tmp/foo.md"
	ctx.EditingTitle = "Foo"
	ctx.EditingMtime = time.Now()

	tg.gui.helpers.Capture().OpenCapture()
	tg.gui.helpers.Capture().CloseCapture()

	if ctx.EditingPath != "" {
		t.Errorf("EditingPath not reset: %q", ctx.EditingPath)
	}
	if ctx.EditingTitle != "" {
		t.Errorf("EditingTitle not reset: %q", ctx.EditingTitle)
	}
	if !ctx.EditingMtime.IsZero() {
		t.Errorf("EditingMtime not reset: %v", ctx.EditingMtime)
	}
}

func TestSubmitCapture_EditModeWritesFile(t *testing.T) {
	dir := t.TempDir()
	notePath := writeNote(t, dir, "edit-save.md", "---\nuuid: save-1\n---\n\nold content\n")

	mock := defaultMock().WithNotes(
		models.Note{UUID: "save-1", Title: "Edit Save", Path: notePath, Created: time.Now()},
	)
	tg := newTestGui(t, mock)
	defer tg.Close()

	note := tg.gui.contexts.Notes.Items[0]
	if err := tg.gui.helpers.Capture().OpenCaptureForEdit(&note); err != nil {
		t.Fatalf("OpenCaptureForEdit: %v", err)
	}

	newContent := "brand new body\nwith two lines"
	if err := tg.gui.helpers.Capture().SubmitCapture(newContent, false); err != nil {
		t.Fatalf("SubmitCapture: %v", err)
	}

	got, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatal(err)
	}
	want := "---\nuuid: save-1\n---\n\nbrand new body\nwith two lines\n"
	if string(got) != want {
		t.Errorf("file content:\n%q\nwant:\n%q", string(got), want)
	}
	if tg.gui.contexts.Capture.EditingPath != "" {
		t.Error("EditingPath should be reset after SubmitCapture")
	}
}

func TestSubmitCapture_EmptyEdit_TreatedAsCancel(t *testing.T) {
	dir := t.TempDir()
	original := "---\nuuid: cancel-1\n---\n\nimportant body\n"
	notePath := writeNote(t, dir, "cancel.md", original)

	mock := defaultMock().WithNotes(
		models.Note{UUID: "cancel-1", Title: "Do Not Wipe", Path: notePath, Created: time.Now()},
	)
	tg := newTestGui(t, mock)
	defer tg.Close()

	note := tg.gui.contexts.Notes.Items[0]
	if err := tg.gui.helpers.Capture().OpenCaptureForEdit(&note); err != nil {
		t.Fatalf("OpenCaptureForEdit: %v", err)
	}

	// User clears the buffer and submits — should be treated as cancel,
	// not a body-wipe save.
	if err := tg.gui.helpers.Capture().SubmitCapture("", false); err != nil {
		t.Fatalf("SubmitCapture: %v", err)
	}

	got, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Errorf("file content changed — empty-save should be cancel\ngot:\n%q\nwant:\n%q", string(got), original)
	}

	// Whitespace-only content is also treated as cancel.
	if err := tg.gui.helpers.Capture().OpenCaptureForEdit(&note); err != nil {
		t.Fatalf("OpenCaptureForEdit (2): %v", err)
	}
	if err := tg.gui.helpers.Capture().SubmitCapture("   \n  \n", false); err != nil {
		t.Fatalf("SubmitCapture (2): %v", err)
	}
	got2, _ := os.ReadFile(notePath)
	if string(got2) != original {
		t.Errorf("whitespace-only save should be cancel, got:\n%q", string(got2))
	}
}

func TestSubmitCapture_MtimeConflict_KeepsPopupOpen(t *testing.T) {
	dir := t.TempDir()
	notePath := writeNote(t, dir, "conflict.md", "---\nuuid: conflict-1\n---\n\noriginal body\n")

	mock := defaultMock().WithNotes(
		models.Note{UUID: "conflict-1", Title: "Conflict Test", Path: notePath, Created: time.Now()},
	)
	tg := newTestGui(t, mock)
	defer tg.Close()

	note := tg.gui.contexts.Notes.Items[0]
	if err := tg.gui.helpers.Capture().OpenCaptureForEdit(&note); err != nil {
		t.Fatalf("OpenCaptureForEdit: %v", err)
	}

	// Simulate an external edit: rewrite the file, then touch mtime forward.
	externallyChanged := "---\nuuid: conflict-1\n---\n\nchanged by another process\n"
	if err := os.WriteFile(notePath, []byte(externallyChanged), 0644); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(notePath, future, future); err != nil {
		t.Fatal(err)
	}

	err := tg.gui.helpers.Capture().SubmitCapture("my in-progress edits", false)
	if err != nil {
		t.Fatalf("SubmitCapture returned error: %v (it should show an error in-GUI and return nil)", err)
	}

	// Popup should remain open so the user can recover their edits.
	if tg.gui.contextMgr.Current() != "capture" {
		t.Errorf("popup should stay open on conflict, current context: %v", tg.gui.contextMgr.Current())
	}
	if tg.gui.contexts.Capture.EditingPath == "" {
		t.Error("EditingPath should remain set on conflict")
	}

	// File on disk should be unchanged from the external edit.
	got, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != externallyChanged {
		t.Errorf("file should be untouched on conflict, got:\n%q", string(got))
	}
}

func TestSubmitCapture_WriteError_KeepsPopupOpen(t *testing.T) {
	dir := t.TempDir()
	notePath := writeNote(t, dir, "readonly.md", "---\nuuid: ro-1\n---\n\noriginal\n")

	mock := defaultMock().WithNotes(
		models.Note{UUID: "ro-1", Title: "Read Only", Path: notePath, Created: time.Now()},
	)
	tg := newTestGui(t, mock)
	defer tg.Close()

	note := tg.gui.contexts.Notes.Items[0]
	if err := tg.gui.helpers.Capture().OpenCaptureForEdit(&note); err != nil {
		t.Fatalf("OpenCaptureForEdit: %v", err)
	}

	// Make the directory read-only so temp-file creation fails. This simulates
	// a disk-full / permission-denied condition for the atomic write path.
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0755) // allow TempDir cleanup
	})

	err := tg.gui.helpers.Capture().SubmitCapture("my precious edits", false)
	if err != nil {
		t.Fatalf("SubmitCapture returned error: %v (it should show an error in-GUI and return nil)", err)
	}
	if tg.gui.contextMgr.Current() != "capture" {
		t.Errorf("popup should stay open on write failure, current context: %v", tg.gui.contextMgr.Current())
	}
}

func TestSaveEdit_AtomicWrite_LeavesNoTempOnSuccess(t *testing.T) {
	dir := t.TempDir()
	notePath := writeNote(t, dir, "atomic.md", "---\nuuid: atomic-1\n---\n\nold\n")

	mock := defaultMock().WithNotes(
		models.Note{UUID: "atomic-1", Title: "Atomic", Path: notePath, Created: time.Now()},
	)
	tg := newTestGui(t, mock)
	defer tg.Close()

	note := tg.gui.contexts.Notes.Items[0]
	if err := tg.gui.helpers.Capture().OpenCaptureForEdit(&note); err != nil {
		t.Fatalf("OpenCaptureForEdit: %v", err)
	}
	if err := tg.gui.helpers.Capture().SubmitCapture("new body", false); err != nil {
		t.Fatalf("SubmitCapture: %v", err)
	}

	// After a successful save, the temp file must be gone.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("temp file %q still present after successful save", e.Name())
		}
	}
}
