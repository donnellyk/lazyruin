package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/donnellyk/lazyruin/pkg/testutil"
)

// composeFixture seeds a real on-disk parent+children layout, builds a
// matching compose JSON payload, and returns a mock configured to drive
// newTestGui into compose mode on startup.
type composeFixture struct {
	dir        string
	childAPath string
	childBPath string
	mock       *testutil.MockExecutor
}

func newComposeFixture(t *testing.T) composeFixture {
	t.Helper()
	dir := t.TempDir()

	childAPath := filepath.Join(dir, "child-a.md")
	if err := os.WriteFile(childAPath, []byte(
		"---\nuuid: child-a-uuid\n---\n"+
			"# Child A Title\n"+
			"\n"+
			"Content line A1\n"+
			"- Task A #todo\n"+
			"Content line A3\n",
	), 0644); err != nil {
		t.Fatalf("write child-a: %v", err)
	}

	childBPath := filepath.Join(dir, "child-b.md")
	if err := os.WriteFile(childBPath, []byte(
		"---\nuuid: child-b-uuid\n---\n"+
			"# Child B Title\n"+
			"\n"+
			"Content line B1\n"+
			"- Task B #todo\n",
	), 0644); err != nil {
		t.Fatalf("write child-b: %v", err)
	}

	// Composed content: child A body (3 lines), blank, child B as H2 + body.
	composedContent := "Content line A1\n" +
		"- Task A #todo\n" +
		"Content line A3\n" +
		"\n" +
		"## Child B Title\n" +
		"\n" +
		"Content line B1\n" +
		"- Task B #todo"

	composeJSON, err := json.Marshal(map[string]any{
		"uuid":             "parent-1",
		"title":            "Daily Journal",
		"path":             filepath.Join(dir, "parent.md"),
		"composed_content": composedContent,
		"source_map": []models.SourceMapEntry{
			{UUID: "child-a-uuid", Path: childAPath, Title: "Child A Title", StartLine: 1, EndLine: 3},
			{UUID: "child-b-uuid", Path: childBPath, Title: "Child B Title", StartLine: 5, EndLine: 8},
		},
	})
	if err != nil {
		t.Fatalf("marshal compose JSON: %v", err)
	}

	mock := testutil.NewMockExecutor().
		WithNotes(
			models.Note{UUID: "child-a-uuid", Title: "Child A Title", Path: childAPath, Created: time.Now()},
			models.Note{UUID: "child-b-uuid", Title: "Child B Title", Path: childBPath, Created: time.Now()},
		).
		WithParents(
			models.ParentBookmark{Name: "journal", UUID: "parent-1", Title: "Daily Journal"},
		).
		WithCompose(composeJSON)

	return composeFixture{dir: dir, childAPath: childAPath, childBPath: childBPath, mock: mock}
}

// cursorLineWithUUID positions the compose cursor on the first SourceLine
// whose UUID matches the target. Fails the test if no such line exists —
// the fixture is expected to resolve every child-body line to its UUID.
func cursorLineWithUUID(t *testing.T, tg *testGui, uuid string) int {
	t.Helper()
	ns := tg.gui.contexts.Compose.NavState()
	for i, sl := range ns.Lines {
		if sl.UUID == uuid && sl.LineNum > 0 {
			ns.CursorLine = i
			return i
		}
	}
	t.Fatalf("no SourceLine attributed to UUID %q (lines=%d)", uuid, len(ns.Lines))
	return -1
}

// invokeComposeBinding finds the binding with the given ID on the compose
// context and invokes its handler. Used to exercise compose controller
// actions without reaching into the headless gocui keybinding dispatch.
func invokeComposeBinding(t *testing.T, tg *testGui, id string) error {
	t.Helper()
	opts := types.KeybindingsOpts{}
	for _, b := range tg.gui.contexts.Compose.GetKeybindings(opts) {
		if b.ID == id {
			return b.Handler()
		}
	}
	t.Fatalf("compose binding %q not registered", id)
	return nil
}

// countCalls returns the number of mock Execute invocations whose first arg
// matches cmd.
func countCalls(mock *testutil.MockExecutor, cmd string) int {
	n := 0
	for _, args := range mock.Calls {
		if len(args) > 0 && args[0] == cmd {
			n++
		}
	}
	return n
}

