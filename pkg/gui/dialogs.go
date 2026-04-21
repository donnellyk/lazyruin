package gui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/donnellyk/lazyruin/pkg/gui/helpers"
	"github.com/donnellyk/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// Dialog view names
const (
	ConfirmView = "confirm"
	InputView   = "input"
	AboutView   = "about"
)

// ruinLogo is the RUIN wordmark rendered in the about splash and the
// onboarding welcome dialog.
var ruinLogo = []string{
	"██████╗  ██╗   ██╗ ██╗ ███╗   ██╗",
	"██╔══██╗ ██║   ██║ ██║ ████╗  ██║",
	"██████╔╝ ██║   ██║ ██║ ██╔██╗ ██║",
	"██╔══██╗ ██║   ██║ ██║ ██║╚██╗██║",
	"██║  ██║ ╚██████╔╝ ██║ ██║ ╚████║",
	"╚═╝  ╚═╝  ╚═════╝  ╚═╝ ╚═╝  ╚═══╝",
}

// ruinEpigraph is the Eliot quotation shown beneath the logo on the about
// splash (but not in the onboarding dialog).
var ruinEpigraph = []string{
	"\"These fragments I have shored",
	"against my ruin\"",
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
	MenuItems     []types.MenuItem
	MenuSelection int
	// Hero, when true, renders the RUIN banner above the confirm dialog's
	// message. Used by the first-run onboarding prompt.
	Hero bool
	// Footer, when set, overrides the default "[y] Yes · [n/Esc] No"
	// footer. Used by the init prompt to show "[y] Yes · [q/Esc] Quit".
	Footer string
	// QuitOnCancel, when true, causes the cancel action (n/q/Esc) to
	// terminate the TUI rather than simply closing the dialog. The
	// ExitError from GuiState is surfaced to the caller of Gui.Run.
	QuitOnCancel bool
	// lastScrolledSelection records the MenuSelection that the menu view
	// was last auto-scrolled to. The layout skips scroll-to-selection when
	// it matches, so mouse-wheel scrolling isn't undone on the next redraw.
	lastScrolledSelection int
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

// ShowHeroConfirm displays a confirmation dialog with the RUIN banner
// rendered above the message. Used for the first-run onboarding prompt.
func (gui *Gui) ShowHeroConfirm(title, message string, onConfirm func() error) {
	gui.state.Dialog = &DialogState{
		Active:    true,
		Type:      "confirm",
		Title:     title,
		Message:   message,
		OnConfirm: onConfirm,
		Hero:      true,
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

// ShowAbout displays the about splash screen.
func (gui *Gui) ShowAbout() {
	gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "about",
		Title:  "About",
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
	gui.g.DeleteView(AboutView)
	// Restore focus to the view for the current context
	gui.g.SetCurrentView(gui.contextToView(gui.contextMgr.Current()))
}

// createConfirmDialog renders the confirmation dialog
func (gui *Gui) createConfirmDialog(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || gui.state.Dialog.Type != "confirm" {
		return nil
	}

	width, height := 50, 5
	if gui.state.Dialog.Hero {
		msgLines := strings.Count(gui.state.Dialog.Message, "\n") + 1
		// top blank + logo + blank + message + frame
		height = 1 + len(ruinLogo) + 1 + msgLines + 2
	}
	x0, y0, x1, y1 := centerRect(maxX, maxY, width, height)

	v, err := g.SetView(ConfirmView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Title = " " + gui.state.Dialog.Title + " "
	if gui.state.Dialog.Footer != "" {
		v.Footer = gui.state.Dialog.Footer
	} else {
		v.Footer = " [y] Yes · [n/Esc] No "
	}
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen
	v.Clear()

	if gui.state.Dialog.Hero {
		innerW, _ := v.InnerSize()
		fmt.Fprintln(v, "")
		for _, line := range ruinLogo {
			fmt.Fprintln(v, centerLine(line, innerW))
		}
		fmt.Fprintln(v, "")
		for _, line := range strings.Split(gui.state.Dialog.Message, "\n") {
			fmt.Fprintln(v, centerLine(line, innerW))
		}
	} else {
		fmt.Fprintln(v, "")
		fmt.Fprintln(v, "  "+gui.state.Dialog.Message)
	}

	g.SetViewOnTop(ConfirmView)
	g.SetCurrentView(ConfirmView)

	return nil
}

// centerLine pads s with leading spaces so its rune content is centered
// within width cells. Lines longer than width are returned unchanged.
func centerLine(s string, width int) string {
	pad := (width - utf8.RuneCountInString(s)) / 2
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + s
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
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen
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
	height := min(
		// border
		len(items)+2, maxY-4)
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
				pad := max(innerWidth-visible-len(item.Hint), 2)
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

	// Scroll to keep selection visible, but only when selection has moved
	// since the last render. The layout runs every tick; re-scrolling on
	// every tick would undo mouse-wheel scrolling.
	if gui.state.Dialog.MenuSelection != gui.state.Dialog.lastScrolledSelection {
		viewHeight := height - 2
		scrollListView(v, gui.state.Dialog.MenuSelection, 1, viewHeight)
		gui.state.Dialog.lastScrolledSelection = gui.state.Dialog.MenuSelection
	}

	g.SetViewOnTop(MenuView)
	g.SetCurrentView(MenuView)

	return nil
}

// createAboutDialog renders the ASCII art about splash along with the
// resolved vault path, its source, and any non-default lazyruin config.
func (gui *Gui) createAboutDialog(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || gui.state.Dialog.Type != "about" {
		return nil
	}

	infoLines := gui.aboutInfoLines()

	width := 60
	// leading blank + logo + blank + info (+ trailing blank if any) + epigraph + frame
	height := 1 + len(ruinLogo) + 1 + len(infoLines) + len(ruinEpigraph) + 2
	if len(infoLines) > 0 {
		height++ // blank between info block and epigraph
	}
	x0, y0, x1, y1 := centerRect(maxX, maxY, width, height)

	v, err := g.SetView(AboutView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen
	v.Clear()

	innerW, _ := v.InnerSize()
	fmt.Fprintln(v, "")
	for _, line := range ruinLogo {
		fmt.Fprintln(v, centerLine(line, innerW))
	}
	fmt.Fprintln(v, "")
	for _, line := range infoLines {
		fmt.Fprintln(v, centerLine(line, innerW))
	}
	if len(infoLines) > 0 {
		fmt.Fprintln(v, "")
	}
	for _, line := range ruinEpigraph {
		fmt.Fprintln(v, centerLine(line, innerW))
	}

	g.SetViewOnTop(AboutView)
	g.SetCurrentView(AboutView)
	return nil
}

// aboutInfoLines builds the vault/config block shown in the about dialog.
// Returns an empty slice when there is nothing to show.
func (gui *Gui) aboutInfoLines() []string {
	var lines []string
	if gui.ruinCmd != nil {
		if p := gui.ruinCmd.VaultPath(); p != "" {
			lines = append(lines, "Vault: "+p)
		}
	}
	if gui.VaultSource != "" {
		lines = append(lines, "Source: "+gui.VaultSource)
	}
	if gui.config != nil {
		var cfgLines []string
		if gui.config.Editor != "" {
			cfgLines = append(cfgLines, "editor: "+gui.config.Editor)
		}
		if gui.config.ChromaTheme != "" {
			cfgLines = append(cfgLines, "chroma_theme: "+gui.config.ChromaTheme)
		}
		if gui.config.ViewOptions.HideDone {
			cfgLines = append(cfgLines, "hide_done: true")
		}
		if len(cfgLines) > 0 {
			if len(lines) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, cfgLines...)
		}
	}
	return lines
}

// renderDialogs renders any active dialog
func (gui *Gui) renderDialogs(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || !gui.state.Dialog.Active {
		g.DeleteView(ConfirmView)
		g.DeleteView(InputView)
		g.DeleteView(MenuView)
		g.DeleteView(AboutView)
		return nil
	}

	switch gui.state.Dialog.Type {
	case "confirm":
		return gui.createConfirmDialog(g, maxX, maxY)
	case "input":
		return gui.createInputDialog(g, maxX, maxY)
	case "menu":
		return gui.createMenuDialog(g, maxX, maxY)
	case "about":
		return gui.createAboutDialog(g, maxX, maxY)
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
	if err := gui.g.SetKeybinding(ConfirmView, 'q', gocui.ModNone, gui.confirmNo); err != nil {
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

	// About dialog
	if err := gui.g.SetKeybinding(AboutView, gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		gui.closeDialog()
		return nil
	}); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(AboutView, gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		gui.closeDialog()
		return nil
	}); err != nil {
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
	// Mouse wheel scrolling for long menus (e.g. the Keybindings help dialog).
	// Scrolls the view origin without moving selection — same model as side
	// panels. menuDown/menuUp move selection and will re-center the origin
	// on the next j/k press.
	if err := gui.g.SetKeybinding(MenuView, gocui.MouseWheelDown, gocui.ModNone, gui.menuWheelDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(MenuView, gocui.MouseWheelUp, gocui.ModNone, gui.menuWheelUp); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) menuWheelDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		helpers.ScrollViewport(v, 3)
	}
	return nil
}

func (gui *Gui) menuWheelUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		helpers.ScrollViewport(v, -3)
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
	quit := gui.state.Dialog != nil && gui.state.Dialog.QuitOnCancel
	gui.closeDialog()
	if quit {
		return gocui.ErrQuit
	}
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
