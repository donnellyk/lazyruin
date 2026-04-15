package gui

import (
	"fmt"
	"testing"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

func TestInsideDynamicEmbed(t *testing.T) {
	tests := []struct {
		name              string
		content           string
		cursor            int
		wantInEmbed       bool
		wantEmbedType     string
		wantInsideOptions bool
	}{
		{"outside", "hello world", 11, false, "", false},
		{"empty buffer", "", 0, false, "", false},
		{"just opened embed", "![[", 3, true, "", false},
		{"inside embed no type", "![[foo", 6, true, "", false},
		{"inside search embed", "![[search: ", 11, true, "search", false},
		{"inside search with query", "![[search: #daily", 17, true, "search", false},
		{"inside pick embed", "![[pick: #followup", 18, true, "pick", false},
		{"inside query embed", "![[query: weekly", 16, true, "query", false},
		{"inside compose embed", "![[compose: alpha", 17, true, "compose", false},
		{"options phase search", "![[search: #daily | ", 20, true, "search", true},
		{"options phase pick", "![[pick: #foo | any, fo", 23, true, "pick", true},
		{"closed embed not matching", "![[search: #foo ]]", 18, false, "", false},
		{"cursor before open", "x ![[foo", 1, false, "", false},
		{"on a different line", "![[\nhello", 9, false, "", false},
		{"previous embed closed, new one open", "![[foo]] ![[bar", 15, true, "", false},
		{"known type via prefix", "![[compose: foo/bar", 19, true, "compose", false},
		{"unknown type falls back to no type", "![[unknown: foo", 15, true, "", false},
		{"cursor at `![[` not `![[ `", "![[", 3, true, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := insideDynamicEmbed(tt.content, tt.cursor)
			if got.inEmbed != tt.wantInEmbed {
				t.Errorf("inEmbed = %v, want %v", got.inEmbed, tt.wantInEmbed)
			}
			if got.embedType != tt.wantEmbedType {
				t.Errorf("embedType = %q, want %q", got.embedType, tt.wantEmbedType)
			}
			if got.insideOptions != tt.wantInsideOptions {
				t.Errorf("insideOptions = %v, want %v", got.insideOptions, tt.wantInsideOptions)
			}
		})
	}
}

func TestEmbedOptionKeyCandidates(t *testing.T) {
	tests := []struct {
		name      string
		embedType string
		filter    string
		wantKeys  []string // expected labels (subset check)
		wantEmpty bool     // if true, result should be nil/empty
	}{
		{
			name:      "search keys",
			embedType: "search",
			wantKeys:  []string{"format=", "sort=", "limit=", "tag-scope=", "depth=", "empty="},
		},
		{
			name:      "pick keys",
			embedType: "pick",
			wantKeys:  []string{"format=", "sort=", "limit=", "filter=", "empty=", "any", "done", "all"},
		},
		{
			name:      "query keys match search",
			embedType: "query",
			wantKeys:  []string{"format=", "sort=", "limit=", "tag-scope=", "depth=", "empty="},
		},
		{
			name:      "compose keys",
			embedType: "compose",
			wantKeys:  []string{"depth="},
		},
		{
			name:      "unknown type",
			embedType: "nope",
			wantEmpty: true,
		},
		{
			name:      "filter by fmt matches format",
			embedType: "search",
			filter:    "fmt",
			wantKeys:  []string{}, // "fmt" not in label or detail
			wantEmpty: true,
		},
		{
			name:      "filter matches sort",
			embedType: "search",
			filter:    "sor",
			wantKeys:  []string{"sort="},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := embedOptionKeyCandidates(tt.embedType, tt.filter)
			if tt.wantEmpty {
				if len(got) != 0 {
					t.Errorf("expected empty, got %d items", len(got))
				}
				return
			}
			labels := make(map[string]bool)
			for _, item := range got {
				labels[item.Label] = true
			}
			for _, want := range tt.wantKeys {
				if !labels[want] {
					t.Errorf("missing key %q in %v", want, labels)
				}
			}
		})
	}
}

