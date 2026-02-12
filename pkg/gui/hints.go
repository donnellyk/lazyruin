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
				{"enter", "View in preview"},
				{"E", "Open in editor"},
				{"n", "New note"},
				{"d", "Delete note"},
				{"y", "Copy note path"},
				{"1", "Cycle tabs"},
			},
			statusBar: []contextHint{
				{"enter", "View"},
				{"E", "Editor"},
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
	case PreviewContext:
		return contextHintDef{
			header: "Preview",
			hints: []contextHint{
				{"j/k", "Scroll line-by-line"},
				{"J/K", "Jump between cards"},
				{"]/[", "Next/prev header"},
				{"x", "Toggle todo"},
				{"d", "Delete card"},
				{"m", "Move card"},
				{"enter", "Focus note"},
				{"f", "Toggle frontmatter"},
				{"t", "Toggle title"},
				{"T", "Toggle global tags"},
				{"M", "Toggle markdown"},
				{"esc", "Back"},
			},
			statusBar: []contextHint{
				{"j/k", "Scroll"},
				{"J/K", "Card"},
				{"x", "Todo"},
				{"d", "Delete"},
				{"m", "Move"},
				{"enter", "Focus Note"},
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
	case PickContext:
		return contextHintDef{
			hints: []contextHint{
				{"enter", "Pick"},
				{"tab", "Complete"},
				{"<c-a>", "Toggle --any"},
				{"esc", "Cancel"},
			},
		}
	case PaletteContext:
		return contextHintDef{
			hints: []contextHint{
				{"enter", "Execute"},
				{"up/down", "Navigate"},
				{"esc", "Cancel"},
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
			{"j/k", "Scroll line-by-line"},
			{"J/K", "Jump between cards"},
			{"]/[", "Next/prev header"},
			{"x", "Toggle todo"},
		}
	default:
		return nil
	}
}

// globalHints returns the global keybinding hints for the help menu.
func globalHints() []contextHint {
	return []contextHint{
		{"/", "Search"},
		{"\\", "Pick"},
		{":", "Command palette"},
		{"p", "Focus preview"},
		{"Tab", "Next panel"},
		{"<c-r>", "Refresh"},
		{"q", "Quit"},
	}
}
