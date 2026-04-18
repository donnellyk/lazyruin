// Package onboarding wraps the first-run walkthrough note: detecting an
// empty vault, installing the note, and cleaning it up later.
package onboarding

import (
	_ "embed"
	"fmt"

	"github.com/donnellyk/lazyruin/pkg/commands"
)

// Tag is the sentinel tag embedded in the onboarding note. It lets the
// cleanup command find the note (or notes, if the user re-added it) without
// depending on title or UUID state.
const Tag = "lazyruin-onboarding"

//go:embed onboarding.md
var content string

// Content returns the full onboarding markdown. The content embeds the
// sentinel #lazyruin-onboarding tag so the note is tagged on creation.
func Content() string { return content }

// IsVaultEmpty reports whether the ruin vault has no notes. A tiny search
// is the cheapest reliable way to ask ruin "any notes at all?".
func IsVaultEmpty(ruinCmd *commands.RuinCommand) (bool, error) {
	notes, err := ruinCmd.Search.Search("", commands.SearchOptions{Limit: 1, Everything: true})
	if err != nil {
		return false, err
	}
	return len(notes) == 0, nil
}

// CreateNote writes the onboarding walkthrough into the vault via `ruin log`.
// The sentinel tag is already in the body, so the note is discoverable by
// `ruin pick #lazyruin-onboarding`.
func CreateNote(ruinCmd *commands.RuinCommand) error {
	if _, err := ruinCmd.Execute("log", content); err != nil {
		return fmt.Errorf("create onboarding note: %w", err)
	}
	return nil
}

// Cleanup deletes every note tagged #lazyruin-onboarding and then removes
// the tag from ruin's index. Returns the number of notes deleted.
//
// Cleanup intentionally deletes by tag rather than by title or UUID: that
// catches re-added walkthroughs and any stray notes a user accidentally
// marked with the sentinel tag, which is what "remove the document and any
// tags it created" implies.
//
// The tag is stored as a *global* tag (appears once at the top of the body
// and is promoted to frontmatter by ruin), so `ruin search #tag` is the
// right lookup path — `ruin pick` only matches inline line-level tags.
func Cleanup(ruinCmd *commands.RuinCommand) (int, error) {
	notes, err := ruinCmd.Search.Search("#"+Tag, commands.SearchOptions{Everything: true})
	if err != nil {
		return 0, fmt.Errorf("find onboarding notes: %w", err)
	}

	seen := make(map[string]bool, len(notes))
	deleted := 0
	for _, n := range notes {
		if n.UUID == "" || seen[n.UUID] {
			continue
		}
		seen[n.UUID] = true
		if err := ruinCmd.Note.Delete(n.UUID); err != nil {
			return deleted, fmt.Errorf("delete onboarding note %s: %w", n.UUID, err)
		}
		deleted++
	}

	// Remove the tag from the index. If ruin has already auto-pruned it
	// after deleting the last referencing note, this returns an error that
	// is safe to ignore — the end state is the same.
	_ = ruinCmd.Tags.Delete(Tag)
	return deleted, nil
}