func TestEmbedOptionValueCandidates(t *testing.T) {
	tests := []struct {
		name       string
		embedType  string
		key        string
		filter     string
		wantValues []string
		wantEmpty  bool
	}{
		{
			name:       "format for search",
			embedType:  "search",
			key:        "format",
			wantValues: []string{"content", "list", "summary"},
		},
		{
			name:       "format for pick",
			embedType:  "pick",
			key:        "format",
			wantValues: []string{"grouped", "flat"},
		},
		{
			name:       "format for query (same as search)",
			embedType:  "query",
			key:        "format",
			wantValues: []string{"content", "list", "summary"},
		},
		{
			name:       "sort values",
			embedType:  "search",
			key:        "sort",
			wantValues: []string{"created:desc", "created:asc", "title:asc", "updated:desc"},
		},
		{
			name:       "tag-scope values",
			embedType:  "search",
			key:        "tag-scope",
			wantValues: []string{"all", "global", "inline"},
		},
		{
			name:       "empty values",
			embedType:  "search",
			key:        "empty",
			wantValues: []string{"hide"},
		},
		{
			name:      "limit has no completion",
			embedType: "search",
			key:       "limit",
			wantEmpty: true,
		},
		{
			name:      "depth has no completion",
			embedType: "compose",
			key:       "depth",
			wantEmpty: true,
		},
		{
			name:      "filter key has no completion",
			embedType: "pick",
			key:       "filter",
			wantEmpty: true,
		},
		{
			name:       "filter narrows",
			embedType:  "search",
			key:        "format",
			filter:     "li",
			wantValues: []string{"list"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := embedOptionValueCandidates(tt.embedType, tt.key, tt.filter)
			if tt.wantEmpty {
				if len(got) != 0 {
					t.Errorf("expected empty, got %d items", len(got))
				}
				return
			}
			labels := make(map[string]bool)
			for _, item := range got {
				labels[item.Label] = true
			}
			for _, want := range tt.wantValues {
				if !labels[want] {
					t.Errorf("missing value %q in %v", want, labels)
				}
			}
		})
	}
}

// TestEmbedOptionValueCandidates_NilContract locks in the contract used by
// dynamicEmbedCandidates: nil means "free-form value, fall through"; a
// non-nil empty slice means "known key with no matches, don't fall through".
func TestEmbedOptionValueCandidates_NilContract(t *testing.T) {
	tests := []struct {
		name      string
		embedType string
		key       string
		filter    string
		wantNil   bool
	}{
		{"free-form filter=", "pick", "filter", "anything", true},
		{"free-form limit=", "search", "limit", "5", true},
		{"free-form depth=", "compose", "depth", "3", true},
		{"unknown key", "search", "bogus", "", true},
		{"known key with matches", "search", "format", "li", false},
		{"known key with zero matches returns empty (not nil)", "search", "format", "zzz-no-match", false},
		{"known key empty filter", "pick", "format", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := embedOptionValueCandidates(tt.embedType, tt.key, tt.filter)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil (free-form signal), got %v", got)
				}
				return
			}
			if got == nil {
				t.Errorf("expected non-nil (known key), got nil — this would trigger fall-through")
			}
		})
	}
}

func TestDetectTrigger_EmbedPrefix(t *testing.T) {
	embedTrigger := types.CompletionTrigger{
		Prefix: "![[",
		Candidates: func(filter string) []types.CompletionItem {
			return []types.CompletionItem{{Label: "embed"}}
		},
	}
	wikiTrigger := types.CompletionTrigger{
		Prefix: "[[",
		Candidates: func(filter string) []types.CompletionItem {
			return []types.CompletionItem{{Label: "wiki"}}
		},
	}
	abbrevTrigger := types.CompletionTrigger{
		Prefix: "!",
		Candidates: func(filter string) []types.CompletionItem {
			return []types.CompletionItem{{Label: "abbr"}}
		},
	}

	tests := []struct {
		name       string
		triggers   []types.CompletionTrigger
		content    string
		cursor     int
		wantPrefix string
		wantFilter string
	}{
		{
			name:       "embed wins over wiki when prefix order is embed-first",
			triggers:   []types.CompletionTrigger{embedTrigger, wikiTrigger},
			content:    "![[foo",
			cursor:     6,
			wantPrefix: "![[",
			wantFilter: "foo",
		},
		{
			name:       "embed wins over abbreviation when prefix order is embed-first",
			triggers:   []types.CompletionTrigger{embedTrigger, abbrevTrigger},
			content:    "![[bar",
			cursor:     6,
			wantPrefix: "![[",
			wantFilter: "bar",
		},
		{
			name:       "wiki-only fires for [[ without !",
			triggers:   []types.CompletionTrigger{embedTrigger, wikiTrigger},
			content:    "[[wikilink",
			cursor:     10,
			wantPrefix: "[[",
			wantFilter: "wikilink",
		},
		{
			name:       "fallback scan finds ![[ with spaces in filter",
			triggers:   []types.CompletionTrigger{embedTrigger, wikiTrigger},
			content:    "![[search: #daily",
			cursor:     17,
			wantPrefix: "![[",
			wantFilter: "search: #daily",
		},
		{
			name:       "closed ![[ does not match; new [[ does",
			triggers:   []types.CompletionTrigger{embedTrigger, wikiTrigger},
			content:    "![[done]] [[new",
			cursor:     15,
			wantPrefix: "[[",
			wantFilter: "new",
		},
		{
			// Verifies the embed scan is line-limited. The `[[` scan still
			// crosses lines (pre-existing behavior, out of scope), so this
			// test asserts the prefix is NOT `![[` — whether `[[` matches
			// is incidental.
			name:       "unclosed ![[ on previous line does not match as embed",
			triggers:   []types.CompletionTrigger{embedTrigger, wikiTrigger},
			content:    "![[\nhello",
			cursor:     9,
			wantPrefix: "[[",
			wantFilter: "\nhello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger, filter, _ := detectTrigger(tt.content, tt.cursor, tt.triggers)
			if tt.wantPrefix == "" {
				if trigger != nil {
					t.Errorf("expected nil trigger, got prefix %q filter %q", trigger.Prefix, filter)
				}
				return
			}
			if trigger == nil {
				t.Fatal("expected non-nil trigger")
			}
			if trigger.Prefix != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", trigger.Prefix, tt.wantPrefix)
			}
			if filter != tt.wantFilter {
				t.Errorf("filter = %q, want %q", filter, tt.wantFilter)
			}
		})
	}
}

