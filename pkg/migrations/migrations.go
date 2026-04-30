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

// AncientVersion is the sentinel version pair used as `prev` when the
// vault has content but no state.json entry — i.e., the user is
// opening lazyruin for the first time after the migration system
// shipped, against an existing vault. Setting prev to "0.0.0/0.0.0"
// makes every currently-applicable registry entry's Applies
// predicate fire so the user gets the same one-time re-index path as
// any other upgrade.
var AncientVersion = VersionPair{Lazyruin: "0.0.0", Ruin: "0.0.0"}

// BootstrapPrev picks the right `prev` value to feed into Detect when
// no state.json entry exists for the current vault. Empty vaults
// return zero-value VersionPair{} so Detect short-circuits on
// "first install — nothing to migrate." Non-empty vaults return
// AncientVersion so existing users upgrading into the migration
// system run their pending migrations.
func BootstrapPrev(vaultIsEmpty bool) VersionPair {
	if vaultIsEmpty {
		return VersionPair{}
	}
	return AncientVersion
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
