package onboarding

import (
	"strings"
	"testing"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/donnellyk/lazyruin/pkg/testutil"
)

func newMockRuin(mock *testutil.MockExecutor) *commands.RuinCommand {
	return commands.NewRuinCommandWithExecutor(mock, "/mock")
}

func TestContent_IncludesSentinelTag(t *testing.T) {
	if !strings.Contains(Content(), "#"+Tag) {
		t.Errorf("onboarding content is missing #%s tag; cleanup will not find it", Tag)
	}
}

func TestContent_ReferencesCleanupCommand(t *testing.T) {
	if !strings.Contains(Content(), "cleanup") {
		t.Error("onboarding content should tell the user how to clean up")
	}
}

func TestIsVaultEmpty_Empty(t *testing.T) {
	mock := testutil.NewMockExecutor() // no notes
	empty, err := IsVaultEmpty(newMockRuin(mock))
	if err != nil {
		t.Fatal(err)
	}
	if !empty {
		t.Error("IsVaultEmpty = false on zero-note vault, want true")
	}
}

func TestIsVaultEmpty_NotEmpty(t *testing.T) {
	mock := testutil.NewMockExecutor().WithNotes(models.Note{UUID: "u1"})
	empty, err := IsVaultEmpty(newMockRuin(mock))
	if err != nil {
		t.Fatal(err)
	}
	if empty {
		t.Error("IsVaultEmpty = true on vault with one note, want false")
	}
}

func TestCreateNote_CallsRuinLogWithContent(t *testing.T) {
	mock := testutil.NewMockExecutor()
	if err := CreateNote(newMockRuin(mock)); err != nil {
		t.Fatal(err)
	}
	if len(mock.Calls) == 0 {
		t.Fatal("no ruin command was invoked")
	}
	found := false
	for _, call := range mock.Calls {
		if len(call) >= 2 && call[0] == "log" && strings.Contains(call[1], "#"+Tag) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected `ruin log` call with #%s tag in content; calls=%v", Tag, mock.Calls)
	}
}

func TestCleanup_DeletesMatchingNotesAndTag(t *testing.T) {
	mock := testutil.NewMockExecutor().WithNotes(
		models.Note{UUID: "onboarding-uuid-1", Tags: []string{Tag}},
		models.Note{UUID: "onboarding-uuid-2", Tags: []string{Tag}},
	)

	deleted, err := Cleanup(newMockRuin(mock))
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 2 {
		t.Errorf("Cleanup deleted %d notes, want 2", deleted)
	}

	var deleteCalls, tagDeleteCalls, searchCalls int
	for _, call := range mock.Calls {
		switch {
		case len(call) >= 3 && call[0] == "note" && call[1] == "delete":
			deleteCalls++
		case len(call) >= 3 && call[0] == "tags" && call[1] == "delete" && call[2] == Tag:
			tagDeleteCalls++
		case len(call) >= 2 && call[0] == "search" && call[1] == "#"+Tag:
			searchCalls++
		}
	}
	if searchCalls != 1 {
		t.Errorf("expected exactly 1 search call for #%s, got %d", Tag, searchCalls)
	}
	if deleteCalls != 2 {
		t.Errorf("expected 2 note delete calls, got %d", deleteCalls)
	}
	if tagDeleteCalls != 1 {
		t.Errorf("expected 1 tags delete call for %q, got %d", Tag, tagDeleteCalls)
	}
}

func TestCleanup_DedupesRepeatedUUIDs(t *testing.T) {
	// Defensive: if search ever returns the same note twice (e.g. due to a
	// bug or a hard-linked file), cleanup must not attempt to delete the
	// same note twice.
	mock := testutil.NewMockExecutor().WithNotes(
		models.Note{UUID: "same-uuid", Tags: []string{Tag}},
		models.Note{UUID: "same-uuid", Tags: []string{Tag}},
	)
	deleted, err := Cleanup(newMockRuin(mock))
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Errorf("Cleanup deleted %d notes, want 1 (deduped)", deleted)
	}
}

func TestCleanup_EmptyVault(t *testing.T) {
	mock := testutil.NewMockExecutor() // no notes
	deleted, err := Cleanup(newMockRuin(mock))
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 0 {
		t.Errorf("Cleanup on empty vault deleted %d notes, want 0", deleted)
	}
}
