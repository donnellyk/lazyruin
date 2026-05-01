package migrations

import "github.com/donnellyk/lazyruin/pkg/commands"

func init() {
	Registry = append(Registry, Migration{
		ID:          "v0.4.0-tag-format",
		Description: "Upgrade to new tag format with ruin v0.4.0",
		Applies: func(curr, prev VersionPair) bool {
			// Fire when previous ruin was < 0.4.0. Covers the AncientVersion
			// bootstrap path (prev.Ruin == "0.0.0"). curr.Ruin doesn't need
			// a check because MinRuinVersion gates startup on >= 0.4.0.
			return prev.Ruin != "" && commands.VersionLess(prev.Ruin, "0.4.0")
		},
		Action: DoctorFullScan,
	})
}
