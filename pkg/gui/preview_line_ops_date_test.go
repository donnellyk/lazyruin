package gui

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/testutil"
)

// These tests cover the SetInlineDate flow on `<c-d>`: picking a date should
// REPLACE any existing inline dates on the line, and the in-popup `<c-x>`
// shortcut should clear all dates without adding one. The flow is exercised
// by opening the popup via the helper, then driving the Config.OnAccept /
// OnCtrlX callbacks directly — that's the same code the keybinding handlers
// call once the user presses Enter / Ctrl-X in the popup.

// dateLineFixture writes a note file whose first content line is `line` and
// seeds the active preview's NavState so resolveTarget() picks that line.
// Returns the gui, the temp dir, and the path of the note file.
type dateLineFixture struct {
	tg     *testGui
	mock   *testutil.MockExecutor
	dir    string
	uuid   string
	path   string
	lineNo int
}

func newDateLineFixture(t *testing.T, line string) *dateLineFixture {
	t.Helper()

	dir := t.TempDir()
	uuid := "n1"
	path := filepath.Join(dir, uuid+".md")
	body := "---\nuuid: " + uuid + "\n---\n" + line
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	mock := defaultMock()
	tg := newTestGui(t, mock)

	ns := tg.gui.contexts.ActivePreview().NavState()
	ns.Lines = []types.SourceLine{
		{Text: line, UUID: uuid, LineNum: 1, Path: path},
	}
	ns.CursorLine = 0

	// Drop the seeded `note` calls (none expected at this point) so each
	// test starts from a clean slate when asserting on mock.Calls.
	mock.Calls = nil

	return &dateLineFixture{tg: tg, mock: mock, dir: dir, uuid: uuid, path: path, lineNo: 1}
}

func TestSetInlineDate_ReplacesExistingDate(t *testing.T) {
	f := newDateLineFixture(t, "Some line @2026-01-01 with a date\n")
	defer f.tg.Close()

	if err := f.tg.gui.helpers.PreviewLineOps().SetInlineDate(); err != nil {
		t.Fatalf("SetInlineDate: %v", err)
	}
	cfg := f.tg.gui.contexts.InputPopup.Config
	if cfg == nil || cfg.OnAccept == nil {
		t.Fatal("popup not opened with OnAccept set")
	}

	if err := cfg.OnAccept("", &types.CompletionItem{InsertText: "@2026-05-01"}); err != nil {
		t.Fatalf("OnAccept: %v", err)
	}

	calls := noteSetCalls(f.mock.Calls)
	if !anyCallContainsAll(calls, []string{"--remove-date", "2026-01-01"}) {
		t.Errorf("expected --remove-date 2026-01-01; calls=%v", calls)
	}
	if !anyCallContainsAll(calls, []string{"--add-date", "2026-05-01"}) {
		t.Errorf("expected --add-date 2026-05-01; calls=%v", calls)
	}
}

func TestSetInlineDate_PickedSameAsExistingIsNoop(t *testing.T) {
	f := newDateLineFixture(t, "Single date @2026-05-01 line\n")
	defer f.tg.Close()

	if err := f.tg.gui.helpers.PreviewLineOps().SetInlineDate(); err != nil {
		t.Fatalf("SetInlineDate: %v", err)
	}
	cfg := f.tg.gui.contexts.InputPopup.Config

	if err := cfg.OnAccept("", &types.CompletionItem{InsertText: "@2026-05-01"}); err != nil {
		t.Fatalf("OnAccept: %v", err)
	}

	calls := noteSetCalls(f.mock.Calls)
	if len(calls) != 0 {
		t.Errorf("expected no `note set` calls when picked == existing; got %v", calls)
	}
}

func TestSetInlineDate_AddsWhenNoExistingDate(t *testing.T) {
	f := newDateLineFixture(t, "Plain line with no date\n")
	defer f.tg.Close()

	if err := f.tg.gui.helpers.PreviewLineOps().SetInlineDate(); err != nil {
		t.Fatalf("SetInlineDate: %v", err)
	}
	cfg := f.tg.gui.contexts.InputPopup.Config

	if err := cfg.OnAccept("", &types.CompletionItem{InsertText: "@2026-05-01"}); err != nil {
		t.Fatalf("OnAccept: %v", err)
	}

	calls := noteSetCalls(f.mock.Calls)
	if anyCallContainsAll(calls, []string{"--remove-date"}) {
		t.Errorf("did not expect --remove-date; calls=%v", calls)
	}
	if !anyCallContainsAll(calls, []string{"--add-date", "2026-05-01"}) {
		t.Errorf("expected --add-date 2026-05-01; calls=%v", calls)
	}
}

