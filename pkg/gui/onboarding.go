package gui

import (
	"fmt"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/gui/onboarding"
	"github.com/jesseduffield/gocui"
)

// maybeOfferInit shows an init prompt when the configured vault path has
// not yet been initialized as a ruin vault. Returns true when a prompt was
// shown (callers should skip the onboarding prompt in that case; it will
// be chained from the init flow on success).
func (gui *Gui) maybeOfferInit() bool {
	if !gui.state.NeedsInit {
		return false
	}

	vaultPath := gui.ruinCmd.VaultPath()
	gui.state.Dialog = &DialogState{
		Active:       true,
		Type:         "confirm",
		Title:        "Initialize vault",
		Message:      fmt.Sprintf("No vault configuration found\n\nInitialize it as %s?", vaultPath),
		Hero:         true,
		Footer:       " [y] Yes · [q/Esc] Quit ",
		QuitOnCancel: true,
		OnConfirm: func() error {
			if err := gui.ruinCmd.Init(); err != nil {
				gui.state.ExitError = err
				return gocui.ErrQuit
			}
			gui.state.NeedsInit = false
			gui.helpers.Refresh().RefreshAll()
			// confirmYes calls closeDialog right after OnConfirm, which
			// would wipe any dialog we set here. Defer the onboarding
			// prompt to the next tick so it lands after closeDialog.
			gui.g.Update(func(*gocui.Gui) error {
				gui.maybeOfferOnboarding()
				return nil
			})
			return nil
		},
		OnCancel: func() {
			gui.state.ExitError = fmt.Errorf("vault not initialized at %s", vaultPath)
		},
	}
	return true
}

// maybeOfferOnboarding shows a first-run confirmation dialog if the vault
// is empty and the user has not been prompted before. The "offered" flag is
// flipped true on either outcome (accept or decline) so this never fires a
// second time, even across multiple launches against empty vaults.
func (gui *Gui) maybeOfferOnboarding() {
	if gui.config == nil || gui.config.OnboardingOffered {
		return
	}
	empty, err := onboarding.IsVaultEmpty(gui.ruinCmd)
	if err != nil || !empty {
		return
	}

	gui.ShowHeroConfirm(
		"Welcome to LazyRuin",
		"Your vault looks empty. Add a walkthrough note?",
		func() error {
			if err := onboarding.CreateNote(gui.ruinCmd); err != nil {
				gui.ShowError(err)
				return nil
			}
			gui.helpers.Refresh().RefreshAll()
			gui.openOnboardingNote()
			return nil
		},
	)
	// ShowConfirm only fires OnConfirm on accept. We flip the flag
	// unconditionally up front so declining also records that we have
	// offered the walkthrough — the prompt never fires twice.
	gui.markOnboardingOffered()
}

func (gui *Gui) markOnboardingOffered() {
	if gui.config == nil || gui.config.OnboardingOffered {
		return
	}
	gui.config.OnboardingOffered = true
	if err := gui.config.Save(); err != nil {
		gui.ShowError(err)
	}
}

// InstallOnboarding creates the walkthrough note via ruin log, then
// refreshes the Notes panel. Used by the "Onboarding: add walkthrough"
// palette command.
func (gui *Gui) InstallOnboarding() error {
	if err := onboarding.CreateNote(gui.ruinCmd); err != nil {
		return err
	}
	gui.helpers.Refresh().RefreshAll()
	gui.openOnboardingNote()
	return nil
}

// openOnboardingNote looks up the freshly created walkthrough note by its
// sentinel tag and opens it in the preview pane. Silent no-op on lookup
// failure — the note is still present in the notes list.
func (gui *Gui) openOnboardingNote() {
	notes, err := gui.ruinCmd.Search.Search("#"+onboarding.Tag, commands.SearchOptions{
		Limit:           1,
		IncludeContent:  true,
		StripGlobalTags: true,
		StripTitle:      true,
		Everything:      true,
	})
	if err != nil || len(notes) == 0 {
		return
	}
	gui.openNote(&notes[0])
}

// CleanupOnboarding deletes every note tagged #lazyruin-onboarding and
// removes the tag from the index. Used by the "Onboarding: cleanup"
// palette command.
func (gui *Gui) CleanupOnboarding() error {
	if _, err := onboarding.Cleanup(gui.ruinCmd); err != nil {
		return err
	}
	gui.helpers.Refresh().RefreshAll()
	return nil
}
