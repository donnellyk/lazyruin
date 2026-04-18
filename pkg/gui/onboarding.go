package gui

import (
	"github.com/donnellyk/lazyruin/pkg/gui/onboarding"
)

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

	gui.ShowConfirm(
		"Welcome to LazyRuin",
		"Your vault looks empty. Add a walkthrough note to get started?",
		func() error {
			if err := onboarding.CreateNote(gui.ruinCmd); err != nil {
				gui.ShowError(err)
				return nil
			}
			gui.helpers.Refresh().RefreshAll()
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
	return nil
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
