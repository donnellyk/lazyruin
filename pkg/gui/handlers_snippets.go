package gui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jesseduffield/gocui"
)

// listSnippets shows all configured abbreviation snippets in a menu dialog.
func (gui *Gui) listSnippets() error {
	if len(gui.config.Abbreviations) == 0 {
		gui.showError(fmt.Errorf("no snippets configured"))
		return nil
	}

	keys := make([]string, 0, len(gui.config.Abbreviations))
	for k := range gui.config.Abbreviations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var items []MenuItem
	for _, k := range keys {
		expansion := gui.config.Abbreviations[k]
		items = append(items, MenuItem{
			Key:   "!" + k,
			Label: expansion,
		})
	}

	gui.state.Dialog = &DialogState{
		Active:    true,
		Type:      "menu",
		Title:     "Snippets",
		MenuItems: items,
	}
	return nil
}

// snippetExpansionTriggers returns the union of search and capture triggers,
// excluding the ! abbreviation trigger to avoid recursion.
// The > trigger is re-bound to use SnippetEditorCompletion for drill tracking.
func (gui *Gui) snippetExpansionTriggers() []CompletionTrigger {
	seen := make(map[string]bool)
	var triggers []CompletionTrigger
	for _, t := range gui.searchTriggers() {
		if t.Prefix == "!" {
			continue
		}
		if !seen[t.Prefix] {
			seen[t.Prefix] = true
			triggers = append(triggers, t)
		}
	}
	for _, t := range gui.captureTriggers() {
		if t.Prefix == "!" || t.Prefix == ">" {
			continue
		}
		if !seen[t.Prefix] {
			seen[t.Prefix] = true
			triggers = append(triggers, t)
		}
	}
	// Add > trigger bound to the snippet editor's own completion state
	if !seen[">"] {
		triggers = append(triggers, CompletionTrigger{
			Prefix:     ">",
			Candidates: gui.parentCandidatesFor(gui.state.SnippetEditorCompletion),
		})
	}
	return triggers
}

// acceptSnippetParentCompletion accepts a parent completion but keeps the >path
// token in the content (unlike acceptParentCompletion which removes it).
// Snippets store >path literally so it can be resolved when the abbreviation is expanded.
func (gui *Gui) acceptSnippetParentCompletion(v *gocui.View, state *CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}

	item := state.Items[state.SelectedIndex]
	cursorPos := viewCursorBytePos(v)
	content := v.TextArea.GetUnwrappedContent()

	// Backspace from cursor to trigger start
	charsToDelete := cursorPos - state.TriggerStart
	for range charsToDelete {
		v.TextArea.BackSpaceChar()
	}

	// Rebuild the full >path token, preserving >> prefix
	prefix := ">"
	triggerEnd := state.TriggerStart + 2
	if triggerEnd <= len(content) && content[state.TriggerStart:triggerEnd] == ">>" {
		prefix = ">>"
	}
	var path strings.Builder
	path.WriteString(prefix)
	for _, entry := range state.ParentDrill {
		path.WriteString(entry.Name)
		path.WriteByte('/')
	}
	path.WriteString(item.Label)

	v.TextArea.TypeString(path.String() + " ")

	// Clear completion and drill state
	state.Active = false
	state.Items = nil
	state.SelectedIndex = 0
	state.ParentDrill = nil

	v.RenderTextArea()
}

// createSnippet opens the two-field stacked snippet editor.
func (gui *Gui) createSnippet() error {
	gui.state.SnippetEditorMode = true
	gui.state.SnippetEditorFocus = 0
	gui.state.SnippetEditorCompletion = NewCompletionState()
	return nil
}

// snippetEditorTab toggles focus between name and expansion fields,
// or accepts completion if active in the expansion field.
func (gui *Gui) snippetEditorTab(g *gocui.Gui, v *gocui.View) error {
	state := gui.state.SnippetEditorCompletion

	// If completion is active in the expansion field, accept it
	if state.Active && gui.state.SnippetEditorFocus == 1 {
		ev, _ := g.View(SnippetExpansionView)
		if ev != nil {
			if isParentCompletion(ev, state) {
				gui.acceptSnippetParentCompletion(ev, state)
			} else {
				gui.acceptCompletion(ev, state, gui.snippetExpansionTriggers())
			}
			ev.RenderTextArea()
		}
		return nil
	}

	// Toggle focus
	if gui.state.SnippetEditorFocus == 0 {
		gui.state.SnippetEditorFocus = 1
	} else {
		gui.state.SnippetEditorFocus = 0
	}
	return nil
}

