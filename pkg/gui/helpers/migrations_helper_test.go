package helpers

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/migrations"
	"github.com/donnellyk/lazyruin/pkg/testutil"
)

// stubPrompter records dialog transitions and lets tests step through
// the helper's state machine deterministically. Update() runs its
// callback synchronously.
type stubPrompter struct {
	mu          sync.Mutex
	promptShown bool
	runShown    bool
	errorShown  error
	closed      bool
	errors      []error
	onRun       func() error
}

func (p *stubPrompter) ShowMigrationPrompt(_, _ migrations.VersionPair, _ string, _ []migrations.Migration, onRun func() error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.promptShown = true
	p.onRun = onRun
}

func (p *stubPrompter) ShowMigrationRunning(_ string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.runShown = true
}

func (p *stubPrompter) ShowMigrationError(err error, _ func() error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.errorShown = err
}

func (p *stubPrompter) CloseDialog() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
}

func (p *stubPrompter) Update(fn func() error) { _ = fn() }

func (p *stubPrompter) ShowError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.errors = append(p.errors, err)
}

func newHelperFixture(t *testing.T, exec testutil.MockExecutor, pending []migrations.Migration) (*MigrationsHelper, *stubPrompter, *migrations.Store) {
	t.Helper()
	storePath := filepath.Join(t.TempDir(), "state.json")
	store := migrations.NewStoreWithPath(storePath)
	if err := store.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	prev := migrations.VersionPair{Lazyruin: "0.1.0", Ruin: "0.3.0"}
	curr := migrations.VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.1"}
	prompter := &stubPrompter{}
	ruin := commands.NewRuinCommandWithExecutor(&exec, "/v")
	h := NewMigrationsHelper(ruin, prompter, store, "/v", prev, curr, pending)
	return h, prompter, store
}

func TestMigrationsHelper_StartFalseWhenEmpty(t *testing.T) {
	h, prompter, _ := newHelperFixture(t, *testutil.NewMockExecutor(), nil)
	if h.Start() {
		t.Error("Start returned true with no pending migrations")
	}
	if prompter.promptShown {
		t.Error("prompt was shown despite empty pending list")
	}
}

func TestMigrationsHelper_HappyPath(t *testing.T) {
	called := 0
	mig := migrations.Migration{
		ID: "happy",
		Action: func(*commands.RuinCommand) error {
			called++
			return nil
		},
	}
	h, prompter, store := newHelperFixture(t, *testutil.NewMockExecutor(), []migrations.Migration{mig})

	if !h.Start() {
		t.Fatal("expected Start to return true")
	}
	// Simulate user pressing y on the prompt.
	if err := prompter.onRun(); err != nil {
		t.Fatalf("onRun: %v", err)
	}

	// execute runs in a goroutine — wait for it.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		prompter.mu.Lock()
		closed := prompter.closed
		prompter.mu.Unlock()
		if closed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	prompter.mu.Lock()
	defer prompter.mu.Unlock()
	if !prompter.runShown {
		t.Error("running modal not shown")
	}
	if !prompter.closed {
		t.Error("dialog not closed after success")
	}
	if prompter.errorShown != nil {
		t.Errorf("unexpected error dialog: %v", prompter.errorShown)
	}
	if called != 1 {
		t.Errorf("Action called %d times, want 1", called)
	}
	entry, ok := store.VaultEntry("/v")
	if !ok {
		t.Fatal("vault entry missing after success")
	}
	if len(entry.AppliedMigrations) != 1 || entry.AppliedMigrations[0] != "happy" {
		t.Errorf("AppliedMigrations = %v, want [happy]", entry.AppliedMigrations)
	}
	if entry.LastRuinVersion != "0.3.1" || entry.LastLazyruinVersion != "0.2.0" {
		t.Errorf("versions not recorded: %+v", entry)
	}
}

func TestMigrationsHelper_FailureSurfacesError(t *testing.T) {
	wantErr := errors.New("doctor exploded")
	mig := migrations.Migration{
		ID:     "boom",
		Action: func(*commands.RuinCommand) error { return wantErr },
	}
	h, prompter, store := newHelperFixture(t, *testutil.NewMockExecutor(), []migrations.Migration{mig})
	h.Start()
	_ = prompter.onRun()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		prompter.mu.Lock()
		errSet := prompter.errorShown != nil
		prompter.mu.Unlock()
		if errSet {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	prompter.mu.Lock()
	defer prompter.mu.Unlock()
	if !errors.Is(prompter.errorShown, wantErr) {
		t.Errorf("error dialog = %v, want %v", prompter.errorShown, wantErr)
	}
	if prompter.closed {
		t.Error("dialog should not have been closed on failure")
	}
	if entry, ok := store.VaultEntry("/v"); ok {
		t.Errorf("state should not record entry on failure, got %+v", entry)
	}
}

func TestMigrationsHelper_RecordNoPending(t *testing.T) {
	h, _, store := newHelperFixture(t, *testutil.NewMockExecutor(), nil)
	h.RecordNoPending()
	entry, ok := store.VaultEntry("/v")
	if !ok {
		t.Fatal("expected entry after RecordNoPending")
	}
	if entry.LastRuinVersion != "0.3.1" {
		t.Errorf("ruin = %q, want 0.3.1", entry.LastRuinVersion)
	}
}
