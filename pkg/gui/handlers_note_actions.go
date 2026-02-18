package gui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

// addGlobalTag opens the input popup to add a global tag.
func (gui *Gui) addGlobalTag(g *gocui.Gui, v *gocui.View) error {
	card := gui.helpers.Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	uuid := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Add Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{{Prefix: "#", Candidates: gui.tagCandidates}}
		},
		OnAccept: func(_ string, item *CompletionItem) error {
			tag := ""
			if item != nil {
				tag = item.Label
			}
			if tag == "" {
				return nil
			}
			err := gui.ruinCmd.Note.AddTag(uuid, tag)
			if err != nil {
				gui.showError(err)
				return nil
			}
			gui.helpers.Preview().ReloadContent()
			gui.refreshTags(false)
			return nil
		},
	})
	return nil
}

// removeTag opens the input popup showing only the current card's tags.
func (gui *Gui) removeTag(g *gocui.Gui, v *gocui.View) error {
	card := gui.helpers.Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	allTags := append(card.Tags, card.InlineTags...)
	if len(allTags) == 0 {
		return nil
	}
	uuid := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Remove Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{{Prefix: "#", Candidates: gui.currentCardTagCandidates}}
		},
		OnAccept: func(_ string, item *CompletionItem) error {
			tag := ""
			if item != nil {
				tag = item.Label
			}
			if tag == "" {
				return nil
			}
			err := gui.ruinCmd.Note.RemoveTag(uuid, tag)
			if err != nil {
				gui.showError(err)
				return nil
			}
			gui.helpers.Preview().ReloadContent()
			gui.refreshTags(false)
			return nil
		},
	})
	return nil
}

// setParentDialog opens the input popup with > / >> parent completion.
func (gui *Gui) setParentDialog(g *gocui.Gui, v *gocui.View) error {
	card := gui.helpers.Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	uuid := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Set Parent",
		Footer: " > bookmarks | >> all notes | / drill | Tab: accept | Esc: cancel ",
		Seed:   ">",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{
				{Prefix: ">", Candidates: gui.parentCandidatesFor(gui.state.InputPopupCompletion)},
			}
		},
		OnAccept: func(raw string, item *CompletionItem) error {
			parentRef := strings.TrimLeft(raw, ">")
			if item != nil {
				parentRef = item.Value
				if parentRef == "" {
					parentRef = item.Label
				}
			}
			if parentRef == "" {
				return nil
			}
			err := gui.ruinCmd.Note.SetParent(uuid, parentRef)
			if err != nil {
				gui.showError(err)
				return nil
			}
			gui.helpers.Preview().ReloadContent()
			return nil
		},
	})
	return nil
}

// removeParent removes the parent from the current card.
func (gui *Gui) removeParent(g *gocui.Gui, v *gocui.View) error {
	card := gui.helpers.Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	err := gui.ruinCmd.Note.RemoveParent(card.UUID)
	if err != nil {
		gui.showError(err)
		return nil
	}
	gui.helpers.Preview().ReloadContent()
	return nil
}

// toggleBookmark toggles a parent bookmark for the current card.
func (gui *Gui) toggleBookmark(g *gocui.Gui, v *gocui.View) error {
	card := gui.helpers.Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	// Check if a bookmark already exists for this note
	bookmarks, err := gui.ruinCmd.Parent.List()
	if err == nil {
		for _, bm := range bookmarks {
			if bm.UUID == card.UUID {
				// Remove existing bookmark
				gui.ruinCmd.Parent.Delete(bm.Name)
				gui.refreshParents(false)
				return nil
			}
		}
	}
	// No bookmark exists -- prompt for a name
	defaultName := card.Title
	if defaultName == "" {
		defaultName = card.UUID[:8]
	}
	cardUUID := card.UUID
	gui.openInputPopup(&InputPopupConfig{
		Title:  "Save Bookmark",
		Footer: " Enter: save | Esc: cancel ",
		Seed:   defaultName,
		OnAccept: func(raw string, _ *CompletionItem) error {
			if raw == "" {
				return nil
			}
			err := gui.ruinCmd.Parent.Save(raw, cardUUID)
			if err != nil {
				gui.showError(err)
				return nil
			}
			gui.refreshParents(false)
			return nil
		},
	})
	return nil
}
