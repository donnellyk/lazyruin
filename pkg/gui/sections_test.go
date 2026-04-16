package gui

import (
	"strings"
	"testing"
)

func TestIsRawHeaderLine(t *testing.T) {
	type tc struct {
		line  string
		level int
		ok    bool
	}
	for _, c := range []tc{
		{"# H1", 1, true},
		{"## H2", 2, true},
		{"#### Deep", 4, true},
		{"  ## indented", 2, true},
		{"#tag", 0, false},
		{"#followup #done", 0, false},
		{"", 0, false},
		{"plain text", 0, false},
	} {
		level, ok := isRawHeaderLine(c.line)
		if ok != c.ok {
			t.Errorf("isRawHeaderLine(%q) ok = %v, want %v", c.line, ok, c.ok)
			continue
		}
		if level != c.level {
			t.Errorf("isRawHeaderLine(%q) level = %d, want %d", c.line, level, c.level)
		}
	}
}

func TestComputeDoneSections(t *testing.T) {
	type expect struct {
		line string
		done bool
	}
	cases := []struct {
		name    string
		content string
		expect  []expect
	}{
		{
			name:    "no headers — nothing is part of a section",
			content: "just prose\n#done only\nplain",
			expect: []expect{
				{"just prose", false},
				{"#done only", false},
				{"plain", false},
			},
		},
		{
			name: "header with #done dims whole section",
			content: strings.Join([]string{
				"# Project #done",
				"",
				"still work left",
				"more stuff",
			}, "\n"),
			expect: []expect{
				{"# Project #done", true},
				{"", true},
				{"still work left", true},
				{"more stuff", true},
			},
		},
		{
			name: "all direct lines tagged → section fully-done",
			content: strings.Join([]string{
				"# Wrap up",
				"task one #done",
				"task two #done",
			}, "\n"),
			expect: []expect{
				{"# Wrap up", true},
				{"task one #done", true},
				{"task two #done", true},
			},
		},
		{
			name: "one untagged line blocks fully-done",
			content: strings.Join([]string{
				"# In progress",
				"task one #done",
				"task two still active",
			}, "\n"),
			expect: []expect{
				{"# In progress", false},
				{"task one #done", false},
				{"task two still active", false},
			},
		},
		{
			name: "mixed blank lines don't count",
			content: strings.Join([]string{
				"# Done bucket",
				"",
				"thing one #done",
				"",
				"thing two #done",
				"",
			}, "\n"),
			expect: []expect{
				{"# Done bucket", true},
				{"", true},
				{"thing one #done", true},
				{"", true},
				{"thing two #done", true},
				{"", true},
			},
		},
		{
			name: "nested sub-section done but parent has ungated line",
			content: strings.Join([]string{
				"# Parent",
				"parent still open",
				"## Child",
				"child task #done",
			}, "\n"),
			expect: []expect{
				{"# Parent", false},
				{"parent still open", false},
				{"## Child", true},
				{"child task #done", true},
			},
		},
		{
			name: "parent done requires sub-section to also be done",
			content: strings.Join([]string{
				"# Parent",
				"direct #done",
				"## Child",
				"child open",
			}, "\n"),
			expect: []expect{
				{"# Parent", false},
				{"direct #done", false},
				{"## Child", false},
				{"child open", false},
			},
		},
		{
			name: "sub-section header #done suffices for the sub",
			content: strings.Join([]string{
				"# Parent",
				"direct #done",
				"## Child #done",
				"child open",
			}, "\n"),
			expect: []expect{
				{"# Parent", true},
				{"direct #done", true},
				{"## Child #done", true},
				{"child open", true},
			},
		},
		{
			// Verifies fence-aware header detection: `# not a header`
			// inside the fence must NOT split the section, and the fence's
			// code lines are treated as non-content. The only direct
			// content is `real task #done`, so the section is fully-done.
			name: "code block with fake headers is ignored (section stays intact)",
			content: strings.Join([]string{
				"# Real",
				"```",
				"# not a header",
				"code line",
				"```",
				"real task #done",
			}, "\n"),
			expect: []expect{
				{"# Real", true},
				{"```", true},
				{"# not a header", true},
				{"code line", true},
				{"```", true},
				{"real task #done", true},
			},
		},
		{
			// Same shape but the outside task is NOT tagged, so the section
			// is blocked. If the fence weren't respected, the fake header
			// would split the section and the result would differ.
			name: "code block doesn't hide an untagged outside line",
			content: strings.Join([]string{
				"# Real",
				"```",
				"# not a header",
				"```",
				"unfinished task",
			}, "\n"),
			expect: []expect{
				{"# Real", false},
				{"```", false},
				{"# not a header", false},
				{"```", false},
				{"unfinished task", false},
			},
		},
		{
			name: "pre-header preamble lines stay outside any section",
			content: strings.Join([]string{
				"preamble line",
				"# After",
				"inside #done",
			}, "\n"),
			expect: []expect{
				{"preamble line", false},
				{"# After", true},
				{"inside #done", true},
			},
		},
		{
			name: "empty-body section (header only) is not fully-done",
			content: strings.Join([]string{
				"# Empty",
				"",
				"",
			}, "\n"),
			expect: []expect{
				{"# Empty", false},
				{"", false},
				{"", false},
			},
		},
		{
			// A parent with only sub-sections (all fully-done) should itself
			// be fully-done — the "empty-body" guard should not fire on
			// "has only sub-sections" because the section is clearly populated.
			name: "parent with only fully-done sub-sections is itself fully-done",
			content: strings.Join([]string{
				"## Feb 25, 2026",
				"### Morning",
				"- task one #done",
				"### Afternoon",
				"- task two #done",
			}, "\n"),
			expect: []expect{
				{"## Feb 25, 2026", true},
				{"### Morning", true},
				{"- task one #done", true},
				{"### Afternoon", true},
				{"- task two #done", true},
			},
		},
		{
			name: "code block inside otherwise fully-done section doesn't block",
			content: strings.Join([]string{
				"# Done with code",
				"task outside #done",
				"```",
				"printf(\"hi\")",
				"```",
			}, "\n"),
			expect: []expect{
				{"# Done with code", true},
				{"task outside #done", true},
				{"```", true},
				{"printf(\"hi\")", true},
				{"```", true},
			},
		},
		{
			name: "tag-only line inside otherwise fully-done section doesn't block",
			content: strings.Join([]string{
				"# Wrapped",
				"#followup, #project",
				"work item #done",
			}, "\n"),
			expect: []expect{
				{"# Wrapped", true},
				{"#followup, #project", true},
				{"work item #done", true},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			contentLines := strings.Split(c.content, "\n")
			got := computeDoneSections(contentLines)
			if len(got) != len(c.expect) {
				t.Fatalf("result length = %d, want %d", len(got), len(c.expect))
			}
			for i, e := range c.expect {
				if contentLines[i] != e.line {
					t.Fatalf("test vector mismatch at line %d: file %q, expect %q", i, contentLines[i], e.line)
				}
				if got[i] != e.done {
					t.Errorf("line %d %q done = %v, want %v", i, e.line, got[i], e.done)
				}
			}
		})
	}
}