func TestSetInlineDate_ReplaceWipesAllExistingDates(t *testing.T) {
	f := newDateLineFixture(t, "Multi-date @2026-01-01 and @2026-02-02 here\n")
	defer f.tg.Close()

	if err := f.tg.gui.helpers.PreviewLineOps().SetInlineDate(); err != nil {
		t.Fatalf("SetInlineDate: %v", err)
	}
	cfg := f.tg.gui.contexts.InputPopup.Config

	if err := cfg.OnAccept("", &types.CompletionItem{InsertText: "@2026-05-01"}); err != nil {
		t.Fatalf("OnAccept: %v", err)
	}

	calls := noteSetCalls(f.mock.Calls)
	got := removedDates(calls)
	sort.Strings(got)
	want := []string{"2026-01-01", "2026-02-02"}
	if !equalStrSlices(got, want) {
		t.Errorf("expected both dates removed; got %v calls=%v", got, calls)
	}
	if !anyCallContainsAll(calls, []string{"--add-date", "2026-05-01"}) {
		t.Errorf("expected --add-date 2026-05-01; calls=%v", calls)
	}
}

func TestSetInlineDate_CtrlXClearsAllDates(t *testing.T) {
	f := newDateLineFixture(t, "Multi-date @2026-01-01 and @2026-02-02 here\n")
	defer f.tg.Close()

	if err := f.tg.gui.helpers.PreviewLineOps().SetInlineDate(); err != nil {
		t.Fatalf("SetInlineDate: %v", err)
	}
	cfg := f.tg.gui.contexts.InputPopup.Config
	if cfg == nil || cfg.OnCtrlX == nil {
		t.Fatal("popup config has no OnCtrlX handler")
	}

	if err := cfg.OnCtrlX(); err != nil {
		t.Fatalf("OnCtrlX: %v", err)
	}

	calls := noteSetCalls(f.mock.Calls)
	got := removedDates(calls)
	sort.Strings(got)
	want := []string{"2026-01-01", "2026-02-02"}
	if !equalStrSlices(got, want) {
		t.Errorf("expected both dates removed via Ctrl-X; got %v calls=%v", got, calls)
	}
	if anyCallContainsAll(calls, []string{"--add-date"}) {
		t.Errorf("Ctrl-X must not add a date; calls=%v", calls)
	}
	if f.tg.gui.contexts.InputPopup.Config != nil {
		t.Error("popup should be closed after Ctrl-X")
	}
}

func TestSetInlineDate_CtrlXNoExistingDatesIsNoop(t *testing.T) {
	f := newDateLineFixture(t, "Plain line no dates\n")
	defer f.tg.Close()

	if err := f.tg.gui.helpers.PreviewLineOps().SetInlineDate(); err != nil {
		t.Fatalf("SetInlineDate: %v", err)
	}
	cfg := f.tg.gui.contexts.InputPopup.Config
	if err := cfg.OnCtrlX(); err != nil {
		t.Fatalf("OnCtrlX: %v", err)
	}
	calls := noteSetCalls(f.mock.Calls)
	if len(calls) != 0 {
		t.Errorf("Ctrl-X with no dates should not call ruin; got %v", calls)
	}
}

// --- helpers ---------------------------------------------------------------

func noteSetCalls(all [][]string) [][]string {
	var out [][]string
	for _, c := range all {
		if len(c) >= 2 && c[0] == "note" && c[1] == "set" {
			out = append(out, c)
		}
	}
	return out
}

func anyCallContainsAll(calls [][]string, want []string) bool {
	for _, c := range calls {
		ok := true
		for _, w := range want {
			found := false
			for _, tok := range c {
				if tok == w {
					found = true
					break
				}
			}
			if !found {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func removedDates(calls [][]string) []string {
	var out []string
	for _, c := range calls {
		for i := 0; i < len(c)-1; i++ {
			if c[i] == "--remove-date" {
				out = append(out, c[i+1])
			}
		}
	}
	return out
}

func equalStrSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
