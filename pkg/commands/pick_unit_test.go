package commands

import (
	"strings"
	"testing"
)

type argCapture struct {
	calls [][]string
}

func (a *argCapture) Execute(args ...string) ([]byte, error) {
	a.calls = append(a.calls, args)
	return []byte("[]"), nil
}

func (a *argCapture) VaultPath() string { return "/mock" }

func (a *argCapture) lastArgs() []string {
	if len(a.calls) == 0 {
		return nil
	}
	return a.calls[len(a.calls)-1]
}

func TestPickCommand_DateAsPositionalArg(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick([]string{"#followup"}, PickOpts{Date: "@2026-02-23"})

	args := cap.lastArgs()
	assertArgsContain(t, args, "@2026-02-23")
	assertArgsNotContain(t, args, "--filter")
}

func TestPickCommand_DateOnlyNoTags(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick(nil, PickOpts{Date: "@today"})

	args := cap.lastArgs()
	assertArgsContain(t, args, "@today")
	if len(args) < 2 {
		t.Fatalf("expected at least 2 args, got %v", args)
	}
	// @today should be positional (right after "pick"), not behind --filter
	if args[1] != "@today" {
		t.Errorf("expected @today as first arg after 'pick', got args: %v", args)
	}
}

func TestPickCommand_FilterStaysAsFlag(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick([]string{"#todo"}, PickOpts{Filter: "@tomorrow"})

	args := cap.lastArgs()
	idx := indexOf(args, "--filter")
	if idx == -1 {
		t.Fatalf("expected --filter flag, got args: %v", args)
	}
	if idx+1 >= len(args) || args[idx+1] != "@tomorrow" {
		t.Errorf("expected --filter @tomorrow, got args: %v", args)
	}
}

func TestPickCommand_DateAndFilterBothPresent(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick([]string{"#followup"}, PickOpts{
		Date:   "@2026-02-23",
		Filter: "@some-note-filter",
	})

	args := cap.lastArgs()
	assertArgsContain(t, args, "@2026-02-23")

	idx := indexOf(args, "--filter")
	if idx == -1 {
		t.Fatalf("expected --filter flag, got args: %v", args)
	}
	if args[idx+1] != "@some-note-filter" {
		t.Errorf("expected --filter value @some-note-filter, got %q", args[idx+1])
	}

	// The positional @date must NOT appear after --filter
	filterValIdx := idx + 1
	dateIdx := indexOf(args, "@2026-02-23")
	if dateIdx > filterValIdx {
		t.Errorf("@date should appear before --filter, got args: %v", args)
	}
}

func TestPickCommand_TodoFlag(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick(nil, PickOpts{Todo: true, Date: "@2026-02-23"})

	args := cap.lastArgs()
	assertArgsContain(t, args, "--todo")
	assertArgsContain(t, args, "@2026-02-23")
}

func TestPickCommand_AllFlag(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick(nil, PickOpts{Todo: true, All: true})

	args := cap.lastArgs()
	assertArgsContain(t, args, "--todo")
	assertArgsContain(t, args, "--all")
}

func TestPickCommand_TodoAllDate(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick(nil, PickOpts{Date: "@2026-02-23", Todo: true, All: true})

	args := cap.lastArgs()
	assertArgsContain(t, args, "@2026-02-23")
	assertArgsContain(t, args, "--todo")
	assertArgsContain(t, args, "--all")
	assertArgsNotContain(t, args, "--filter")
}

func TestPickCommand_EmptyOptsNoExtraFlags(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick([]string{"#tag"}, PickOpts{})

	args := cap.lastArgs()
	if len(args) != 2 {
		t.Errorf("expected [pick #tag], got %v", args)
	}
}

func TestPickCommand_DateNotBehindFilterFlag(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick([]string{"#tag"}, PickOpts{Date: "@2026-02-23"})

	args := cap.lastArgs()
	argsStr := strings.Join(args, " ")
	if strings.Contains(argsStr, "--filter @2026-02-23") {
		t.Errorf("date should be positional, not behind --filter: %v", args)
	}
}

func TestPickCommand_ParentAndNotesStillWork(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick([]string{"#tag"}, PickOpts{
		Parent: "parent-uuid",
		Notes:  []string{"note1", "note2"},
	})

	args := cap.lastArgs()
	assertArgsContain(t, args, "--parent")
	assertArgsContain(t, args, "parent-uuid")
	assertArgsContain(t, args, "--notes")
	assertArgsContain(t, args, "note1,note2")
}

func TestPickCommand_DoneTagStillAppended(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick([]string{"#done"}, PickOpts{})

	args := cap.lastArgs()
	assertArgsContain(t, args, "#done")
	assertArgsContain(t, args, "--done")
}

func TestPickCommand_ArgOrdering(t *testing.T) {
	cap := &argCapture{}
	ruin := NewRuinCommandWithExecutor(cap, cap.VaultPath())

	_, _ = ruin.Pick.Pick([]string{"#followup"}, PickOpts{
		Date: "@2026-02-23",
		Todo: true,
		All:  true,
	})

	args := cap.lastArgs()
	// Expected order: pick, #followup, @2026-02-23, --todo, --all
	if args[0] != "pick" {
		t.Errorf("first arg should be 'pick', got %q", args[0])
	}
	tagIdx := indexOf(args, "#followup")
	dateIdx := indexOf(args, "@2026-02-23")
	todoIdx := indexOf(args, "--todo")
	allIdx := indexOf(args, "--all")

	if tagIdx == -1 || dateIdx == -1 || todoIdx == -1 || allIdx == -1 {
		t.Fatalf("missing expected args in %v", args)
	}

	// Tags come before date (both positional)
	if tagIdx > dateIdx {
		t.Errorf("tags should come before date, got args: %v", args)
	}
	// Positional args come before flags
	if dateIdx > todoIdx {
		t.Errorf("date should come before --todo, got args: %v", args)
	}
}

func assertArgsContain(t *testing.T, args []string, want string) {
	t.Helper()
	if indexOf(args, want) == -1 {
		t.Errorf("expected args to contain %q, got %v", want, args)
	}
}

func assertArgsNotContain(t *testing.T, args []string, unwanted string) {
	t.Helper()
	if indexOf(args, unwanted) != -1 {
		t.Errorf("expected args NOT to contain %q, got %v", unwanted, args)
	}
}

func indexOf(args []string, target string) int {
	for i, a := range args {
		if a == target {
			return i
		}
	}
	return -1
}
