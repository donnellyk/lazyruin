package gui

import (
	"slices"

	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// filteredTagItems returns tags visible under the current tab.
func (gui *Gui) filteredTagItems() []models.Tag {
	switch gui.state.Tags.CurrentTab {
	case TagsTabGlobal:
		return filterTagsByScope(gui.state.Tags.Items, "global")
	case TagsTabInline:
		return filterTagsByScope(gui.state.Tags.Items, "inline")
	default:
		return gui.state.Tags.Items
	}
}

func filterTagsByScope(tags []models.Tag, scope string) []models.Tag {
	var out []models.Tag
	for _, t := range tags {
		if slices.Contains(t.Scope, scope) {
			out = append(out, t)
		}
	}
	return out
}

// selectedFilteredTag returns the currently selected tag from the filtered list, or nil.
func (gui *Gui) selectedFilteredTag() *models.Tag {
	items := gui.filteredTagItems()
	if len(items) == 0 {
		return nil
	}
	idx := gui.state.Tags.SelectedIndex
	if idx >= len(items) {
		idx = 0
	}
	return &items[idx]
}

func (gui *Gui) tagsPanel() *listPanel {
	return &listPanel{
		selectedIndex: &gui.state.Tags.SelectedIndex,
		itemCount:     func() int { return len(gui.filteredTagItems()) },
		render:        gui.renderTags,
		updatePreview: gui.updatePreviewForTags,
		context:       TagsContext,
	}
}

// cycleTagsTab cycles through All -> Global -> Inline tabs.
func (gui *Gui) cycleTagsTab() {
	idx := (gui.tagsTabIndex() + 1) % len(tagsTabs)
	gui.state.Tags.CurrentTab = tagsTabs[idx]
	gui.state.Tags.SelectedIndex = 0
	gui.updateTagsTab()
	gui.renderTags()
	gui.updatePreviewForTags()
}

// switchTagsTabByIndex handles mouse clicks on tab headers.
func (gui *Gui) switchTagsTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(tagsTabs) {
		return nil
	}
	gui.state.Tags.CurrentTab = tagsTabs[tabIndex]
	gui.state.Tags.SelectedIndex = 0
	gui.updateTagsTab()
	gui.renderTags()
	gui.updatePreviewForTags()
	gui.setContext(TagsContext)
	return nil
}

func (gui *Gui) tagsDown(g *gocui.Gui, v *gocui.View) error {
	return gui.tagsPanel().listDown(g, v)
}

func (gui *Gui) tagsUp(g *gocui.Gui, v *gocui.View) error {
	return gui.tagsPanel().listUp(g, v)
}

func (gui *Gui) tagsClick(g *gocui.Gui, v *gocui.View) error {
	idx := listClickIndex(v, 1)
	items := gui.filteredTagItems()
	if idx >= 0 && idx < len(items) {
		gui.state.Tags.SelectedIndex = idx
	}
	gui.setContext(TagsContext)
	return nil
}

func (gui *Gui) tagsWheelDown(g *gocui.Gui, v *gocui.View) error {
	scrollViewport(v, 3)
	return nil
}

func (gui *Gui) tagsWheelUp(g *gocui.Gui, v *gocui.View) error {
	scrollViewport(v, -3)
	return nil
}

func (gui *Gui) filterByTag(g *gocui.Gui, v *gocui.View) error {
	tag := gui.selectedFilteredTag()
	if tag == nil {
		return nil
	}

	if gui.state.Tags.CurrentTab == TagsTabInline {
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

	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.views.Preview.Title = " Preview: #" + tag.Name + " "
	gui.renderPreview()
	gui.setContext(PreviewContext)
	return nil
}

func (gui *Gui) filterByTagPick(tag *models.Tag) error {
	results, err := gui.ruinCmd.Pick.Pick([]string{tag.Name}, false)
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

	if gui.state.Tags.CurrentTab == TagsTabInline {
		gui.updatePreviewPickResults(tag)
		return
	}

	gui.updatePreviewCardList(" Preview: #"+tag.Name+" ", func() ([]models.Note, error) {
		return gui.ruinCmd.Search.Search(tag.Name, gui.buildSearchOptions())
	})
}

func (gui *Gui) updatePreviewPickResults(tag *models.Tag) {
	results, err := gui.ruinCmd.Pick.Pick([]string{tag.Name}, false)
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