func TestComposeEnter_OpensChildInCardList(t *testing.T) {
	fx := newComposeFixture(t)
	tg := newTestGuiWithOpts(t, fx.mock, testGuiOpts{OpenRef: "journal"})
	defer tg.Close()

	if tg.gui.contextMgr.Current() != "compose" {
		t.Fatalf("precondition: CurrentContext = %v, want compose", tg.gui.contextMgr.Current())
	}

	cursorLineWithUUID(t, tg, "child-b-uuid")

	if err := tg.gui.helpers.PreviewNav().PreviewEnter(); err != nil {
		t.Fatalf("PreviewEnter: %v", err)
	}

	if tg.gui.contextMgr.Current() != "cardList" {
		t.Errorf("CurrentContext = %v, want cardList", tg.gui.contextMgr.Current())
	}
	cards := tg.gui.contexts.CardList.Cards
	if len(cards) != 1 {
		t.Fatalf("CardList.Cards = %d, want 1", len(cards))
	}
	if cards[0].UUID != "child-b-uuid" {
		t.Errorf("CardList.Cards[0].UUID = %q, want child-b-uuid", cards[0].UUID)
	}
}

func TestComposeEnter_OnUnresolvableLine_IsNoOp(t *testing.T) {
	fx := newComposeFixture(t)
	tg := newTestGuiWithOpts(t, fx.mock, testGuiOpts{OpenRef: "journal"})
	defer tg.Close()

	if tg.gui.contextMgr.Current() != "compose" {
		t.Fatalf("precondition: CurrentContext = %v, want compose", tg.gui.contextMgr.Current())
	}

	// Find a line with no source identity (card separators and the blank
	// line between children carry an empty UUID or a zero LineNum).
	ns := tg.gui.contexts.Compose.NavState()
	placed := false
	for i, sl := range ns.Lines {
		if sl.UUID == "" || sl.LineNum == 0 {
			ns.CursorLine = i
			placed = true
			break
		}
	}
	if !placed {
		t.Fatalf("no unresolvable line in fixture (lines=%d)", len(ns.Lines))
	}

	if err := tg.gui.helpers.PreviewNav().PreviewEnter(); err != nil {
		t.Fatalf("PreviewEnter: %v", err)
	}

	if tg.gui.contextMgr.Current() != "compose" {
		t.Errorf("CurrentContext = %v, want compose (Enter should no-op on unresolvable line)", tg.gui.contextMgr.Current())
	}
}

func TestComposeEditInline_OpensCaptureForChild(t *testing.T) {
	fx := newComposeFixture(t)
	tg := newTestGuiWithOpts(t, fx.mock, testGuiOpts{OpenRef: "journal"})
	defer tg.Close()

	cursorLineWithUUID(t, tg, "child-a-uuid")

	if err := invokeComposeBinding(t, tg, "compose.edit_inline"); err != nil {
		t.Fatalf("invoke compose.edit_inline: %v", err)
	}

	if tg.gui.contextMgr.Current() != "capture" {
		t.Errorf("CurrentContext = %v, want capture", tg.gui.contextMgr.Current())
	}
	if got := tg.gui.contexts.Capture.EditingPath; got != fx.childAPath {
		t.Errorf("Capture.EditingPath = %q, want %q", got, fx.childAPath)
	}
	if got := tg.gui.contexts.Capture.EditingTitle; got != "Child A Title" {
		t.Errorf("Capture.EditingTitle = %q, want %q", got, "Child A Title")
	}
}

func TestReloadActivePreview_Compose_ReRunsParentCompose(t *testing.T) {
	fx := newComposeFixture(t)
	tg := newTestGuiWithOpts(t, fx.mock, testGuiOpts{OpenRef: "journal"})
	defer tg.Close()

	if tg.gui.contexts.ActivePreviewKey != "compose" {
		t.Fatalf("precondition: ActivePreviewKey = %v, want compose", tg.gui.contexts.ActivePreviewKey)
	}

	before := countCalls(fx.mock, "compose")
	tg.gui.helpers.Preview().ReloadActivePreview()
	after := countCalls(fx.mock, "compose")

	if after <= before {
		t.Errorf("ReloadActivePreview did not re-run compose (before=%d after=%d)", before, after)
	}
}
