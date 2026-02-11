package gui

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

// contextHintDefs returns the hint definitions for the current context.
// This is the single source of truth consumed by both updateStatusBar() and showHelp().
func (gui *Gui) contextHintDefs() contextHintDef {
	switch gui.state.CurrentContext {
	case NotesContext:
		return contextHintDef{
			header: "Notes",
			hints: []contextHint{
				{"e/enter", "Edit note in $EDITOR"},
				{"E", "Enter edit mode"},
				{"n", "New note"},
				{"d", "Delete note"},
				{"y", "Copy note path"},
				{"1", "Cycle tabs"},
			},
			statusBar: []contextHint{
				{"e/enter", "Edit"},
				{"E", "Edit Mode"},
				{"n", "New"},
				{"d", "Delete"},
				{"/", "Search"},
				{"1", "Tab"},
				{"y", "Copy Path"},
				{"?", "Keybindings"},
			},
		}
	case QueriesContext:
		if gui.state.Queries.CurrentTab == QueriesTabParents {
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
	case TagsContext:
		return contextHintDef{
			header: "Tags",
			hints: []contextHint{
				{"enter", "Filter notes by tag"},
				{"r", "Rename tag"},
				{"d", "Delete tag"},
			},
			statusBar: []contextHint{
				{"enter", "Filter"},
				{"r", "Rename"},
				{"d", "Delete"},
				{"?", "Keybindings"},
			},
		}
	case PreviewContext:
		if gui.state.Preview.EditMode {
			return contextHintDef{
				header: "Edit Mode",
				hints: []contextHint{
					{"d", "Delete card"},
					{"m", "Move card"},
					{"M", "Merge card"},
					{"esc", "Exit edit mode"},
				},
				statusBar: []contextHint{
					{"d", "Delete"},
					{"m", "Move"},
					{"M", "Merge"},
					{"j/k", "Navigate"},
					{"esc", "Back"},
				},
			}
		}
		return contextHintDef{
			header: "Preview",
			hints: []contextHint{
				{"enter", "Focus note"},
				{"f", "Toggle frontmatter"},
				{"t", "Toggle title"},
				{"T", "Toggle global tags"},
				{"M", "Toggle markdown"},
				{"esc", "Back"},
			},
			statusBar: []contextHint{
				{"j/k", "Navigate"},
				{"enter", "Focus Note"},
				{"f", "Frontmatter"},
				{"M", "Markdown"},
				{"esc", "Back"},
				{"?", "Keybindings"},
			},
		}
	case SearchContext:
		return contextHintDef{
			hints: []contextHint{
				{"enter", "Search"},
				{"tab", "Complete"},
				{"esc", "Cancel"},
			},
		}
	case CaptureContext:
		return contextHintDef{
			hints: []contextHint{
				{"<c-s>", "Save"},
				{"esc", "Cancel"},
				{"/", "Formatting"},
				{">", "Parent"},
			},
		}
	case SearchFilterContext:
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
	switch gui.state.CurrentContext {
	case NotesContext, QueriesContext, TagsContext:
		return []contextHint{
			{"j/k", "Move down/up"},
			{"g", "Go to top"},
			{"G", "Go to bottom"},
		}
	case PreviewContext:
		return []contextHint{
			{"j/k", "Scroll down/up"},
		}
	default:
		return nil
	}
}

// globalHints returns the global keybinding hints for the help menu.
func globalHints() []contextHint {
	return []contextHint{
		{"/", "Search"},
		{"p", "Focus preview"},
		{"Tab", "Next panel"},
		{"<c-r>", "Refresh"},
		{"q", "Quit"},
	}
}