// snippetEditorEnter saves the snippet (or accepts completion if active).
func (gui *Gui) snippetEditorEnter(g *gocui.Gui, v *gocui.View) error {
	state := gui.state.SnippetEditorCompletion

	// Accept completion if active
	if state.Active {
		ev, _ := g.View(SnippetExpansionView)
		if ev != nil {
			if isParentCompletion(ev, state) {
				gui.acceptSnippetParentCompletion(ev, state)
			} else {
				gui.acceptCompletion(ev, state, gui.snippetExpansionTriggers())
			}
			ev.RenderTextArea()
		}
		return nil
	}

	// Read both fields
	nv, _ := g.View(SnippetNameView)
	ev, _ := g.View(SnippetExpansionView)
	if nv == nil || ev == nil {
		return nil
	}

	name := strings.TrimLeft(strings.TrimSpace(nv.TextArea.GetUnwrappedContent()), "!")
	expansion := strings.TrimSpace(ev.TextArea.GetUnwrappedContent())

	if name == "" {
		gui.showError(fmt.Errorf("snippet name cannot be empty"))
		return nil
	}
	if expansion == "" {
		gui.showError(fmt.Errorf("expansion cannot be empty"))
		return nil
	}
	if gui.config.Abbreviations != nil {
		if _, exists := gui.config.Abbreviations[name]; exists {
			gui.showError(fmt.Errorf("snippet !%s already exists", name))
			return nil
		}
	}

	if gui.config.Abbreviations == nil {
		gui.config.Abbreviations = make(map[string]string)
	}
	gui.config.Abbreviations[name] = expansion
	if err := gui.config.Save(); err != nil {
		return err
	}

	return gui.closeSnippetEditor(g)
}

// snippetEditorClickName sets focus to the name field on mouse click.
func (gui *Gui) snippetEditorClickName(g *gocui.Gui, v *gocui.View) error {
	gui.state.SnippetEditorFocus = 0
	return nil
}

// snippetEditorClickExpansion sets focus to the expansion field on mouse click.
func (gui *Gui) snippetEditorClickExpansion(g *gocui.Gui, v *gocui.View) error {
	gui.state.SnippetEditorFocus = 1
	return nil
}

// snippetEditorEsc dismisses completion or closes the editor.
func (gui *Gui) snippetEditorEsc(g *gocui.Gui, v *gocui.View) error {
	state := gui.state.SnippetEditorCompletion
	if state.Active {
		state.Active = false
		state.Items = nil
		state.SelectedIndex = 0
		return nil
	}
	return gui.closeSnippetEditor(g)
}

// closeSnippetEditor tears down the snippet editor views and restores focus.
func (gui *Gui) closeSnippetEditor(g *gocui.Gui) error {
	gui.state.SnippetEditorMode = false
	gui.state.SnippetEditorCompletion = NewCompletionState()
	g.DeleteView(SnippetNameView)
	g.DeleteView(SnippetExpansionView)
	g.DeleteView(SnippetSuggestView)
	g.Cursor = false
	gui.setContext(gui.state.PreviousContext)
	return nil
}

// deleteSnippet shows a menu of snippets and deletes the selected one after confirmation.
func (gui *Gui) deleteSnippet() error {
	if len(gui.config.Abbreviations) == 0 {
		gui.showError(fmt.Errorf("no snippets to delete"))
		return nil
	}

	keys := make([]string, 0, len(gui.config.Abbreviations))
	for k := range gui.config.Abbreviations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var items []MenuItem
	for _, k := range keys {
		name := k
		expansion := gui.config.Abbreviations[name]
		detail := expansion
		if len(detail) > 40 {
			detail = detail[:37] + "..."
		}
		items = append(items, MenuItem{
			Key:   "!" + name,
			Label: detail,
			OnRun: func() error {
				gui.showConfirm("Delete Snippet", fmt.Sprintf("Delete snippet !%s?", name), func() error {
					delete(gui.config.Abbreviations, name)
					return gui.config.Save()
				})
				return nil
			},
		})
	}

	gui.state.Dialog = &DialogState{
		Active:    true,
		Type:      "menu",
		Title:     "Delete Snippet",
		MenuItems: items,
	}
	return nil
}
