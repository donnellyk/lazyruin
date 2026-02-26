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
		wantFlags  PickFlags
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
		{
			name:      "--any flag",
			raw:       "#followup --any",
			wantTags:  []string{"#followup"},
			wantFlags: PickFlags{Any: true},
		},
		{
			name:      "--todo flag",
			raw:       "#followup --todo",
			wantTags:  []string{"#followup"},
			wantFlags: PickFlags{Todo: true},
		},
		{
			name:      "both flags",
			raw:       "#followup --any --todo @today",
			wantTags:  []string{"#followup"},
			wantDate:  "@today",
			wantFlags: PickFlags{Any: true, Todo: true},
		},
		{
			name:      "flags among tags",
			raw:       "--todo #urgent --any",
			wantTags:  []string{"#urgent"},
			wantFlags: PickFlags{Any: true, Todo: true},
		},
		{
			name:      "--all-tags flag alone",
			raw:       "--all-tags",
			wantFlags: PickFlags{AllTags: true},
		},
		{
			name:      "--all-tags with other flags",
			raw:       "#followup --all-tags --todo",
			wantTags:  []string{"#followup"},
			wantFlags: PickFlags{Todo: true, AllTags: true},
		},
		{
			name:      "--all-tags with tags and date",
			raw:       "--all-tags #urgent @today",
			wantTags:  []string{"#urgent"},
			wantDate:  "@today",
			wantFlags: PickFlags{AllTags: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, date, filter, flags := ParsePickQuery(tt.raw)

			if !slicesEqual(tags, tt.wantTags) {
				t.Errorf("tags = %v, want %v", tags, tt.wantTags)
			}
			if date != tt.wantDate {
				t.Errorf("date = %q, want %q", date, tt.wantDate)
			}
			if filter != tt.wantFilter {
				t.Errorf("filter = %q, want %q", filter, tt.wantFilter)
			}
			if flags != tt.wantFlags {
				t.Errorf("flags = %+v, want %+v", flags, tt.wantFlags)
			}
		})
	}
}

func TestMergeTagsDedup(t *testing.T) {
	tests := []struct {
		name   string
		typed  []string
		scoped []string
		want   []string
	}{
		{
			name:   "no overlap",
			typed:  []string{"#a"},
			scoped: []string{"#b", "#c"},
			want:   []string{"#a", "#b", "#c"},
		},
		{
			name:   "case-insensitive dedup preserves typed casing",
			typed:  []string{"#Followup"},
			scoped: []string{"#followup", "#urgent"},
			want:   []string{"#Followup", "#urgent"},
		},
		{
			name:   "empty typed",
			typed:  nil,
			scoped: []string{"#a", "#b"},
			want:   []string{"#a", "#b"},
		},
		{
			name:   "empty scoped",
			typed:  []string{"#a"},
			scoped: nil,
			want:   []string{"#a"},
		},
		{
			name:   "both empty",
			typed:  nil,
			scoped: nil,
			want:   nil,
		},
		{
			name:   "hash prefix normalization",
			typed:  []string{"#x"},
			scoped: []string{"y"},
			want:   []string{"#x", "#y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeTagsDedup(tt.typed, tt.scoped)
			if !slicesEqual(got, tt.want) {
				t.Errorf("mergeTagsDedup(%v, %v) = %v, want %v", tt.typed, tt.scoped, got, tt.want)
			}
		})
	}
}

