package gui

import "kvnd/lazyruin/pkg/gui/context"

// contextHint defines a single keybinding hint shared between the status bar and help menu.
type contextHint struct {
	key    string
	action string
}

// contextHintDef holds hints for a specific UI context.
type contextHintDef struct {
	header    string
	hints     []contextHint
	statusBar []contextHint // if set, overrides hints for the status bar (shorter list)
}

// previewNavHints returns navigation hints shared by all preview contexts.
func previewNavHints() []contextHint {
	return []contextHint{
		{"j/k", "Scroll line-by-line"},
		{"J/K", "Jump between cards"},
		{"]/[", "Next/prev header"},
		{"l/L", "Next/prev link"},
		{"o", "Open link"},
		{"f", "Toggle frontmatter"},
		{"v", "View options"},
		{"esc", "Back"},
	}
}

// previewNavStatusBar returns status bar hints shared by all preview contexts.
func previewNavStatusBar() []contextHint {
	return []contextHint{
		{"j/k", "Scroll"},
		{"v", "View"},
		{"esc", "Back"},
		{"?", "Keys"},
	}
}

// contextHintDefs returns the hint definitions for the current context.
// This is the single source of truth consumed by both updateStatusBar() and showHelp().
func (gui *Gui) contextHintDefs() contextHintDef {
	switch gui.contextMgr.Current() {
	case "notes":
		return contextHintDef{
			header: "Notes",
			hints: []contextHint{
				{"enter", "View in preview"},
				{"E", "Open in editor"},
				{"n", "New note"},
				{"d", "Delete note"},
				{"y", "Copy note path"},
				{"t", "Add tag"},
				{"T", "Remove tag"},
				{">", "Set parent"},
				{"P", "Remove parent"},
				{"b", "Toggle bookmark"},
				{"s", "Show info"},
				{"1", "Cycle tabs"},
			},
			statusBar: []contextHint{
				{"enter", "View"},
				{"E", "Editor"},
				{"n", "New"},
				{"d", "Delete"},
				{"t", "Tag"},
				{"b", "Bookmark"},
				{"/", "Search"},
				{"1", "Tab"},
				{"?", "Keybindings"},
			},
		}
	case "queries":
		if gui.contexts.Queries.CurrentTab == context.QueriesTabParents {
			return contextHintDef{
				header: "Parents",
				hints: []contextHint{
					{"enter", "View parent"},
					{"d", "Delete parent"},
					{"2", "Cycle tabs"},
				},
				statusBar: []contextHint{
					{"enter", "View"},
					{"d", "Delete"},
					{"2", "Tab"},
					{"?", "Keybindings"},
				},
			}
		}
		return contextHintDef{
			header: "Queries",
			hints: []contextHint{
				{"enter", "Run query"},
				{"d", "Delete query"},
				{"2", "Cycle tabs"},
			},
			statusBar: []contextHint{
				{"enter", "Run"},
				{"d", "Delete"},
				{"2", "Tab"},
				{"?", "Keybindings"},
			},
		}
	case "tags":
		return contextHintDef{
			header: "Tags",
			hints: []contextHint{
				{"enter", "Filter notes by tag"},
				{"r", "Rename tag"},
				{"d", "Delete tag"},
				{"3", "Cycle tabs"},
			},
			statusBar: []contextHint{
				{"enter", "Filter"},
				{"r", "Rename"},
				{"d", "Delete"},
				{"3", "Tab"},
				{"?", "Keybindings"},
			},
		}
	case "cardList":
		hints := previewNavHints()
		hints = append(hints,
			contextHint{"x", "Toggle todo"},
			contextHint{"d", "Delete card"},
			contextHint{"D", "Append #done"},
			contextHint{"E", "Open in editor"},
			contextHint{"m", "Move card"},
			contextHint{"M", "Merge notes"},
			contextHint{"t", "Add tag"},
			contextHint{"T", "Remove tag"},
			contextHint{"<c-t>", "Toggle inline tag"},
			contextHint{">", "Set parent"},
			contextHint{"P", "Remove parent"},
			contextHint{"b", "Toggle bookmark"},
			contextHint{"s", "Show info"},
			contextHint{"enter", "Focus note"},
		)
		return contextHintDef{
			header: "Preview",
			hints:  hints,
			statusBar: []contextHint{
				{"j/k", "Scroll"},
				{"x", "Todo"},
				{"d", "Del"},
				{"m", "Move"},
				{"t", "Tag"},
				{"v", "View"},
				{"enter", "Focus"},
				{"esc", "Back"},
				{"?", "Keys"},
			},
		}
	case "pickResults":
		hints := previewNavHints()
		return contextHintDef{
			header:    "Pick Results",
			hints:     hints,
			statusBar: previewNavStatusBar(),
		}
	case "compose":
		hints := previewNavHints()
		return contextHintDef{
			header:    "Compose",
			hints:     hints,
			statusBar: previewNavStatusBar(),
		}
	case "search":
		return contextHintDef{
			hints: []contextHint{
				{"enter", "Search"},
				{"tab", "Complete"},
				{"esc", "Cancel"},
			},
		}
	case "capture":
		return contextHintDef{
			hints: []contextHint{
				{"<c-s>", "Save"},
				{"esc", "Cancel"},
				{"/", "Formatting"},
				{">", "Parent"},
			},
		}
	case "pick":
		return contextHintDef{
			hints: []contextHint{
				{"enter", "Pick"},
				{"tab", "Complete"},
				{"<c-a>", "Toggle --any"},
				{"esc", "Cancel"},
			},
		}
	case "palette":
		return contextHintDef{
			hints: []contextHint{
				{"enter", "Execute"},
				{"up/down", "Navigate"},
				{"esc", "Cancel"},
			},
		}
	case "searchFilter":
		return contextHintDef{
			header: "Search Filter",
			hints: []contextHint{
				{"x", "Clear filter"},
			},
			statusBar: []contextHint{
				{"x", "Clear"},
				{"?", "Keybindings"},
			},
		}
	default:
		return contextHintDef{
			hints: []contextHint{
				{"q", "Quit"},
				{"?", "Keybindings"},
			},
		}
	}
}

// navigationHints returns the navigation section hints for the help menu.
func (gui *Gui) navigationHints() []contextHint {
	switch gui.contextMgr.Current() {
	case "notes", "queries", "tags":
		return []contextHint{
			{"j/k", "Move down/up"},
			{"g", "Go to top"},
			{"G", "Go to bottom"},
		}
	case "cardList", "pickResults", "compose":
		return []contextHint{
			{"j/k", "Scroll line-by-line"},
			{"J/K", "Jump between cards"},
			{"]/[", "Next/prev header"},
			{"l/L", "Next/prev link"},
		}
	default:
		return nil
	}
}

// globalHints returns the global keybinding hints for the help menu.
func globalHints() []contextHint {
	return []contextHint{
		{"/", "Search"},
		{"p", "Pick"},
		{":", "Command palette"},
		{"c", "Calendar"},
		{"C", "Contributions"},
		{"Tab", "Next panel"},
		{"<c-r>", "Refresh"},
		{"q", "Quit"},
	}
}
