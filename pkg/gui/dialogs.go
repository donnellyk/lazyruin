package gui

import (
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// Dialog view names
const (
	ConfirmView = "confirm"
	InputView   = "input"
)

// DialogState tracks the current dialog state
type DialogState struct {
	Active        bool
	Type          string
	Title         string
	Message       string
	OnConfirm     func() error
	OnCancel      func()
	InputBuffer   string
	MenuItems     []types.MenuItem
	MenuSelection int
}

// centerRect computes centered coordinates for a dialog of the given size.
func centerRect(maxX, maxY, width, height int) (x0, y0, x1, y1 int) {
	x0 = (maxX - width) / 2
	y0 = (maxY - height) / 2
	return x0, y0, x0 + width, y0 + height
}

// showConfirm displays a confirmation dialog
func (gui *Gui) ShowConfirm(title, message string, onConfirm func() error) {
	gui.state.Dialog = &DialogState{
		Active:    true,
		Type:      "confirm",
		Title:     title,
		Message:   message,
		OnConfirm: onConfirm,
	}
}

// showInput displays a text input dialog
func (gui *Gui) ShowInput(title, message string, onConfirm func(input string) error) {
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
	items := gui.helpDialogItems()

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

// closeDialog closes any open dialog
func (gui *Gui) closeDialog() {
	if gui.state.Dialog != nil && gui.state.Dialog.OnCancel != nil {
		gui.state.Dialog.OnCancel()
	}
	gui.state.Dialog = nil
	gui.g.DeleteView(ConfirmView)
	gui.g.DeleteView(InputView)
	gui.g.DeleteView(MenuView)
	// Restore focus to the view for the current context
	gui.g.SetCurrentView(gui.contextToView(gui.contextMgr.Current()))
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

// menuEditor handles shortcut key presses in menu dialogs.
// Keybindings (j/k/arrows/enter/esc) are handled by gocui before reaching
// the Editor, so this only fires for unbound keys — exactly the shortcut
// keys displayed on menu items.
type menuEditor struct {
	gui *Gui
}

func (e *menuEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if e.gui.state.Dialog == nil {
		return false
	}

	// j/k navigation — the framework blocks view-specific char keybindings
	// on editable views, so handle them here.
	if ch == 'j' {
		e.gui.menuDown(e.gui.g, v)
		return true
	}
	if ch == 'k' {
		e.gui.menuUp(e.gui.g, v)
		return true
	}

	if ch == 0 {
		return false
	}
	pressed := string(ch)

	// KeepOpenKey: run the selected item's action without closing.
	sel := e.gui.state.Dialog.MenuSelection
	items := e.gui.state.Dialog.MenuItems
	if sel >= 0 && sel < len(items) {
		item := items[sel]
		if item.KeepOpenKey == pressed && item.OnRun != nil {
			item.OnRun()
			return true
		}
	}

	for _, item := range items {
		if item.Key == pressed && item.OnRun != nil {
			e.gui.closeDialog()
			item.OnRun()
			return true
		}
	}
	return false
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
	width := 50
	for _, item := range items {
		var l int
		if item.IsHeader {
			l = len(item.Label) + 10 // " --- Label --- "
			if item.Hint != "" {
				l += len(item.Hint) + 3 // "  (hint)"
			}
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
	v.Editable = true
	v.Editor = &menuEditor{gui: gui}
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
			header := fmt.Sprintf(" %s--- %s ---%s", AnsiCyan, item.Label, AnsiReset)
			if item.Hint != "" {
				visible := len(item.Label) + 10 // " --- Label --- "
				pad := innerWidth - visible - len(item.Hint)
				if pad < 2 {
					pad = 2
				}
				header += fmt.Sprintf("%s%s%s%s", strings.Repeat(" ", pad), AnsiDim, item.Hint, AnsiReset)
			}
			fmt.Fprintln(v, header)
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
