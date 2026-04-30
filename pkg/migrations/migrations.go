package migrations

import (
	"slices"

	"github.com/donnellyk/lazyruin/pkg/commands"
)

// Migration describes one upgrade-time remediation. The Action receives
// the configured RuinCommand so it can issue subprocess calls (typically
// `ruin doctor` for a full vault scan).
type Migration struct {
	ID          string
	Description string
	Applies     func(curr, prev VersionPair) bool
	Action      func(c *commands.RuinCommand) error
}

// Registry holds the chronologically-ordered list of migrations the
// current lazyruin build knows about. Entries here are evaluated against
// the previous-vs-current version pair on every launch; ones that match
// and aren't already recorded as applied are surfaced to the user.
//
// New entries are appended as features ship that put existing vaults
// into a state requiring re-indexing.
var Registry = []Migration{}

// DoctorFullScan is the default Action for a migration that needs a full
// vault re-index. Mirrors `ruin doctor` with no path argument.
func DoctorFullScan(c *commands.RuinCommand) error {
	return c.DoctorFullScan()
}

// Detect computes the list of pending migrations for a given launch.
// Returns nil when no migrations apply: dev lazyruin builds, first
// launches (either half of prev unknown — defensive: a previous
// `ruin --version` failure would record an empty Ruin half, and we
// don't want a migration's Applies predicate seeing "" as input), or
// when versions haven't changed.
func Detect(curr, prev VersionPair, applied []string) []Migration {
	if curr.IsDev() {
		return nil
	}
	if prev.Lazyruin == "" || prev.Ruin == "" {
		return nil
	}
	if curr.Lazyruin == prev.Lazyruin && curr.Ruin == prev.Ruin {
		return nil
	}
	return pending(curr, prev, applied, Registry)
}

// pending is the registry-driven matcher, separated from Detect so
// tests can inject a custom registry.
func pending(curr, prev VersionPair, applied []string, registry []Migration) []Migration {
	var out []Migration
	for _, m := range registry {
		if slices.Contains(applied, m.ID) {
			continue
		}
		if m.Applies != nil && m.Applies(curr, prev) {
			out = append(out, m)
		}
	}
	return out
}
