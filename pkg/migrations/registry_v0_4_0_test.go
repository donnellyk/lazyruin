package migrations

import "testing"

const v04ID = "v0.4.0-tag-format"

func findV04Migration(t *testing.T) Migration {
	t.Helper()
	for _, m := range Registry {
		if m.ID == v04ID {
			return m
		}
	}
	t.Fatalf("migration %q not found in Registry", v04ID)
	return Migration{}
}

func TestV04TagFormat_Registered(t *testing.T) {
	m := findV04Migration(t)
	if m.Applies == nil {
		t.Errorf("migration %q has nil Applies", v04ID)
	}
	if m.Action == nil {
		t.Errorf("migration %q has nil Action", v04ID)
	}
	if m.Description == "" {
		t.Errorf("migration %q has empty Description", v04ID)
	}
}

func TestV04TagFormat_Applies(t *testing.T) {
	m := findV04Migration(t)
	curr := VersionPair{Lazyruin: "1.0.0", Ruin: "0.4.0"}

	tests := []struct {
		name string
		prev VersionPair
		want bool
	}{
		{"prev ruin 0.3.5 → fires", VersionPair{Lazyruin: "0.9.0", Ruin: "0.3.5"}, true},
		{"prev ruin 0.4.0 → skip (already migrated)", VersionPair{Lazyruin: "0.9.0", Ruin: "0.4.0"}, false},
		{"prev ruin 0.4.1 → skip", VersionPair{Lazyruin: "0.9.0", Ruin: "0.4.1"}, false},
		{"prev ruin 0.0.0 (AncientVersion bootstrap) → fires", AncientVersion, true},
		{"prev ruin empty → skip (defensive; Detect already short-circuits)", VersionPair{Lazyruin: "0.9.0", Ruin: ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Applies(curr, tt.prev); got != tt.want {
				t.Errorf("Applies(curr=%+v, prev=%+v) = %v, want %v", curr, tt.prev, got, tt.want)
			}
		})
	}
}
