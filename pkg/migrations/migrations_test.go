package migrations

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/commands"
)

// fixture builds a registry of three migrations gated on Ruin version,
// for predictable test output.
func fixture() []Migration {
	noop := func(*commands.RuinCommand) error { return nil }
	return []Migration{
		{
			ID:          "ruin-0.3.0",
			Description: "0.3.0 cutover",
			Applies: func(curr, prev VersionPair) bool {
				return prev.Ruin == "0.2.0" && curr.Ruin == "0.3.0"
			},
			Action: noop,
		},
		{
			ID:          "ruin-0.3.1",
			Description: "0.3.1 cutover",
			Applies: func(curr, prev VersionPair) bool {
				return prev.Ruin == "0.3.0" && curr.Ruin == "0.3.1"
			},
			Action: noop,
		},
		{
			ID:          "always",
			Description: "fires every upgrade",
			Applies:     func(curr, prev VersionPair) bool { return true },
			Action:      noop,
		},
	}
}

func TestPending_FiltersByApplies(t *testing.T) {
	prev := VersionPair{Lazyruin: "0.1.0", Ruin: "0.3.0"}
	curr := VersionPair{Lazyruin: "0.1.0", Ruin: "0.3.1"}
	got := pending(curr, prev, nil, fixture())
	ids := idsOf(got)
	want := []string{"ruin-0.3.1", "always"}
	if !equal(ids, want) {
		t.Errorf("ids = %v, want %v", ids, want)
	}
}

func TestPending_SkipsAppliedIDs(t *testing.T) {
	prev := VersionPair{Lazyruin: "0.1.0", Ruin: "0.3.0"}
	curr := VersionPair{Lazyruin: "0.1.0", Ruin: "0.3.1"}
	got := pending(curr, prev, []string{"ruin-0.3.1"}, fixture())
	ids := idsOf(got)
	want := []string{"always"}
	if !equal(ids, want) {
		t.Errorf("ids = %v, want %v", ids, want)
	}
}

func TestDetect_FirstLaunchSkips(t *testing.T) {
	prev := VersionPair{} // empty — never seen before.
	curr := VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.1"}
	old := Registry
	defer func() { Registry = old }()
	Registry = fixture()
	if got := Detect(curr, prev, nil); len(got) != 0 {
		t.Errorf("expected empty on first launch, got %d entries", len(got))
	}
}

func TestDetect_DevBuildSuppresses(t *testing.T) {
	prev := VersionPair{Lazyruin: "0.1.0", Ruin: "0.3.0"}
	curr := VersionPair{Lazyruin: "dev", Ruin: "0.3.1"}
	old := Registry
	defer func() { Registry = old }()
	Registry = fixture()
	if got := Detect(curr, prev, nil); len(got) != 0 {
		t.Errorf("expected empty on dev build, got %d entries", len(got))
	}
}

func TestBootstrapPrev_EmptyVault(t *testing.T) {
	got := BootstrapPrev(true)
	if got.Lazyruin != "" || got.Ruin != "" {
		t.Errorf("empty vault should bootstrap to zero pair, got %+v", got)
	}
}

func TestBootstrapPrev_NonEmptyVault(t *testing.T) {
	got := BootstrapPrev(false)
	if got != AncientVersion {
		t.Errorf("non-empty vault should bootstrap to AncientVersion, got %+v", got)
	}
}

func TestDetect_AncientPrevMatchesAllRegistryEntries(t *testing.T) {
	// Existing user upgrading into the migration system: prev =
	// AncientVersion, curr = real current version. Every registry
	// entry whose Applies fires for this transition should be
	// returned.
	curr := VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.1"}
	old := Registry
	defer func() { Registry = old }()
	Registry = fixture()
	got := Detect(curr, AncientVersion, nil)
	ids := idsOf(got)
	if !equal(ids, []string{"always"}) {
		t.Errorf("ids = %v, want [always] (only the unconditional fixture entry matches a 0.0.0 prev)", ids)
	}
}

func TestDetect_NoChangeReturnsEmpty(t *testing.T) {
	v := VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.1"}
	old := Registry
	defer func() { Registry = old }()
	Registry = fixture()
	if got := Detect(v, v, nil); len(got) != 0 {
		t.Errorf("expected empty when versions match, got %v", idsOf(got))
	}
}

func TestDetect_PrevLazyruinOnlyChange(t *testing.T) {
	prev := VersionPair{Lazyruin: "0.1.0", Ruin: "0.3.1"}
	curr := VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.1"}
	old := Registry
	defer func() { Registry = old }()
	Registry = fixture()
	got := Detect(curr, prev, nil)
	ids := idsOf(got)
	if !equal(ids, []string{"always"}) {
		t.Errorf("ids = %v, want [always]", ids)
	}
}

func idsOf(ms []Migration) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		out = append(out, m.ID)
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
