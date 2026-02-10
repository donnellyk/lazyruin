package gui

import (
	"os"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) deleteCardFromPreview(g *gocui.Gui, v *gocui.View) error {
	if !gui.state.Preview.EditMode {
		return nil
	}
	if len(gui.state.Preview.Cards) == 0 {
		return nil
	}

	card := gui.state.Preview.Cards[gui.state.Preview.SelectedCardIndex]
	title := card.Title
	if title == "" {
		title = card.Path
	}
	if len(title) > 30 {
		title = title[:30] + "..."
	}

	gui.showConfirm("Delete Note", "Delete \""+title+"\"?", func() error {
		err := os.Remove(card.Path)
		if err != nil {
			return nil
		}
		idx := gui.state.Preview.SelectedCardIndex
		gui.state.Preview.Cards = append(gui.state.Preview.Cards[:idx], gui.state.Preview.Cards[idx+1:]...)
		if gui.state.Preview.SelectedCardIndex >= len(gui.state.Preview.Cards) && gui.state.Preview.SelectedCardIndex > 0 {
			gui.state.Preview.SelectedCardIndex--
		}
		gui.refreshNotes(false)
		gui.renderPreview()
		return nil
	})
	return nil
}

func (gui *Gui) moveCardHandler(g *gocui.Gui, v *gocui.View) error {
	if !gui.state.Preview.EditMode {
		return nil
	}
	if len(gui.state.Preview.Cards) <= 1 {
		return nil
	}
	gui.showMoveOverlay()
	return nil
}

func (gui *Gui) showMoveOverlay() {
	gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "menu",
		Title:  "Move",
		MenuItems: []MenuItem{
			{Label: "Move card up", Key: "u", OnRun: func() error { return gui.moveCard("up") }},
			{Label: "Move card down", Key: "d", OnRun: func() error { return gui.moveCard("down") }},
		},
		MenuSelection: 0,
	}
}

func (gui *Gui) moveCard(direction string) error {
	idx := gui.state.Preview.SelectedCardIndex
	if direction == "up" {
		if idx <= 0 {
			return nil
		}
		gui.state.Preview.Cards[idx], gui.state.Preview.Cards[idx-1] = gui.state.Preview.Cards[idx-1], gui.state.Preview.Cards[idx]
		gui.state.Preview.SelectedCardIndex--
	} else {
		if idx >= len(gui.state.Preview.Cards)-1 {
			return nil
		}
		gui.state.Preview.Cards[idx], gui.state.Preview.Cards[idx+1] = gui.state.Preview.Cards[idx+1], gui.state.Preview.Cards[idx]
		gui.state.Preview.SelectedCardIndex++
	}
	gui.renderPreview()
	return nil
}

func (gui *Gui) mergeCardHandler(g *gocui.Gui, v *gocui.View) error {
	if !gui.state.Preview.EditMode {
		return nil
	}
	if len(gui.state.Preview.Cards) <= 1 {
		return nil
	}
	gui.showMergeOverlay()
	return nil
}

func (gui *Gui) executeMerge(direction string) error {
	idx := gui.state.Preview.SelectedCardIndex
	var targetIdx, sourceIdx int
	if direction == "down" {
		if idx >= len(gui.state.Preview.Cards)-1 {
			return nil
		}
		targetIdx = idx
		sourceIdx = idx + 1
	} else {
		if idx <= 0 {
			return nil
		}
		targetIdx = idx
		sourceIdx = idx - 1
	}

	target := gui.state.Preview.Cards[targetIdx]
	source := gui.state.Preview.Cards[sourceIdx]

	// Read both files' raw content (after stripping frontmatter)
	targetContent, err := gui.loadNoteContent(target.Path)
	if err != nil {
		return nil
	}
	sourceContent, err := gui.loadNoteContent(source.Path)
	if err != nil {
		return nil
	}

	// Merge tags (union)
	tagSet := make(map[string]bool)
	for _, t := range target.Tags {
		tagSet[t] = true
	}
	for _, t := range source.Tags {
		tagSet[t] = true
	}
	var mergedTags []string
	for t := range tagSet {
		mergedTags = append(mergedTags, t)
	}

	// Combine content
	combined := strings.TrimRight(targetContent, "\n") + "\n\n" + strings.TrimRight(sourceContent, "\n") + "\n"

	// Rewrite target file
	err = gui.writeNoteFile(target.Path, combined, mergedTags)
	if err != nil {
		return nil
	}

	// Delete source file
	os.Remove(source.Path)

	// Remove source from cards
	gui.state.Preview.Cards = append(gui.state.Preview.Cards[:sourceIdx], gui.state.Preview.Cards[sourceIdx+1:]...)
	if gui.state.Preview.SelectedCardIndex >= len(gui.state.Preview.Cards) {
		gui.state.Preview.SelectedCardIndex = len(gui.state.Preview.Cards) - 1
	}
	if gui.state.Preview.SelectedCardIndex < 0 {
		gui.state.Preview.SelectedCardIndex = 0
	}

	gui.refreshNotes(false)
	gui.renderPreview()
	return nil
}

// writeNoteFile rewrites a note file preserving uuid/created/updated, with merged tags and new content.
func (gui *Gui) writeNoteFile(path, content string, tags []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Extract existing frontmatter fields
	raw := string(data)
	uuid := ""
	created := ""
	updated := ""
	title := ""

	if strings.HasPrefix(raw, "---") {
		rest := raw[3:]
		if idx := strings.Index(rest, "\n---"); idx != -1 {
			fmBlock := rest[:idx]
			for _, line := range strings.Split(fmBlock, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "uuid:") {
					uuid = strings.TrimSpace(strings.TrimPrefix(line, "uuid:"))
				} else if strings.HasPrefix(line, "created:") {
					created = strings.TrimSpace(strings.TrimPrefix(line, "created:"))
				} else if strings.HasPrefix(line, "updated:") {
					updated = strings.TrimSpace(strings.TrimPrefix(line, "updated:"))
				} else if strings.HasPrefix(line, "title:") {
					title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
				}
			}
		}
	}

	// Build new frontmatter
	var fm strings.Builder
	fm.WriteString("---\n")
	if uuid != "" {
		fm.WriteString("uuid: " + uuid + "\n")
	}
	if created != "" {
		fm.WriteString("created: " + created + "\n")
	}
	if updated != "" {
		fm.WriteString("updated: " + updated + "\n")
	}
	if title != "" {
		fm.WriteString("title: " + title + "\n")
	}
	if len(tags) > 0 {
		fm.WriteString("tags:\n")
		for _, t := range tags {
			fm.WriteString("  - " + t + "\n")
		}
	} else {
		fm.WriteString("tags: []\n")
	}
	fm.WriteString("---\n")

	return os.WriteFile(path, []byte(fm.String()+content), 0644)
}
