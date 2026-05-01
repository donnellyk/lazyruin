package commands

import (
	"fmt"
	"testing"
)

func TestParseRuinVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"release standard", "ruin version 0.1.0\n", "0.1.0", false},
		{"release no newline", "ruin version 0.1.0", "0.1.0", false},
		{"release extra whitespace", "  ruin  version  0.2.3  \n", "0.2.3", false},
		{"release patch zero", "ruin version 1.0.0\n", "1.0.0", false},
		{"dev build with v prefix", "ruin version v0.1.0\n", "0.1.0", false},
		{"dev build with git describe suffix", "ruin version v0.1.0-1-gda17746\n", "0.1.0", false},
		{"dev build with dirty suffix", "ruin version v0.1.0-1-gda17746-dirty\n", "0.1.0", false},
		{"dev build without v prefix", "ruin version 0.1.0-1-gda17746\n", "0.1.0", false},
		{"version with dash but no v", "ruin version 1.2.3-rc.1\n", "1.2.3", false},
		{"empty", "", "", true},
		{"too few fields", "ruin version\n", "", true},
		{"wrong name", "goruin version 0.1.0\n", "", true},
		{"wrong middle word", "ruin v 0.1.0\n", "", true},
		{"bare semver", "0.1.0\n", "", true},
		{"v prefix alone", "ruin version v\n", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRuinVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseRuinVersion(%q) = %q, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("parseRuinVersion(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseRuinVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestVersionLess(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		// strictly less
		{"0.0.9", "0.1.0", true},
		{"0.1.0", "0.2.0", true},
		{"0.1.0", "1.0.0", true},
		{"0.1.9", "0.1.10", true},
		{"1.2.3", "1.2.4", true},
		{"0.0.0", "0.0.1", true},

		// equal
		{"0.1.0", "0.1.0", false},
		{"1.2.3", "1.2.3", false},

		// strictly greater
		{"0.1.1", "0.1.0", false},
		{"0.2.0", "0.1.99", false},
		{"1.0.0", "0.99.99", false},

		// pre-release suffixes stripped to base version
		{"0.1.0-rc.1", "0.1.0", false}, // treated as equal (both 0.1.0)
		{"0.0.9-rc.1", "0.1.0", true},

		// malformed inputs treated as zeros
		{"", "0.0.1", true},
		{"abc", "0.0.1", true},
		{"0.0.1", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.a+" < "+tt.b, func(t *testing.T) {
			if got := VersionLess(tt.a, tt.b); got != tt.want {
				t.Errorf("VersionLess(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSplitSemver(t *testing.T) {
	tests := []struct {
		in   string
		want [3]int
	}{
		{"0.1.0", [3]int{0, 1, 0}},
		{"1.2.3", [3]int{1, 2, 3}},
		{"0.0.0", [3]int{0, 0, 0}},
		{"10.20.30", [3]int{10, 20, 30}},
		{"0.1.0-rc.1", [3]int{0, 1, 0}},
		{"0.1.0+build.1", [3]int{0, 1, 0}},
		{"0.1", [3]int{0, 1, 0}},
		{"0", [3]int{0, 0, 0}},
		{"", [3]int{0, 0, 0}},
		{"not.a.number", [3]int{0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := splitSemver(tt.in)
			if got != tt.want {
				t.Errorf("splitSemver(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestMinRuinVersionConstant(t *testing.T) {
	// Guard against accidental whitespace or malformed constant.
	if _, err := parseRuinVersion("ruin version " + MinRuinVersion); err != nil {
		t.Errorf("MinRuinVersion %q is not parseable by parseRuinVersion: %v", MinRuinVersion, err)
	}
	v := splitSemver(MinRuinVersion)
	if v[0] == 0 && v[1] == 0 && v[2] == 0 {
		t.Errorf("MinRuinVersion %q parses to 0.0.0 — likely malformed", MinRuinVersion)
	}
}

func TestVersion_WithMockExecutor(t *testing.T) {
	tests := []struct {
		name       string
		mockOutput string
		want       string
		wantErr    bool
	}{
		{"release format", "ruin version 0.1.0\n", "0.1.0", false},
		{"dev build format", "ruin version v0.1.0-1-gda17746\n", "0.1.0", false},
		{"dev build with dirty suffix", "ruin version v0.1.0-1-gda17746-dirty\n", "0.1.0", false},
		{"future release", "ruin version 1.5.2\n", "1.5.2", false},
		{"malformed output", "garbage\n", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockExecutor().WithVersion(tt.mockOutput)
			r := NewRuinCommandWithExecutor(mock, "/mock/vault")
			got, err := r.Version()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Version() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Errorf("Version() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Version() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersion_ExecutorError(t *testing.T) {
	mock := NewMockExecutor().WithError(fmt.Errorf("binary not found"))
	r := NewRuinCommandWithExecutor(mock, "/mock/vault")
	if _, err := r.Version(); err == nil {
		t.Errorf("Version() expected error, got nil")
	}
}

func TestCheckVersion_WithMockExecutor(t *testing.T) {
	tests := []struct {
		name       string
		mockOutput string
		wantOK     bool
		wantGot    string
		wantErr    bool
	}{
		{
			name:       "equal to minimum",
			mockOutput: "ruin version " + MinRuinVersion + "\n",
			wantOK:     true,
			wantGot:    MinRuinVersion,
		},
		{
			name:       "above minimum",
			mockOutput: "ruin version 1.2.3\n",
			wantOK:     true,
			wantGot:    "1.2.3",
		},
		{
			name:       "dev build above minimum",
			mockOutput: "ruin version v1.0.0-2-gabc1234\n",
			wantOK:     true,
			wantGot:    "1.0.0",
		},
		{
			name:       "dev build at minimum",
			mockOutput: "ruin version v" + MinRuinVersion + "-1-gda17746\n",
			wantOK:     true,
			wantGot:    MinRuinVersion,
		},
		{
			name:       "below minimum",
			mockOutput: "ruin version 0.0.9\n",
			wantOK:     false,
			wantGot:    "0.0.9",
		},
		{
			name:       "malformed output",
			mockOutput: "not a version\n",
			wantOK:     false,
			wantGot:    "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockExecutor().WithVersion(tt.mockOutput)
			r := NewRuinCommandWithExecutor(mock, "/mock/vault")
			ok, got, err := r.CheckVersion()
			if ok != tt.wantOK {
				t.Errorf("CheckVersion() ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.wantGot {
				t.Errorf("CheckVersion() got = %q, want %q", got, tt.wantGot)
			}
			if tt.wantErr && err == nil {
				t.Errorf("CheckVersion() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("CheckVersion() unexpected error: %v", err)
			}
		})
	}
}

func TestCheckVersion_ExecutorError(t *testing.T) {
	mock := NewMockExecutor().WithError(fmt.Errorf("binary not found"))
	r := NewRuinCommandWithExecutor(mock, "/mock/vault")
	ok, got, err := r.CheckVersion()
	if ok {
		t.Errorf("CheckVersion() ok = true on executor error, want false")
	}
	if got != "" {
		t.Errorf("CheckVersion() got = %q, want empty on error", got)
	}
	if err == nil {
		t.Errorf("CheckVersion() expected error, got nil")
	}
}
