//go:build smoketest

package migrations

func init() {
	// Smoke-test build tag: registers a synthetic migration that always
	// applies, used by scripts/smoke-test.sh to exercise the prompt →
	// run → state cycle end-to-end without a real shipped migration.
	// Production builds omit this file (no -tags=smoketest), so the env
	// var no longer matters.
	Registry = append(Registry, Migration{
		ID:          "test-migration",
		Description: "Test migration (smoketest build)",
		Applies:     func(curr, prev VersionPair) bool { return true },
		Action:      DoctorFullScan,
	})
}
