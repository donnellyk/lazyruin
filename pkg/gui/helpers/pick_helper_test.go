package helpers

import (
	"testing"
)

func TestParsePickQuery(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantTags   []string
		wantDate   string
		wantFilter string
	}{
		{
			name:     "single tag",
			raw:      "#followup",
			wantTags: []string{"#followup"},
		},
		{
			name:     "tag without hash gets prefixed",
			raw:      "followup",
			wantTags: []string{"#followup"},
		},
		{
			name:     "multiple tags",
			raw:      "#followup #urgent",
			wantTags: []string{"#followup", "#urgent"},
		},
		{
			name:     "date only",
			raw:      "@2026-02-23",
			wantDate: "@2026-02-23",
		},
		{
			name:     "date shorthand",
			raw:      "@today",
			wantDate: "@today",
		},
		{
			name:     "tag and date",
			raw:      "#followup @2026-02-23",
			wantTags: []string{"#followup"},
			wantDate: "@2026-02-23",
		},
		{
			name:     "date before tag",
			raw:      "@2026-02-23 #followup",
			wantTags: []string{"#followup"},
			wantDate: "@2026-02-23",
		},
		{
			name:     "multiple tags and date",
			raw:      "#followup @2026-02-23 #urgent",
			wantTags: []string{"#followup", "#urgent"},
			wantDate: "@2026-02-23",
		},
		{
			name:     "bare words become tags",
			raw:      "followup @tomorrow urgent",
			wantTags: []string{"#followup", "#urgent"},
			wantDate: "@tomorrow",
		},
		{
			name:       "empty input",
			raw:        "",
			wantTags:   nil,
			wantDate:   "",
			wantFilter: "",
		},
		{
			name:     "only whitespace",
			raw:      "   ",
			wantTags: nil,
			wantDate: "",
		},
		{
			name:     "multiple dates keeps last one",
			raw:      "@2026-01-01 @2026-02-23",
			wantDate: "@2026-02-23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, date, filter := ParsePickQuery(tt.raw)

			if !slicesEqual(tags, tt.wantTags) {
				t.Errorf("tags = %v, want %v", tags, tt.wantTags)
			}
			if date != tt.wantDate {
				t.Errorf("date = %q, want %q", date, tt.wantDate)
			}
			if filter != tt.wantFilter {
				t.Errorf("filter = %q, want %q", filter, tt.wantFilter)
			}
		})
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
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
