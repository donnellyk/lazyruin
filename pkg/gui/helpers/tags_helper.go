package helpers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"
)

// TagsHelper handles tag domain operations.
type TagsHelper struct {
	c *HelperCommon
}

// NewTagsHelper creates a new TagsHelper.
func NewTagsHelper(c *HelperCommon) *TagsHelper {
	return &TagsHelper{c: c}
}

// RefreshTags fetches all tags and re-renders the list.
// If preserve is true, the current selection is preserved by ID.
func (self *TagsHelper) RefreshTags(preserve bool) {
	gui := self.c.GuiCommon()
	tagsCtx := gui.Contexts().Tags
	prevID := tagsCtx.GetSelectedItemId()

	tags, err := self.c.RuinCmd().Tags.List()
	if err != nil {
		return
	}
	tagsCtx.Items = tags

	if preserve && prevID != "" {
		if newIdx := tagsCtx.GetList().FindIndexById(prevID); newIdx >= 0 {
			tagsCtx.SetSelectedLineIdx(newIdx)
		}
	} else {
		tagsCtx.SetSelectedLineIdx(0)
	}
	tagsCtx.ClampSelection()

	gui.RenderTags()
}

// CycleTagsTab cycles through All -> Global -> Inline tabs.
func (self *TagsHelper) CycleTagsTab() {
	gui := self.c.GuiCommon()
	tagsCtx := gui.Contexts().Tags
	idx := (tagsCtx.TabIndex() + 1) % len(context.TagsTabs)
	tagsCtx.CurrentTab = context.TagsTabs[idx]
	tagsCtx.SetSelectedLineIdx(0)
	gui.UpdateTagsTab()
	gui.RenderTags()
	self.UpdatePreviewForTags()
}

// SwitchTagsTabByIndex handles mouse clicks on tab headers.
func (self *TagsHelper) SwitchTagsTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(context.TagsTabs) {
		return nil
	}
	gui := self.c.GuiCommon()
	tagsCtx := gui.Contexts().Tags
	tagsCtx.CurrentTab = context.TagsTabs[tabIndex]
	tagsCtx.SetSelectedLineIdx(0)
	gui.UpdateTagsTab()
	gui.RenderTags()
	self.UpdatePreviewForTags()
	gui.PushContextByKey("tags")
	return nil
}

// FilterByTag dispatches to search or pick based on the tag scope.
func (self *TagsHelper) FilterByTag(tag *models.Tag) error {
	if tag == nil {
		return nil
	}
	tagsCtx := self.c.GuiCommon().Contexts().Tags
	if tagsCtx.CurrentTab == context.TagsTabInline {
		return self.FilterByTagPick(tag)
	}
	return self.FilterByTagSearch(tag)
}

// FilterByTagSearch searches for notes with the given tag.
func (self *TagsHelper) FilterByTagSearch(tag *models.Tag) error {
	gui := self.c.GuiCommon()
	opts := self.c.Helpers().Preview().BuildSearchOptions()
	notes, err := self.c.RuinCmd().Search.Search(tag.Name, opts)
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	self.c.Helpers().PreviewNav().PushNavHistory()
	self.c.Helpers().Preview().ShowCardList(" Tag: #"+tag.Name+" ", notes)
	gui.PushContextByKey("cardList")
	return nil
}

// FilterByTagPick runs a pick query for the given tag.
func (self *TagsHelper) FilterByTagPick(tag *models.Tag) error {
	gui := self.c.GuiCommon()
	results, err := self.c.RuinCmd().Pick.Pick([]string{tag.Name}, false, "")
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	// Store query so ReloadPickResults can re-run it after line edits.
	pickCtx := gui.Contexts().Pick
	pickCtx.Query = tag.Name
	pickCtx.AnyMode = false

	self.c.Helpers().Preview().ShowPickResults(" Pick: "+tag.Name+" ", results)
	gui.PushContextByKey("pickResults")
	return nil
}

// RenameTag prompts for a new name and renames the tag.
func (self *TagsHelper) RenameTag(tag *models.Tag) error {
	if tag == nil {
		return nil
	}
	gui := self.c.GuiCommon()

	gui.ShowInput("Rename Tag", "New name for #"+tag.Name+":", func(newName string) error {
		if newName == "" || newName == tag.Name {
			return nil
		}
		err := self.c.RuinCmd().Tags.Rename(tag.Name, newName)
		if err != nil {
			gui.ShowError(err)
			return nil
		}
		self.c.Helpers().Tags().RefreshTags(false)
		self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
		return nil
	})
	return nil
}

// DeleteTag shows confirmation and deletes the tag from all notes.
func (self *TagsHelper) DeleteTag(tag *models.Tag) error {
	if tag == nil {
		return nil
	}
	gui := self.c.GuiCommon()

	gui.ShowConfirm("Delete Tag", "Delete #"+tag.Name+" from all notes?", func() error {
		err := self.c.RuinCmd().Tags.Delete(tag.Name)
		if err != nil {
			gui.ShowError(err)
			return nil
		}
		self.c.Helpers().Tags().RefreshTags(false)
		self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
		return nil
	})
	return nil
}

// UpdatePreviewForTags updates the preview pane based on the selected tag.
func (self *TagsHelper) UpdatePreviewForTags() {
	gui := self.c.GuiCommon()
	tag := gui.Contexts().Tags.Selected()
	if tag == nil {
		return
	}

	tagsCtx := gui.Contexts().Tags
	if tagsCtx.CurrentTab == context.TagsTabInline {
		self.UpdatePreviewPickResults(tag)
		return
	}

	self.c.Helpers().Preview().UpdatePreviewCardList(" Tag: #"+tag.Name+" ", func() ([]models.Note, error) {
		return self.c.RuinCmd().Search.Search(tag.Name, self.c.Helpers().Preview().BuildSearchOptions())
	})
}

// UpdatePreviewPickResults runs a pick for the given tag and shows results in preview.
func (self *TagsHelper) UpdatePreviewPickResults(tag *models.Tag) {
	gui := self.c.GuiCommon()
	results, err := self.c.RuinCmd().Pick.Pick([]string{tag.Name}, false, "")
	if err != nil {
		return
	}

	// Store query so ReloadPickResults can re-run it after line edits.
	pickCtx := gui.Contexts().Pick
	pickCtx.Query = tag.Name
	pickCtx.AnyMode = false

	self.c.Helpers().Preview().ShowPickResults(" Pick: "+tag.Name+" ", results)
}
