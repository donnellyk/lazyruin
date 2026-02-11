package gui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

// Dialog view names
const (
	ConfirmView = "confirm"
	InputView   = "input"
)

// MenuItem represents a single item in a menu dialog
type MenuItem struct {
	Label    string
	Key      string // shortcut key hint (e.g. "j", "k")
	OnRun    func() error
	IsHeader bool
}

// DialogState tracks the current dialog state
type DialogState struct {
	Active        bool
	Type          string
	Title         string
	Message       string
	OnConfirm     func() error
	OnCancel      func()
	InputBuffer   string
	MenuItems     []MenuItem
	MenuSelection int
}

// centerRect computes centered coordinates for a dialog of the given size.
func centerRect(maxX, maxY, width, height int) (x0, y0, x1, y1 int) {
	x0 = (maxX - width) / 2
	y0 = (maxY - height) / 2
	return x0, y0, x0 + width, y0 + height
}

// showConfirm displays a confirmation dialog
func (gui *Gui) showConfirm(title, message string, onConfirm func() error) {
	gui.state.Dialog = &DialogState{
		Active:    true,
		Type:      "confirm",
		Title:     title,
		Message:   message,
		OnConfirm: onConfirm,
	}
}

// showInput displays a text input dialog
func (gui *Gui) showInput(title, message string, onConfirm func(input string) error) {
	gui.state.Dialog = &DialogState{
		Active:  true,
		Type:    "input",
		Title:   title,
		Message: message,
		OnConfirm: func() error {
			return onConfirm(gui.state.Dialog.InputBuffer)
		},
	}
}

// showHelp displays a context-sensitive keybindings menu
func (gui *Gui) showHelp() {
	var items []MenuItem

	// Context-specific section from shared hint definitions
	def := gui.contextHintDefs()
	if def.header != "" {
		items = append(items, MenuItem{Label: def.header, IsHeader: true})
	}
	for _, h := range def.hints {
		items = append(items, MenuItem{Key: h.key, Label: h.action})
	}

	// Blank separator
	items = append(items, MenuItem{})

	// Global section
	items = append(items, MenuItem{Label: "Global", IsHeader: true})
	for _, h := range globalHints() {
		items = append(items, MenuItem{Key: h.key, Label: h.action})
	}

	// Navigation section
	navHints := gui.navigationHints()
	if len(navHints) > 0 {
		items = append(items, MenuItem{}) // blank separator
		items = append(items, MenuItem{Label: "Navigation", IsHeader: true})
		for _, h := range navHints {
			items = append(items, MenuItem{Key: h.key, Label: h.action})
		}
	}

	// Find first non-header item for initial selection
	initialSel := 0
	for i, item := range items {
		if !item.IsHeader && item.Label != "" {
			initialSel = i
			break
		}
	}

	gui.state.Dialog = &DialogState{
		Active:        true,
		Type:          "menu",
		Title:         "Keybindings",
		MenuItems:     items,
		MenuSelection: initialSel,
	}
}

// showMergeOverlay displays the merge direction chooser as a menu
func (gui *Gui) showMergeOverlay() {
	gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "menu",
		Title:  "Merge",
		MenuItems: []MenuItem{
			{Label: "Merge with note above", Key: "u", OnRun: func() error { return gui.executeMerge("up") }},
			{Label: "Merge with note below", Key: "d", OnRun: func() error { return gui.executeMerge("down") }},
		},
		MenuSelection: 0,
	}
}

// closeDialog closes any open dialog
func (gui *Gui) closeDialog() {
	if gui.state.Dialog != nil && gui.state.Dialog.OnCancel != nil {
		gui.state.Dialog.OnCancel()
	}
	gui.state.Dialog = nil
	gui.g.DeleteView(ConfirmView)
	gui.g.DeleteView(InputView)
	gui.g.DeleteView(MenuView)
}

// createConfirmDialog renders the confirmation dialog
func (gui *Gui) createConfirmDialog(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || gui.state.Dialog.Type != "confirm" {
		return nil
	}

	x0, y0, x1, y1 := centerRect(maxX, maxY, 50, 7)

	v, err := g.SetView(ConfirmView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Title = " " + gui.state.Dialog.Title + " "
	setRoundedCorners(v)
	v.Clear()

	fmt.Fprintln(v, "")
	fmt.Fprintln(v, "  "+gui.state.Dialog.Message)
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, "  [y] Yes    [n/Esc] No")

	g.SetViewOnTop(ConfirmView)
	g.SetCurrentView(ConfirmView)

	return nil
}

