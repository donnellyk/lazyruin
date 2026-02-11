package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) tagsPanel() *listPanel {
	return &listPanel{
		selectedIndex: &gui.state.Tags.SelectedIndex,
		itemCount:     func() int { return len(gui.state.Tags.Items) },
		render:        gui.renderTags,
		updatePreview: gui.updatePreviewForTags,
		context:       TagsContext,
	}
}

func (gui *Gui) tagsDown(g *gocui.Gui, v *gocui.View) error {
	return gui.tagsPanel().listDown(g, v)
}

func (gui *Gui) tagsUp(g *gocui.Gui, v *gocui.View) error {
	return gui.tagsPanel().listUp(g, v)
}

func (gui *Gui) tagsClick(g *gocui.Gui, v *gocui.View) error {
	idx := listClickIndex(v, 1)
	if idx >= 0 && idx < len(gui.state.Tags.Items) {
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
	if len(gui.state.Tags.Items) == 0 {
		return nil
	}

	tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]
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

func (gui *Gui) renameTag(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Tags.Items) == 0 {
		return nil
	}

	tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]

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
	if len(gui.state.Tags.Items) == 0 {
		return nil
	}

	tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]

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
	if len(gui.state.Tags.Items) == 0 {
		return
	}

	tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]
	gui.updatePreviewCardList(" Preview: #"+tag.Name+" ", func() ([]models.Note, error) {
		return gui.ruinCmd.Search.Search(tag.Name, gui.buildSearchOptions())
	})
}
