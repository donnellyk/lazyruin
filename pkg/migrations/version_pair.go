// Package migrations detects when an upgrade has put the user's vault
// into a state that requires `ruin doctor` to run, prompts under a
// blocking modal, runs the migration, and persists the outcome.
package migrations

// VersionPair captures the lazyruin / ruin-cli version pair seen on a
// given launch. Empty strings mean "unknown" — useful for representing
// the pre-state of a vault that's never been seen before.
type VersionPair struct {
	Lazyruin string
	Ruin     string
}

// LazyruinDevTag is the build-time `version` value for unreleased builds.
// Migration prompts are suppressed when the running lazyruin reports
// this tag so branch-hopping doesn't re-trigger the modal.
const LazyruinDevTag = "dev"

// IsDev reports whether the lazyruin half of the pair is a dev build.
func (v VersionPair) IsDev() bool {
	return v.Lazyruin == LazyruinDevTag
}