func TestDetectTrigger_HashAtOptionBoundaries(t *testing.T) {
	hashTrigger := types.CompletionTrigger{
		Prefix: "#",
		Candidates: func(filter string) []types.CompletionItem {
			return []types.CompletionItem{{Label: "hash"}}
		},
	}
	triggers := []types.CompletionTrigger{hashTrigger}

	tests := []struct {
		name       string
		content    string
		cursor     int
		wantPrefix string
		wantFilter string
	}{
		{"`#` after `=` (filter=#tag)", "filter=#r", 9, "#", "r"},
		{"`#` after `,`", "sort=created,#foo", 17, "#", "foo"},
		{"`#` right after `|`", "pick: a |#todo", 14, "#", "todo"},
		{"`#` mid-word does NOT match", "word#nottag", 11, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger, filter, _ := detectTrigger(tt.content, tt.cursor, triggers)
			if tt.wantPrefix == "" {
				if trigger != nil {
					t.Errorf("expected nil trigger, got %q", trigger.Prefix)
				}
				return
			}
			if trigger == nil {
				t.Fatal("expected non-nil trigger")
			}
			if trigger.Prefix != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", trigger.Prefix, tt.wantPrefix)
			}
			if filter != tt.wantFilter {
				t.Errorf("filter = %q, want %q", filter, tt.wantFilter)
			}
		})
	}
}

func TestIsTriggerBoundary(t *testing.T) {
	tests := []struct {
		content string
		i       int
		want    bool
	}{
		{"abc", 0, true},   // start of string
		{" abc", 1, true},  // after space
		{"\tabc", 1, true}, // after tab
		{"a\nb", 2, true},  // after newline
		{"!tag", 1, true},  // negation
		{"a=b", 2, true},   // after `=`
		{"a,b", 2, true},   // after `,`
		{"a|b", 2, true},   // after `|`
		{"word", 2, false}, // mid-word
		{"foo#", 3, false}, // `#` itself is not preceded by boundary
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q/%d", tt.content, tt.i), func(t *testing.T) {
			got := isTriggerBoundary(tt.content, tt.i)
			if got != tt.want {
				t.Errorf("isTriggerBoundary(%q, %d) = %v, want %v", tt.content, tt.i, got, tt.want)
			}
		})
	}
}

func TestWithoutAbbreviationTrigger(t *testing.T) {
	triggers := []types.CompletionTrigger{
		{Prefix: "!"},
		{Prefix: "![["},
		{Prefix: "#"},
		{Prefix: "@"},
	}
	got := withoutAbbreviationTrigger(triggers)
	if len(got) != 3 {
		t.Fatalf("expected 3 triggers, got %d", len(got))
	}
	for _, tr := range got {
		if tr.Prefix == "!" {
			t.Errorf("abbreviation trigger should be removed, still present")
		}
	}
}

func TestDynamicEmbedFilter(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		cursor    int
		state     embedState
		wantFilt  string
		wantKey   string
		wantStart int
	}{
		{
			name:      "query partial",
			content:   "![[query: weekl",
			cursor:    15,
			state:     embedState{inEmbed: true, embedType: "query"},
			wantFilt:  "weekl",
			wantStart: 10,
		},
		{
			name:      "compose partial",
			content:   "![[compose: alpha",
			cursor:    17,
			state:     embedState{inEmbed: true, embedType: "compose"},
			wantFilt:  "alpha",
			wantStart: 12,
		},
		{
			name:      "options key phase",
			content:   "![[search: #foo | form",
			cursor:    22,
			state:     embedState{inEmbed: true, embedType: "search", insideOptions: true},
			wantFilt:  "form",
			wantStart: 18,
		},
		{
			name:      "options value phase",
			content:   "![[search: #foo | format=lis",
			cursor:    28,
			state:     embedState{inEmbed: true, embedType: "search", insideOptions: true},
			wantFilt:  "lis",
			wantKey:   "format",
			wantStart: 25,
		},
		{
			name:      "second option after comma",
			content:   "![[search: #foo | sort=created:desc, limit=",
			cursor:    43,
			state:     embedState{inEmbed: true, embedType: "search", insideOptions: true},
			wantFilt:  "",
			wantKey:   "limit",
			wantStart: 43,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, key, start := dynamicEmbedFilter(tt.content, tt.cursor, tt.state)
			if filter != tt.wantFilt {
				t.Errorf("filter = %q, want %q", filter, tt.wantFilt)
			}
			if key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
			if start != tt.wantStart {
				t.Errorf("start = %d, want %d", start, tt.wantStart)
			}
		})
	}
}