// TestAllTagsExpansion simulates the --all-tags resolution that
// executePickDialog performs, then verifies the resolved query stored in
// pd.Query can be re-parsed by ReloadPickDialog to produce identical
// pick arguments — without needing PickContext state.
func TestAllTagsExpansion(t *testing.T) {
	tests := []struct {
		name       string
		raw        string        // what the user typed
		scopedTags []string      // simulated ScopedInlineTags() return
		anyToggle  bool          // PickContext.AnyMode
		todoToggle bool          // PickContext.TodoMode
		wantTags   []string      // expected tags after re-parse
		wantAny    bool          // expected --any after re-parse
		wantTodo   bool          // expected --todo after re-parse
		wantDate   string        // expected date after re-parse
		wantNoAllTags bool       // resolved query must NOT contain --all-tags
	}{
		{
			name:          "--all-tags expands scoped tags and forces --any",
			raw:           "--all-tags",
			scopedTags:    []string{"#followup", "#idea", "#urgent"},
			wantTags:      []string{"#followup", "#idea", "#urgent"},
			wantAny:       true,
			wantNoAllTags: true,
		},
		{
			name:          "--all-tags with typed tag deduplicates",
			raw:           "#followup --all-tags",
			scopedTags:    []string{"#followup", "#idea", "#urgent"},
			wantTags:      []string{"#followup", "#idea", "#urgent"},
			wantAny:       true,
			wantNoAllTags: true,
		},
		{
			name:          "--all-tags excludes #done from expansion",
			raw:           "--all-tags",
			scopedTags:    []string{"#followup", "#done", "#idea"},
			wantTags:      []string{"#followup", "#idea"},
			wantAny:       true,
			wantNoAllTags: true,
		},
		{
			name:          "--all-tags with date and --todo",
			raw:           "--all-tags @today --todo",
			scopedTags:    []string{"#followup"},
			wantTags:      []string{"#followup"},
			wantAny:       true,
			wantTodo:      true,
			wantDate:      "@today",
			wantNoAllTags: true,
		},
		{
			name:          "context TodoMode baked in without --todo in raw",
			raw:           "--all-tags",
			scopedTags:    []string{"#followup"},
			todoToggle:    true,
			wantTags:      []string{"#followup"},
			wantAny:       true,
			wantTodo:      true,
			wantNoAllTags: true,
		},
		{
			name:       "no --all-tags: resolved query matches raw parse",
			raw:        "#followup --any",
			scopedTags: []string{"#idea", "#urgent"},
			wantTags:   []string{"#followup"},
			wantAny:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Simulate executePickDialog resolution ---
			tags, date, _, flags := ParsePickQuery(tt.raw)
			allTagsActive := flags.AllTags // (would also check ctx.AllTagsMode)
			anyMode := tt.anyToggle || flags.Any
			todoMode := tt.todoToggle || flags.Todo

			if allTagsActive {
				filtered := make([]string, 0, len(tt.scopedTags))
				for _, tg := range tt.scopedTags {
					if tg != "#done" && tg != "#Done" {
						filtered = append(filtered, tg)
					}
				}
				tags = mergeTagsDedup(tags, filtered)
				anyMode = true
			}

			resolved := buildResolvedQuery(tags, date, anyMode, todoMode)

			// --- Verify resolved query has no --all-tags ---
			if tt.wantNoAllTags {
				rTags, _, _, rFlags := ParsePickQuery(resolved)
				if rFlags.AllTags {
					t.Errorf("resolved query still contains --all-tags: %q", resolved)
				}
				// Verify expanded tags are present
				if !slicesEqual(rTags, tt.wantTags) {
					t.Errorf("re-parsed tags = %v, want %v", rTags, tt.wantTags)
				}
			}

			// --- Simulate ReloadPickDialog: re-parse the resolved query ---
			reloadTags, reloadDate, _, reloadFlags := ParsePickQuery(resolved)

			if !slicesEqual(reloadTags, tt.wantTags) {
				t.Errorf("reload tags = %v, want %v", reloadTags, tt.wantTags)
			}
			if reloadDate != tt.wantDate {
				t.Errorf("reload date = %q, want %q", reloadDate, tt.wantDate)
			}
			if reloadFlags.Any != tt.wantAny {
				t.Errorf("reload Any = %v, want %v", reloadFlags.Any, tt.wantAny)
			}
			if reloadFlags.Todo != tt.wantTodo {
				t.Errorf("reload Todo = %v, want %v", reloadFlags.Todo, tt.wantTodo)
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
