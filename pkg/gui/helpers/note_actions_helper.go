package helpers

import (
	"strings"

	"kvnd/lazyruin/pkg/gui/types"
)

// NoteActionsHelper handles note-level mutations (tags, parents, bookmarks).
type NoteActionsHelper struct {
	c *HelperCommon
}

// NewNoteActionsHelper creates a new NoteActionsHelper.
func NewNoteActionsHelper(c *HelperCommon) *NoteActionsHelper {
	return &NoteActionsHelper{c: c}
}

// AddGlobalTag opens the input popup to add a global tag to the current preview card.
func (self *NoteActionsHelper) AddGlobalTag() error {
	gui := self.c.GuiCommon()
	card := gui.PreviewCurrentCard()
	if card == nil {
		return nil
	}
	uuid := card.UUID
	gui.OpenInputPopup(&types.InputPopupConfig{
		Title:  "Add Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{{Prefix: "#", Candidates: gui.TagCandidates}}
		},
		OnAccept: func(_ string, item *types.CompletionItem) error {
			tag := ""
			if item != nil {
				tag = item.Label
			}
			if tag == "" {
				return nil
			}
			err := self.c.RuinCmd().Note.AddTag(uuid, tag)
			if err != nil {
				gui.ShowError(err)
				return nil
			}
			gui.PreviewReloadContent()
			gui.RefreshTags(false)
			return nil
		},
	})
	return nil
}

// RemoveTag opens the input popup showing only the current card's tags for removal.
func (self *NoteActionsHelper) RemoveTag() error {
	gui := self.c.GuiCommon()
	card := gui.PreviewCurrentCard()
	if card == nil {
		return nil
	}
	allTags := append(card.Tags, card.InlineTags...)
	if len(allTags) == 0 {
		return nil
	}
	uuid := card.UUID
	gui.OpenInputPopup(&types.InputPopupConfig{
		Title:  "Remove Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{{Prefix: "#", Candidates: gui.CurrentCardTagCandidates}}
		},
		OnAccept: func(_ string, item *types.CompletionItem) error {
			tag := ""
			if item != nil {
				tag = item.Label
			}
			if tag == "" {
				return nil
			}
			err := self.c.RuinCmd().Note.RemoveTag(uuid, tag)
			if err != nil {
				gui.ShowError(err)
				return nil
			}
			gui.PreviewReloadContent()
			gui.RefreshTags(false)
			return nil
		},
	})
	return nil
}

// SetParentDialog opens the input popup with > / >> parent completion.
func (self *NoteActionsHelper) SetParentDialog() error {
	gui := self.c.GuiCommon()
	card := gui.PreviewCurrentCard()
	if card == nil {
		return nil
	}
	uuid := card.UUID
	gui.OpenInputPopup(&types.InputPopupConfig{
		Title:  "Set Parent",
		Footer: " > bookmarks | >> all notes | / drill | Tab: accept | Esc: cancel ",
		Seed:   ">",
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{
				{Prefix: ">", Candidates: gui.ParentCandidatesFor(gui.GetInputPopupCompletion())},
			}
		},
		OnAccept: func(raw string, item *types.CompletionItem) error {
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
			err := self.c.RuinCmd().Note.SetParent(uuid, parentRef)
			if err != nil {
				gui.ShowError(err)
				return nil
			}
			gui.PreviewReloadContent()
			return nil
		},
	})
	return nil
}

// RemoveParent removes the parent from the current card.
func (self *NoteActionsHelper) RemoveParent() error {
	gui := self.c.GuiCommon()
	card := gui.PreviewCurrentCard()
	if card == nil {
		return nil
	}
	err := self.c.RuinCmd().Note.RemoveParent(card.UUID)
	if err != nil {
		gui.ShowError(err)
		return nil
	}
	gui.PreviewReloadContent()
	return nil
}

// ToggleBookmark toggles a parent bookmark for the current card.
func (self *NoteActionsHelper) ToggleBookmark() error {
	gui := self.c.GuiCommon()
	card := gui.PreviewCurrentCard()
	if card == nil {
		return nil
	}
	// Check if a bookmark already exists for this note
	bookmarks, err := self.c.RuinCmd().Parent.List()
	if err == nil {
		for _, bm := range bookmarks {
			if bm.UUID == card.UUID {
				// Remove existing bookmark
				self.c.RuinCmd().Parent.Delete(bm.Name)
				gui.RefreshParents(false)
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
	gui.OpenInputPopup(&types.InputPopupConfig{
		Title:  "Save Bookmark",
		Footer: " Enter: save | Esc: cancel ",
		Seed:   defaultName,
		OnAccept: func(raw string, _ *types.CompletionItem) error {
			if raw == "" {
				return nil
			}
			err := self.c.RuinCmd().Parent.Save(raw, cardUUID)
			if err != nil {
				gui.ShowError(err)
				return nil
			}
			gui.RefreshParents(false)
			return nil
		},
	})
	return nil
}
