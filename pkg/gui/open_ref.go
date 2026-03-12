package gui

import (
	"time"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/models"
)

// openInitialRef resolves the --open reference and shows it in the preview.
// Resolution order: parent bookmark name, then note by path, then note by title.
func (gui *Gui) openInitialRef(ref string) {
	// 1. Try parent bookmark
	parents, err := gui.ruinCmd.Parent.List()
	if err == nil {
		for _, p := range parents {
			if p.Name == ref {
				gui.openParentBookmark(&p)
				return
			}
		}
	}

	opts := commands.SearchOptions{IncludeContent: true, StripTitle: true, StripGlobalTags: true}

	// 2. Try note by path
	note, err := gui.ruinCmd.Search.GetByPath(ref, opts)
	if err == nil && note != nil {
		gui.openNote(note)
		return
	}

	// 3. Try note by title
	note, err = gui.ruinCmd.Search.GetByTitle(ref, opts)
	if err == nil && note != nil {
		gui.openNote(note)
		return
	}

	// 4. Fallback to today view
	gui.helpers.DatePreview().LoadDatePreview(time.Now().Format("2006-01-02"))
}

func (gui *Gui) openNote(note *models.Note) {
	gui.helpers.Preview().ShowCardList(note.Title, []models.Note{*note})
	gui.pushContextByKey("cardList")
}

func (gui *Gui) openParentBookmark(parent *models.ParentBookmark) {
	composed, sourceMap, err := gui.ruinCmd.Parent.ComposeFlat(parent.UUID, parent.Title)
	if err != nil {
		gui.helpers.DatePreview().LoadDatePreview(time.Now().Format("2006-01-02"))
		return
	}
	gui.helpers.Preview().ShowCompose("Parent: "+parent.Name, composed, sourceMap, parent.UUID, parent.Title)
	gui.pushContextByKey("compose")
}