// createInputDialog renders the input dialog
func (gui *Gui) createInputDialog(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || gui.state.Dialog.Type != "input" {
		return nil
	}

	x0, y0, x1, y1 := centerRect(maxX, maxY, 50, 7)

	v, err := g.SetView(InputView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Title = " " + gui.state.Dialog.Title + " "
	v.Editable = true
	setRoundedCorners(v)
	v.Clear()

	fmt.Fprintln(v, "")
	fmt.Fprintln(v, "  "+gui.state.Dialog.Message)
	fmt.Fprintln(v, "")

	g.SetViewOnTop(InputView)
	g.SetCurrentView(InputView)
	g.Cursor = true

	return nil
}

// createMenuDialog renders a navigable menu list overlay
func (gui *Gui) createMenuDialog(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || gui.state.Dialog.Type != "menu" {
		return nil
	}

	items := gui.state.Dialog.MenuItems

	// Check if any item is actionable (has OnRun)
	hasActions := false
	for _, item := range items {
		if item.OnRun != nil {
			hasActions = true
			break
		}
	}

	// Find max key width for column alignment
	maxKeyLen := 0
	for _, item := range items {
		if !item.IsHeader && len(item.Key) > maxKeyLen {
			maxKeyLen = len(item.Key)
		}
	}

	// Size: width based on longest item, height based on item count
	width := 30
	for _, item := range items {
		var l int
		if item.IsHeader {
			l = len(item.Label) + 10 // " --- Label --- "
		} else if item.Key != "" {
			l = 1 + maxKeyLen + 2 + len(item.Label) + 1 // " key  label "
		} else {
			l = 1 + maxKeyLen + 2 + len(item.Label) + 1
		}
		if l > width {
			width = l
		}
	}
	if width > maxX-4 {
		width = maxX - 4
	}
	height := len(items) + 2 // border
	if height > maxY-4 {
		height = maxY - 4
	}
	x0, y0, x1, y1 := centerRect(maxX, maxY, width, height)

	v, err := g.SetView(MenuView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Title = " " + gui.state.Dialog.Title + " "
	if hasActions {
		v.Footer = fmt.Sprintf("%d of %d", gui.state.Dialog.MenuSelection+1, len(items))
	} else {
		v.Footer = ""
	}
	v.Highlight = false
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen
	v.Clear()

	innerWidth, _ := v.InnerSize()
	if innerWidth < 10 {
		innerWidth = width - 2
	}

	for i, item := range items {
		// Header items
		if item.IsHeader {
			fmt.Fprintf(v, " %s--- %s ---%s\n", AnsiCyan, item.Label, AnsiReset)
			continue
		}

		selected := i == gui.state.Dialog.MenuSelection

		// Build column-aligned line: " key  label"
		keyPad := maxKeyLen - len(item.Key)
		var line string
		if item.Key != "" {
			line = fmt.Sprintf(" %s%s  %s", item.Key, strings.Repeat(" ", keyPad), item.Label)
		} else {
			line = fmt.Sprintf(" %s  %s", strings.Repeat(" ", maxKeyLen), item.Label)
		}

		if selected {
			pad := innerWidth - len([]rune(line))
			if pad > 0 {
				line += strings.Repeat(" ", pad)
			}
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line, AnsiReset)
		} else {
			if item.Key != "" {
				fmt.Fprintf(v, " %s%-*s%s  %s\n", AnsiGreen, maxKeyLen, item.Key, AnsiReset, item.Label)
			} else {
				fmt.Fprintf(v, " %s  %s\n", strings.Repeat(" ", maxKeyLen), item.Label)
			}
		}
	}

	// Scroll to keep selection visible
	viewHeight := height - 2
	scrollListView(v, gui.state.Dialog.MenuSelection, 1, viewHeight)

	g.SetViewOnTop(MenuView)
	g.SetCurrentView(MenuView)

	return nil
}

// renderDialogs renders any active dialog
func (gui *Gui) renderDialogs(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || !gui.state.Dialog.Active {
		g.DeleteView(ConfirmView)
		g.DeleteView(InputView)
		g.DeleteView(MenuView)
		return nil
	}

	switch gui.state.Dialog.Type {
	case "confirm":
		return gui.createConfirmDialog(g, maxX, maxY)
	case "input":
		return gui.createInputDialog(g, maxX, maxY)
	case "menu":
		return gui.createMenuDialog(g, maxX, maxY)
	}

	return nil
}

// setupDialogKeybindings sets up keybindings for dialogs
func (gui *Gui) setupDialogKeybindings() error {
	// Confirm dialog
	if err := gui.g.SetKeybinding(ConfirmView, 'y', gocui.ModNone, gui.confirmYes); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(ConfirmView, 'n', gocui.ModNone, gui.confirmNo); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(ConfirmView, gocui.KeyEsc, gocui.ModNone, gui.confirmNo); err != nil {
		return err
	}

	// Input dialog
	if err := gui.g.SetKeybinding(InputView, gocui.KeyEnter, gocui.ModNone, gui.inputConfirm); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(InputView, gocui.KeyEsc, gocui.ModNone, gui.inputCancel); err != nil {
		return err
	}

	// Menu dialog
	if err := gui.g.SetKeybinding(MenuView, 'j', gocui.ModNone, gui.menuDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(MenuView, 'k', gocui.ModNone, gui.menuUp); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(MenuView, gocui.KeyArrowDown, gocui.ModNone, gui.menuDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(MenuView, gocui.KeyArrowUp, gocui.ModNone, gui.menuUp); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(MenuView, gocui.KeyEnter, gocui.ModNone, gui.menuConfirm); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(MenuView, gocui.KeyEsc, gocui.ModNone, gui.menuCancel); err != nil {
		return err
	}

	return nil
}

// Dialog handlers

func (gui *Gui) confirmYes(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Dialog != nil && gui.state.Dialog.OnConfirm != nil {
		err := gui.state.Dialog.OnConfirm()
		gui.closeDialog()
		return err
	}
	gui.closeDialog()
	return nil
}

func (gui *Gui) confirmNo(g *gocui.Gui, v *gocui.View) error {
	gui.closeDialog()
	return nil
}

func (gui *Gui) inputConfirm(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Dialog != nil && gui.state.Dialog.OnConfirm != nil {
		gui.state.Dialog.InputBuffer = strings.TrimSpace(v.Buffer())
		err := gui.state.Dialog.OnConfirm()
		gui.closeDialog()
		g.Cursor = false
		return err
	}
	gui.closeDialog()
	g.Cursor = false
	return nil
}

func (gui *Gui) inputCancel(g *gocui.Gui, v *gocui.View) error {
	gui.closeDialog()
	g.Cursor = false
	return nil
}

func (gui *Gui) menuDown(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Dialog == nil {
		return nil
	}
	items := gui.state.Dialog.MenuItems
	for _, item := range items {
		if item.Key == "j" && item.OnRun != nil {
			gui.closeDialog()
			return item.OnRun()
		}
	}
	// Find next non-header item
	sel := gui.state.Dialog.MenuSelection + 1
	for sel < len(items) && items[sel].IsHeader {
		sel++
	}
	if sel < len(items) {
		gui.state.Dialog.MenuSelection = sel
	}
	return nil
}

func (gui *Gui) menuUp(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Dialog == nil {
		return nil
	}
	items := gui.state.Dialog.MenuItems
	for _, item := range items {
		if item.Key == "k" && item.OnRun != nil {
			gui.closeDialog()
			return item.OnRun()
		}
	}
	// Find previous non-header item
	sel := gui.state.Dialog.MenuSelection - 1
	for sel >= 0 && items[sel].IsHeader {
		sel--
	}
	if sel >= 0 {
		gui.state.Dialog.MenuSelection = sel
	}
	return nil
}

func (gui *Gui) menuConfirm(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Dialog == nil {
		return nil
	}
	item := gui.state.Dialog.MenuItems[gui.state.Dialog.MenuSelection]
	gui.closeDialog()
	if item.OnRun != nil {
		return item.OnRun()
	}
	return nil
}

func (gui *Gui) menuCancel(g *gocui.Gui, v *gocui.View) error {
	gui.closeDialog()
	return nil
}
