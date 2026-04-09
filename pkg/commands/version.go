package commands

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// MinRuinVersion is the minimum ruin CLI version lazyruin expects. Bump this
// when lazyruin starts using a ruin feature that requires a newer version.
// A runtime check compares the installed ruin binary's --version output
// against this value and surfaces a warning in the status bar if below.
const MinRuinVersion = "0.1.0"

// Version runs `ruin --version` and returns the parsed version string.
// Expected output format: "ruin version X.Y.Z" (release) or
// "ruin version vX.Y.Z-N-gHASH" (dev build). See parseRuinVersion.
//
// Routes through the injected Executor when set (for tests) so mocks can
// return canned version output. The injected executor path is the only
// case where --version reaches the executor; otherwise the real binary is
// invoked directly without --json or --vault.
func (r *RuinCommand) Version() (string, error) {
	var out []byte
	var err error
	if r.executor != nil {
		out, err = r.executor.Execute("--version")
	} else {
		if r.bin == "" {
			return "", fmt.Errorf("ruin binary path not set")
		}
		cmd := exec.Command(r.bin, "--version")
		out, err = cmd.Output()
	}
	if err != nil {
		return "", fmt.Errorf("failed to run ruin --version: %w", err)
	}
	return parseRuinVersion(string(out))
}

// CheckVersion compares the installed ruin version against MinRuinVersion.
// Returns ok=true only when the installed version could be determined and
// is >= MinRuinVersion. ok=false covers both "below minimum" and "could not
// determine"; callers should differentiate via the err return:
//
//	ok=true,  err=nil — version ok, no warning
//	ok=false, err=nil — version below minimum, got contains installed version
//	ok=false, err!=nil — version could not be determined
//
// The caller is responsible for deciding what to do; this method never
// blocks startup.
func (r *RuinCommand) CheckVersion() (ok bool, got string, err error) {
	v, verr := r.Version()
	if verr != nil {
		return false, "", verr
	}
	if versionLess(v, MinRuinVersion) {
		return false, v, nil
	}
	return true, v, nil
}

// parseRuinVersion extracts a semver string from ruin's --version output.
//
// Accepted forms:
//   - release build:  "ruin version 0.1.0"
//   - dev/debug build: "ruin version v0.1.0-1-gda17746" (git describe --tags)
//
// The leading "v" and the git-describe suffix (anything from the first "-") are
// stripped, yielding a bare "MAJOR.MINOR.PATCH" string that splitSemver can read.
func parseRuinVersion(output string) (string, error) {
	fields := strings.Fields(output)
	if len(fields) < 3 || fields[0] != "ruin" || fields[1] != "version" {
		return "", fmt.Errorf("unexpected ruin --version output: %q", strings.TrimSpace(output))
	}
	raw := fields[2]
	raw = strings.TrimPrefix(raw, "v")
	if dash := strings.Index(raw, "-"); dash >= 0 {
		raw = raw[:dash]
	}
	if raw == "" {
		return "", fmt.Errorf("unexpected ruin --version output: %q", strings.TrimSpace(output))
	}
	return raw, nil
}

// versionLess reports whether semver a is strictly less than b.
// Both inputs are expected as "MAJOR.MINOR.PATCH". Non-numeric fields or
// malformed input are treated as zero (fail-open).
func versionLess(a, b string) bool {
	av := splitSemver(a)
	bv := splitSemver(b)
	for i := range 3 {
		if av[i] < bv[i] {
			return true
		}
		if av[i] > bv[i] {
			return false
		}
	}
	return false
}

// splitSemver returns [major, minor, patch] as ints. Missing or malformed
// fields become 0.
func splitSemver(v string) [3]int {
	var out [3]int
	parts := strings.SplitN(v, ".", 3)
	for i := range min(3, len(parts)) {
		// Strip any pre-release suffix (e.g., "0-rc.1" → "0")
		p := parts[i]
		if dash := strings.IndexAny(p, "-+"); dash >= 0 {
			p = p[:dash]
		}
		n, _ := strconv.Atoi(p)
		out[i] = n
	}
	return out
}
