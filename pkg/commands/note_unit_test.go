package commands

import (
	"fmt"
	"testing"
)

func TestNoteCommand_Delete(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.Delete("abc-123")

	args := cap.lastArgs()
	assertArgsContain(t, args, "note")
	assertArgsContain(t, args, "delete")
	assertArgsContain(t, args, "abc-123")
	assertArgsContain(t, args, "-f")
}

func TestNoteCommand_ToggleTodo(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.ToggleTodo("abc-123", 5)

	args := cap.lastArgs()
	assertArgsContain(t, args, "note")
	assertArgsContain(t, args, "set")
	assertArgsContain(t, args, "abc-123")
	assertArgsContain(t, args, "--toggle-todo")
	assertArgsContain(t, args, "--line")
	assertArgsContain(t, args, "5")
	assertArgsContain(t, args, "--sink")
	assertArgsContain(t, args, "-f")
}

func TestNoteCommand_SetParent(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.SetParent("child-uuid", "parent-uuid")

	args := cap.lastArgs()
	assertArgsContain(t, args, "note")
	assertArgsContain(t, args, "set")
	assertArgsContain(t, args, "child-uuid")
	assertArgsContain(t, args, "--parent")
	assertArgsContain(t, args, "parent-uuid")
	assertArgsContain(t, args, "-f")
}

func TestNoteCommand_RemoveParent(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.RemoveParent("abc-123")

	args := cap.lastArgs()
	assertArgsContain(t, args, "--no-parent")
	assertArgsContain(t, args, "-f")
}

func TestNoteCommand_AddTag(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.AddTag("abc-123", "#meeting")

	args := cap.lastArgs()
	assertArgsContain(t, args, "--add-tag")
	assertArgsContain(t, args, "#meeting")
}

func TestNoteCommand_RemoveTag(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.RemoveTag("abc-123", "#old")

	args := cap.lastArgs()
	assertArgsContain(t, args, "--remove-tag")
	assertArgsContain(t, args, "#old")
}

func TestNoteCommand_AddTagToLine(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.AddTagToLine("abc-123", "#done", 3)

	args := cap.lastArgs()
	assertArgsContain(t, args, "--add-tag")
	assertArgsContain(t, args, "#done")
	assertArgsContain(t, args, "--line")
	assertArgsContain(t, args, "3")
}

func TestNoteCommand_RemoveTagFromLine(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.RemoveTagFromLine("abc-123", "#done", 3)

	args := cap.lastArgs()
	assertArgsContain(t, args, "--remove-tag")
	assertArgsContain(t, args, "#done")
	assertArgsContain(t, args, "--line")
	assertArgsContain(t, args, "3")
}

func TestNoteCommand_AddDateToLine(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.AddDateToLine("abc-123", "2026-03-01", 2)

	args := cap.lastArgs()
	assertArgsContain(t, args, "--add-date")
	assertArgsContain(t, args, "2026-03-01")
	assertArgsContain(t, args, "--line")
	assertArgsContain(t, args, "2")
}

func TestNoteCommand_RemoveDateFromLine(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.RemoveDateFromLine("abc-123", "2026-03-01", 2)

	args := cap.lastArgs()
	assertArgsContain(t, args, "--remove-date")
	assertArgsContain(t, args, "2026-03-01")
	assertArgsContain(t, args, "--line")
	assertArgsContain(t, args, "2")
}

func TestNoteCommand_SetOrder(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.SetOrder("abc-123", 42)

	args := cap.lastArgs()
	assertArgsContain(t, args, "--order")
	assertArgsContain(t, args, "42")
}

func TestNoteCommand_RemoveOrder(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.RemoveOrder("abc-123")

	args := cap.lastArgs()
	assertArgsContain(t, args, "--no-order")
	assertArgsContain(t, args, "-f")
}

func TestNoteCommand_SetField(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.SetField("abc-123", "status", "active")

	args := cap.lastArgs()
	assertArgsContain(t, args, "--field")
	assertArgsContain(t, args, "status=active")
}

func TestNoteCommand_Append_AtEnd(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.Append("abc-123", "new text", 0, false)

	args := cap.lastArgs()
	assertArgsContain(t, args, "note")
	assertArgsContain(t, args, "append")
	assertArgsContain(t, args, "abc-123")
	assertArgsContain(t, args, "new text")
	assertArgsNotContain(t, args, "--line")
	assertArgsNotContain(t, args, "--suffix")
	assertArgsContain(t, args, "-f")
}

func TestNoteCommand_Append_AtLine(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.Append("abc-123", "inserted", 3, false)

	args := cap.lastArgs()
	assertArgsContain(t, args, "--line")
	assertArgsContain(t, args, "3")
	assertArgsNotContain(t, args, "--suffix")
}

func TestNoteCommand_Append_Suffix(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_ = ruin.Note.Append("abc-123", " appended", 2, true)

	args := cap.lastArgs()
	assertArgsContain(t, args, "--suffix")
	assertArgsContain(t, args, "--line")
	assertArgsContain(t, args, "2")
}

func TestNoteCommand_Merge_Basic(t *testing.T) {
	cap := &argCapture{}
	cap.response = []byte(`{"target_uuid":"t","source_uuid":"s","source_deleted":false}`)
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	result, err := ruin.Note.Merge("target", "source", false, false)

	if err != nil {
		t.Fatalf("Merge error: %v", err)
	}
	args := cap.lastArgs()
	assertArgsContain(t, args, "note")
	assertArgsContain(t, args, "merge")
	assertArgsContain(t, args, "target")
	assertArgsContain(t, args, "source")
	assertArgsNotContain(t, args, "--delete-source")
	assertArgsNotContain(t, args, "--strip-title")
	assertArgsContain(t, args, "-f")

	if result.TargetUUID != "t" {
		t.Errorf("TargetUUID = %q, want %q", result.TargetUUID, "t")
	}
}

func TestNoteCommand_Merge_WithFlags(t *testing.T) {
	cap := &argCapture{}
	cap.response = []byte(`{"target_uuid":"t","source_uuid":"s","source_deleted":true}`)
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	result, err := ruin.Note.Merge("target", "source", true, true)

	if err != nil {
		t.Fatalf("Merge error: %v", err)
	}
	args := cap.lastArgs()
	assertArgsContain(t, args, "--delete-source")
	assertArgsContain(t, args, "--strip-title")

	if !result.SourceDeleted {
		t.Error("expected SourceDeleted to be true")
	}
}

func TestNoteCommand_Merge_InvalidJSON(t *testing.T) {
	cap := &argCapture{}
	cap.response = []byte(`not json`)
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, err := ruin.Note.Merge("target", "source", false, false)
	if err == nil {
		t.Error("expected JSON parse error")
	}
}

func TestNoteCommand_ExecutorError(t *testing.T) {
	cap := &argCapture{}
	cap.err = fmt.Errorf("vault not found")
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	err := ruin.Note.Delete("abc-123")
	if err == nil {
		t.Error("expected error from executor")
	}
	if err.Error() != "vault not found" {
		t.Errorf("error = %q, want %q", err.Error(), "vault not found")
	}
}
