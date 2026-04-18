package gui

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/donnellyk/lazyruin/pkg/testutil"
)

// TestOnboarding_EmptyVault_ShowsPrompt verifies the first-run confirmation
// fires when the vault is empty and the config flag is unset.
func TestOnboarding_EmptyVault_ShowsPrompt(t *testing.T) {
	mock := testutil.NewMockExecutor() // no notes
	tg := newTestGui(t, mock)
	defer tg.Close()

	// Test fixture default-suppresses onboarding; undo that and re-run layout.
	tg.gui.config.OnboardingOffered = false
	tg.gui.state.Initialized = false
	tg.gui.state.Dialog = nil
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatal(err)
	}

	if tg.gui.state.Dialog == nil {
		t.Fatal("expected onboarding confirm dialog on empty vault with unset flag")
	}
	if tg.gui.state.Dialog.Type != "confirm" {
		t.Errorf("Dialog.Type = %q, want confirm", tg.gui.state.Dialog.Type)
	}
	if !tg.gui.config.OnboardingOffered {
		t.Error("config.OnboardingOffered should flip to true once the prompt has been shown")
	}
}

// TestOnboarding_NonEmptyVault_NoPrompt verifies the prompt does NOT fire on
// vaults that already contain notes, regardless of the flag state.
func TestOnboarding_NonEmptyVault_NoPrompt(t *testing.T) {
	mock := testutil.NewMockExecutor().WithNotes(models.Note{UUID: "u1", Title: "existing"})
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.config.OnboardingOffered = false
	tg.gui.state.Initialized = false
	tg.gui.state.Dialog = nil
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatal(err)
	}

	if tg.gui.state.Dialog != nil {
		t.Errorf("unexpected dialog on non-empty vault: %+v", tg.gui.state.Dialog)
	}
	if tg.gui.config.OnboardingOffered {
		t.Error("OnboardingOffered should stay false on non-empty vault — prompt never fired")
	}
}

// TestOnboarding_AlreadyOffered_NoPrompt verifies the prompt does not re-fire
// on empty vaults when the flag is already set.
func TestOnboarding_AlreadyOffered_NoPrompt(t *testing.T) {
	mock := testutil.NewMockExecutor() // empty vault
	tg := newTestGui(t, mock)
	defer tg.Close()

	// Default fixture sets OnboardingOffered=true; confirm nothing appeared.
	if tg.gui.state.Dialog != nil {
		t.Errorf("unexpected dialog when OnboardingOffered=true: %+v", tg.gui.state.Dialog)
	}
}
