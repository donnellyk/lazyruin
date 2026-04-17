package gui

import (
	"time"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/models"
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
	noteCopy := *note
	_ = gui.helpers.Navigator().NavigateTo("cardList", noteCopy.Title, func() error {
		source := gui.helpers.Preview().NewSingleNoteSource(noteCopy.UUID)
		gui.helpers.Preview().ShowCardList(noteCopy.Title, []models.Note{noteCopy}, source)
		return nil
	})
}

func (gui *Gui) openParentBookmark(parent *models.ParentBookmark) {
	parentCopy := *parent
	composed, sourceMap, err := gui.ruinCmd.Parent.Compose(parentCopy)
	if err != nil {
		gui.helpers.DatePreview().LoadDatePreview(time.Now().Format("2006-01-02"))
		return
	}
	_ = gui.helpers.Navigator().NavigateTo("compose", "Parent: "+parentCopy.Name, func() error {
		gui.helpers.Preview().ShowCompose("Parent: "+parentCopy.Name, composed, sourceMap, parentCopy)
		gui.contexts.Compose.Requery = func() (models.Note, []models.SourceMapEntry, error) {
			return gui.ruinCmd.Parent.Compose(parentCopy)
		}
		return nil
	})
}
