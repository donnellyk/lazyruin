package helpers

import (
	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/migrations"
)

// MigrationsHelper coordinates the upgrade-migration UX: it owns the
// pending list and the on-disk state store, and it knows how to swap
// between the prompt / running / error dialog states from the goroutine
// that drives `ruin doctor`.
type MigrationsHelper struct {
	ruinCmd  *commands.RuinCommand
	store    *migrations.Store
	pending  []migrations.Migration
	curr     migrations.VersionPair
	prev     migrations.VersionPair
	vault    string
	prompter Prompter
}

// Prompter is the GUI surface the migrations helper drives. The gui
// package satisfies this on *Gui so the helper avoids importing gui.
type Prompter interface {
	ShowMigrationPrompt(prev, curr migrations.VersionPair, vaultPath string, pending []migrations.Migration, onRun func() error)
	ShowMigrationRunning(vaultPath string)
	ShowMigrationError(err error, onRetry func() error)
	CloseDialog()
	Update(func() error)
	ShowError(error)
}

// NewMigrationsHelper creates a helper bound to the given pending list,
// store, and version pair. Construction is cheap; the helper does no
// disk or subprocess work until Start is called.
func NewMigrationsHelper(ruinCmd *commands.RuinCommand, prompter Prompter, store *migrations.Store, vault string, prev, curr migrations.VersionPair, pending []migrations.Migration) *MigrationsHelper {
	return &MigrationsHelper{
		ruinCmd:  ruinCmd,
		store:    store,
		pending:  pending,
		curr:     curr,
		prev:     prev,
		vault:    vault,
		prompter: prompter,
	}
}

// Pending reports the migrations the helper will run.
func (h *MigrationsHelper) Pending() []migrations.Migration { return h.pending }

// Start surfaces the prompt dialog. Returns true when a dialog was
// shown (caller should suppress other startup prompts), false when
// there's nothing to migrate.
func (h *MigrationsHelper) Start() bool {
	if len(h.pending) == 0 {
		return false
	}
	h.prompter.ShowMigrationPrompt(h.prev, h.curr, h.vault, h.pending, h.runAll)
	return true
}

// runAll is invoked by the prompt's OnConfirm. It swaps to the running
// modal and dispatches the doctor calls on a goroutine.
func (h *MigrationsHelper) runAll() error {
	h.prompter.ShowMigrationRunning(h.vault)
	go h.execute()
	return nil
}

// execute runs every pending migration's Action sequentially. Doctor
// failures stop the loop; the helper switches to the error dialog and
// the user can retry. State is recorded on success only.
func (h *MigrationsHelper) execute() {
	var ranIDs []string
	var firstErr error
	for _, m := range h.pending {
		if err := m.Action(h.ruinCmd); err != nil {
			firstErr = err
			break
		}
		ranIDs = append(ranIDs, m.ID)
	}

	h.prompter.Update(func() error {
		if firstErr != nil {
			h.prompter.ShowMigrationError(firstErr, h.runAll)
			return nil
		}
		h.store.RecordApplied(h.vault, h.curr, ranIDs)
		if err := h.store.Save(); err != nil {
			// Surface via the error dialog so the failure isn't swallowed
			// by an immediate CloseDialog; user can quit and retry.
			h.prompter.ShowMigrationError(err, h.runAll)
			return nil
		}
		h.prompter.CloseDialog()
		return nil
	})
}

// RecordNoPending writes the current versions for the vault when no
// migrations were pending on this launch. Updates the on-disk state so
// a future migration's Applies predicate sees an accurate "previous"
// version.
func (h *MigrationsHelper) RecordNoPending() {
	h.store.RecordVersions(h.vault, h.curr)
	if err := h.store.Save(); err != nil {
		h.prompter.ShowError(err)
	}
}
