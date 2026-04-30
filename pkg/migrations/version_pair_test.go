package migrations

import "testing"

func TestVersionPair_IsDev(t *testing.T) {
	cases := []struct {
		name string
		p    VersionPair
		want bool
	}{
		{"dev tag", VersionPair{Lazyruin: "dev", Ruin: "0.3.0"}, true},
		{"release", VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.0"}, false},
		{"empty lazyruin", VersionPair{Lazyruin: "", Ruin: "0.3.0"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.p.IsDev(); got != tc.want {
				t.Errorf("IsDev() = %v, want %v", got, tc.want)
			}
		})
	}
}
