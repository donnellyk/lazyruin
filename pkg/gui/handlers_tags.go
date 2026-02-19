package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// selectedFilteredTag delegates to TagsContext.
func (gui *Gui) selectedFilteredTag() *models.Tag {
	return gui.contexts.Tags.Selected()
}

// switchTagsTabByIndex handles mouse clicks on tab headers.
func (gui *Gui) switchTagsTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(tagsTabsNew) {
		return nil
	}
	tagsCtx := gui.contexts.Tags
	tagsCtx.CurrentTab = tagsTabsNew[tabIndex]
	tagsCtx.SetSelectedLineIdx(0)
	gui.updateTagsTab()
	gui.renderTags()
	gui.updatePreviewForTags()
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

	gui.helpers.Preview().PushNavHistory()
	gui.helpers.Preview().ShowCardList(" Tag: #"+tag.Name+" ", notes)
	gui.setContext(PreviewContext)
	return nil
}

func (gui *Gui) filterByTagPick(tag *models.Tag) error {
	results, err := gui.ruinCmd.Pick.Pick([]string{tag.Name}, false, "")
	if err != nil {
		gui.showError(err)
		return nil
	}

	gui.helpers.Preview().ShowPickResults(" Pick: #"+tag.Name+" ", results)
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
		gui.RefreshTags(false)
		gui.RefreshNotes(false)
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
		gui.RefreshTags(false)
		gui.RefreshNotes(false)
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

	gui.helpers.Preview().UpdatePreviewCardList(" Tag: #"+tag.Name+" ", func() ([]models.Note, error) {
		return gui.ruinCmd.Search.Search(tag.Name, gui.buildSearchOptions())
	})
}

func (gui *Gui) updatePreviewPickResults(tag *models.Tag) {
	results, err := gui.ruinCmd.Pick.Pick([]string{tag.Name}, false, "")
	if err != nil {
		return
	}

	gui.helpers.Preview().ShowPickResults(" Pick: #"+tag.Name+" ", results)
}
