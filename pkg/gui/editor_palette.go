package gui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

// paletteEditor intercepts Up/Down arrows for list navigation and
// delegates all other keys to SimpleEditor for text input. After
// every keystroke it re-filters the command list.
type paletteEditor struct {
	gui *Gui
}

func (e *paletteEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	state := e.gui.state.Palette
	if state == nil {
		return false
	}

	// Arrow keys and j/k for list navigation
	switch {
	case key == gocui.KeyArrowDown, key == 0 && ch == 'j' && mod == gocui.ModAlt:
		e.gui.paletteSelectMove(1)
		return true
	case key == gocui.KeyArrowUp, key == 0 && ch == 'k' && mod == gocui.ModAlt:
		e.gui.paletteSelectMove(-1)
		return true
	}

	// Delegate to SimpleEditor for text input
	handled := gocui.SimpleEditor(v, key, ch, mod)

	// Re-filter based on current text
	content := strings.TrimSpace(v.TextArea.GetContent())
	e.gui.filterPaletteCommands(content)
	e.gui.renderPaletteList()
	e.gui.scrollPaletteToSelection()

	return handled
}
