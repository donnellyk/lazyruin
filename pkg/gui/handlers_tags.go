package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// filteredTagItems delegates to TagsContext.
func (gui *Gui) filteredTagItems() []models.Tag {
	return gui.contexts.Tags.FilteredItems()
}

// selectedFilteredTag delegates to TagsContext.
func (gui *Gui) selectedFilteredTag() *models.Tag {
	return gui.contexts.Tags.Selected()
}

// cycleTagsTab cycles through All -> Global -> Inline tabs.
func (gui *Gui) cycleTagsTab() {
	tagsCtx := gui.contexts.Tags
	idx := (tagsCtx.TabIndex() + 1) % len(tagsTabs)
	tagsCtx.CurrentTab = tagsTabsNew[idx]
	tagsCtx.SetSelectedLineIdx(0)
	gui.syncTagsToLegacy()
	gui.updateTagsTab()
	gui.renderTags()
	gui.updatePreviewForTags()
}

// switchTagsTabByIndex handles mouse clicks on tab headers.
func (gui *Gui) switchTagsTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(tagsTabs) {
		return nil
	}
	tagsCtx := gui.contexts.Tags
	tagsCtx.CurrentTab = tagsTabsNew[tabIndex]
	tagsCtx.SetSelectedLineIdx(0)
	gui.syncTagsToLegacy()
	gui.updateTagsTab()
	gui.renderTags()
	gui.updatePreviewForTags()
	gui.setContext(TagsContext)
	return nil
}

func (gui *Gui) tagsClick(g *gocui.Gui, v *gocui.View) error {
	tagsCtx := gui.contexts.Tags
	idx := listClickIndex(gui.views.Tags, 1)
	items := tagsCtx.FilteredItems()
	if idx >= 0 && idx < len(items) {
		tagsCtx.SetSelectedLineIdx(idx)
		gui.syncTagsToLegacy()
	}
	gui.setContext(TagsContext)
	return nil
}

func (gui *Gui) filterByTag(g *gocui.Gui, v *gocui.View) error {
	tag := gui.selectedFilteredTag()
	if tag == nil {
		return nil
	}

	tagsCtx := gui.contexts.Tags
	if tagsCtx.CurrentTab == "inline" {
		return gui.filterByTagPick(tag)
	}
	return gui.filterByTagSearch(tag)
}

func (gui *Gui) filterByTagSearch(tag *models.Tag) error {
	opts := gui.buildSearchOptions()
	notes, err := gui.ruinCmd.Search.Search(tag.Name, opts)
	if err != nil {
		gui.showError(err)
		return nil
	}

	gui.preview.pushNavHistory()
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.views.Preview.Title = " Tag: #" + tag.Name + " "
	gui.renderPreview()
	gui.setContext(PreviewContext)
	return nil
}

func (gui *Gui) filterByTagPick(tag *models.Tag) error {
	results, err := gui.ruinCmd.Pick.Pick([]string{tag.Name}, false, "")
	if err != nil {
		gui.showError(err)
		return nil
	}

	gui.state.Preview.Mode = PreviewModePickResults
	gui.state.Preview.PickResults = results
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.CursorLine = 1
	gui.state.Preview.ScrollOffset = 0
	gui.views.Preview.Title = " Pick: #" + tag.Name + " "
	gui.renderPreview()
	gui.setContext(PreviewContext)
	return nil
}

func (gui *Gui) renameTag(g *gocui.Gui, v *gocui.View) error {
	tag := gui.selectedFilteredTag()
	if tag == nil {
		return nil
	}

	gui.showInput("Rename Tag", "New name for #"+tag.Name+":", func(newName string) error {
		if newName == "" || newName == tag.Name {
			return nil
		}
		err := gui.ruinCmd.Tags.Rename(tag.Name, newName)
		if err != nil {
			gui.showError(err)
			return nil
		}
		gui.refreshTags(false)
		gui.refreshNotes(false)
		return nil
	})
	return nil
}

func (gui *Gui) deleteTag(g *gocui.Gui, v *gocui.View) error {
	tag := gui.selectedFilteredTag()
	if tag == nil {
		return nil
	}

	gui.showConfirm("Delete Tag", "Delete #"+tag.Name+" from all notes?", func() error {
		err := gui.ruinCmd.Tags.Delete(tag.Name)
		if err != nil {
			gui.showError(err)
			return nil
		}
		gui.refreshTags(false)
		gui.refreshNotes(false)
		return nil
	})
	return nil
}

func (gui *Gui) updatePreviewForTags() {
	tag := gui.selectedFilteredTag()
	if tag == nil {
		return
	}

	tagsCtx := gui.contexts.Tags
	if tagsCtx.CurrentTab == "inline" {
		gui.updatePreviewPickResults(tag)
		return
	}

	gui.preview.updatePreviewCardList(" Tag: #"+tag.Name+" ", func() ([]models.Note, error) {
		return gui.ruinCmd.Search.Search(tag.Name, gui.buildSearchOptions())
	})
}

func (gui *Gui) updatePreviewPickResults(tag *models.Tag) {
	results, err := gui.ruinCmd.Pick.Pick([]string{tag.Name}, false, "")
	if err != nil {
		return
	}

	gui.state.Preview.Mode = PreviewModePickResults
	gui.state.Preview.PickResults = results
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.CursorLine = 1
	gui.state.Preview.ScrollOffset = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Pick: #" + tag.Name + " "
	}
	gui.renderPreview()
}
