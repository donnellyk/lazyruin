package helpers

import (
	"strings"
	"testing"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/types"
)

func TestDepthTreePrefixes(t *testing.T) {
	tests := []struct {
		name  string
		items []depthLabel
		want  []string
	}{
		{
			name: "single root",
			items: []depthLabel{
				{0, "Root"},
			},
			want: []string{"Root"},
		},
		{
			name: "root with children",
			items: []depthLabel{
				{0, "Daily Log"},
				{1, "Feb 25, 2026"},
				{2, "Fix"},
				{1, "Feb 24, 2026"},
			},
			want: []string{
				"Daily Log",
				"├─ Feb 25, 2026",
				"│  └─ Fix",
				"└─ Feb 24, 2026",
			},
		},
		{
			name: "multiple roots with children",
			items: []depthLabel{
				{0, "Header A"},
				{1, "Sub A1"},
				{1, "Sub A2"},
				{0, "Header B"},
				{1, "Sub B1"},
			},
			want: []string{
				"Header A",
				"├─ Sub A1",
				"└─ Sub A2", // last at depth 1 before Header B (depth 0)
				"Header B",
				"└─ Sub B1",
			},
		},
		{
			name: "deep nesting",
			items: []depthLabel{
				{0, "Root"},
				{1, "A"},
				{2, "A1"},
				{3, "A1a"},
				{1, "B"},
			},
			want: []string{
				"Root",
				"├─ A",
				"│  └─ A1",
				"│     └─ A1a",
				"└─ B",
			},
		},
		{
			name: "continuation lines from open ancestors",
			items: []depthLabel{
				{0, "Root"},
				{1, "A"},
				{2, "A1"},
				{2, "A2"},
				{1, "B"},
				{2, "B1"},
			},
			want: []string{
				"Root",
				"├─ A",
				"│  ├─ A1",
				"│  └─ A2",
				"└─ B",
				"   └─ B1",
			},
		},
		{
			name:  "empty",
			items: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := depthTreePrefixes(tt.items)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d\ngot:\n%s", len(got), len(tt.want), formatLines(got))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d:\n  got:  %q\n  want: %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestAppendTreeItems_BoxDrawing(t *testing.T) {
	children := []commands.TreeNode{
		{Title: "Daily Log", UUID: "a", Children: []commands.TreeNode{
			{Title: "Feb 25", UUID: "b", Children: []commands.TreeNode{
				{Title: "Fix", UUID: "c"},
			}},
			{Title: "Feb 24", UUID: "d"},
		}},
		{Title: "Ruin Log", UUID: "e"},
	}

	helper := &PreviewInfoHelper{}
	items := helper.appendTreeItems(nil, children, "", 5)

	want := []string{
		"├─ Daily Log",
		"│  ├─ Feb 25",
		"│  │  └─ Fix",
		"│  └─ Feb 24",
		"└─ Ruin Log",
	}

	if len(items) != len(want) {
		t.Fatalf("len = %d, want %d", len(items), len(want))
	}
	for i, item := range items {
		if item.Label != want[i] {
			t.Errorf("item %d:\n  got:  %q\n  want: %q", i, item.Label, want[i])
		}
	}
}

func TestTreeConnector(t *testing.T) {
	conn, childPfx := treeConnector("│  ", false)
	if conn != "│  ├─ " {
		t.Errorf("mid connector = %q, want %q", conn, "│  ├─ ")
	}
	if childPfx != "│  │  " {
		t.Errorf("mid childPrefix = %q, want %q", childPfx, "│  │  ")
	}

	conn, childPfx = treeConnector("│  ", true)
	if conn != "│  └─ " {
		t.Errorf("last connector = %q, want %q", conn, "│  └─ ")
	}
	if childPfx != "│     " {
		t.Errorf("last childPrefix = %q, want %q", childPfx, "│     ")
	}
}

func formatLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// Verify appendTreeItems doesn't need a real HelperCommon (OnRun can be nil for label tests).
func TestAppendTreeItems_MaxDepth(t *testing.T) {
	children := []commands.TreeNode{
		{Title: "A", UUID: "1", Children: []commands.TreeNode{
			{Title: "B", UUID: "2", Children: []commands.TreeNode{
				{Title: "C", UUID: "3"},
			}},
		}},
	}

	helper := &PreviewInfoHelper{}
	items := helper.appendTreeItems(nil, children, "", 1)

	// maxDepth=1 means only one level of children
	if len(items) != 1 {
		labels := make([]string, len(items))
		for i, it := range items {
			labels[i] = it.Label
		}
		t.Fatalf("expected 1 item at maxDepth=1, got %d: %v", len(items), labels)
	}
	if items[0].Label != "└─ A" {
		t.Errorf("got %q, want %q", items[0].Label, "└─ A")
	}
}

// Verify the MenuItem.OnRun closures are nil-safe when PreviewInfoHelper.c is nil.
// appendTreeItems sets OnRun closures that capture childUUID; we just verify labels here.
var _ = types.MenuItem{}
