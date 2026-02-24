package helpers

import (
	"fmt"
	"sort"

	"kvnd/lazyruin/pkg/gui/types"
)

// SnippetHelper encapsulates snippet management logic.
type SnippetHelper struct {
	c *HelperCommon
}

func NewSnippetHelper(c *HelperCommon) *SnippetHelper {
	return &SnippetHelper{c: c}
}

// ListSnippets shows all configured abbreviation snippets in a menu dialog.
func (self *SnippetHelper) ListSnippets() error {
	cfg := self.c.Config()
	gui := self.c.GuiCommon()
	vaultPath := self.c.RuinCmd().VaultPath()
	abbrevs := cfg.VaultAbbreviations(vaultPath)

	if len(abbrevs) == 0 {
		gui.ShowError(fmt.Errorf("no snippets configured"))
		return nil
	}

	keys := make([]string, 0, len(abbrevs))
	for k := range abbrevs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var items []types.MenuItem
	for _, k := range keys {
		expansion := abbrevs[k]
		items = append(items, types.MenuItem{
			Key:   "!" + k,
			Label: expansion,
		})
	}

	gui.ShowMenuDialog("Snippets", items)
	return nil
}

// CreateSnippet opens the two-field snippet editor.
func (self *SnippetHelper) CreateSnippet() error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().SnippetEditor
	ctx.Focus = 0
	ctx.Completion = types.NewCompletionState()
	gui.PushContextByKey("snippetName")
	return nil
}

// SaveSnippet validates and saves a new snippet.
func (self *SnippetHelper) SaveSnippet(name, expansion string) error {
	cfg := self.c.Config()
	gui := self.c.GuiCommon()
	vaultPath := self.c.RuinCmd().VaultPath()

	if name == "" {
		gui.ShowError(fmt.Errorf("snippet name cannot be empty"))
		return nil
	}
	if expansion == "" {
		gui.ShowError(fmt.Errorf("expansion cannot be empty"))
		return nil
	}
	abbrevs := cfg.VaultAbbreviations(vaultPath)
	if _, exists := abbrevs[name]; exists {
		gui.ShowError(fmt.Errorf("snippet !%s already exists", name))
		return nil
	}

	cfg.SetVaultAbbreviation(vaultPath, name, expansion)
	return cfg.Save()
}

// CloseEditor tears down the snippet editor views and restores focus.
func (self *SnippetHelper) CloseEditor() error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().SnippetEditor
	ctx.Completion = types.NewCompletionState()
	gui.DeleteView("snippetName")
	gui.DeleteView("snippetExpansion")
	gui.DeleteView("snippetSuggest")
	gui.SetCursorEnabled(false)
	gui.PopContext()
	return nil
}

// DeleteSnippet shows a menu of snippets and deletes the selected one after confirmation.
func (self *SnippetHelper) DeleteSnippet() error {
	cfg := self.c.Config()
	gui := self.c.GuiCommon()
	vaultPath := self.c.RuinCmd().VaultPath()
	abbrevs := cfg.VaultAbbreviations(vaultPath)

	if len(abbrevs) == 0 {
		gui.ShowError(fmt.Errorf("no snippets to delete"))
		return nil
	}

	keys := make([]string, 0, len(abbrevs))
	for k := range abbrevs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var items []types.MenuItem
	for _, k := range keys {
		name := k
		expansion := abbrevs[name]
		detail := expansion
		if len(detail) > 40 {
			detail = detail[:37] + "..."
		}
		items = append(items, types.MenuItem{
			Key:   "!" + name,
			Label: detail,
			OnRun: func() error {
				gui.ShowConfirm("Delete Snippet", fmt.Sprintf("Delete snippet !%s?", name), func() error {
					cfg.DeleteVaultAbbreviation(vaultPath, name)
					return cfg.Save()
				})
				return nil
			},
		})
	}

	gui.ShowMenuDialog("Delete Snippet", items)
	return nil
}
