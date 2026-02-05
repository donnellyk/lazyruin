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
	HelpView    = "help"
)

// DialogState tracks the current dialog state
type DialogState struct {
	Active      bool
	Type        string
	Title       string
	Message     string
	OnConfirm   func() error
	OnCancel    func()
	InputBuffer string
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

// showHelp displays the help overlay
func (gui *Gui) showHelp() {
	gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "help",
		Title:  "Keybindings",
	}
}

// showMergeOverlay displays the merge direction chooser
func (gui *Gui) showMergeOverlay() {
	gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "merge",
		Title:  "Merge",
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
	gui.g.DeleteView(HelpView)
	gui.g.DeleteView(MergeView)
}

// createConfirmDialog renders the confirmation dialog
func (gui *Gui) createConfirmDialog(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || gui.state.Dialog.Type != "confirm" {
		return nil
	}

	width := 50
	height := 7
	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

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

	width := 50
	height := 7
	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

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

// createHelpDialog renders the help overlay
func (gui *Gui) createHelpDialog(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || gui.state.Dialog.Type != "help" {
		return nil
	}

	width := 60
	height := 25
	if height > maxY-4 {
		height = maxY - 4
	}
	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

	v, err := g.SetView(HelpView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Title = " Keybindings "
	setRoundedCorners(v)
	v.Clear()

	help := `
  Global
  ──────────────────────────────────────────
  q / Ctrl+C     Quit
  Tab            Next panel
  1 / 2 / 3      Focus Notes / Queries / Tags
  p              Focus Preview
  /              Search
  Ctrl+R         Refresh all
  ?              Show this help

  Notes Panel
  ──────────────────────────────────────────
  1              Cycle tabs (All/Today/Recent)
  j / k          Move down / up
  g / G          Go to top / bottom
  Enter / e      Edit note in $EDITOR
  E              Enter edit mode (bulk ops)
  n              New note
  d              Delete note
  y              Copy note path

  Tags Panel
  ──────────────────────────────────────────
  j / k          Move down / up
  Enter          Filter notes by tag
  r              Rename tag
  d              Delete tag

  Queries Panel
  ──────────────────────────────────────────
  j / k          Move down / up
  Enter          Run query
  d              Delete query

  Preview Panel
  ──────────────────────────────────────────
  j / k          Scroll down / up
  Enter          Focus selected card
  Esc            Back to previous panel
  f              Toggle frontmatter
  t              Toggle title
  T              Toggle global tags

  Edit Mode (Preview)
  ──────────────────────────────────────────
  d              Delete card
  J / K          Move card down / up
  m              Merge card (j=down, k=up)
  Esc            Exit edit mode

  Press any key to close
`
	fmt.Fprint(v, help)

	g.SetViewOnTop(HelpView)
	g.SetCurrentView(HelpView)

	return nil
}

// createMergeDialog renders the merge direction chooser overlay
func (gui *Gui) createMergeDialog(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || gui.state.Dialog.Type != "merge" {
		return nil
	}

	width := 44
	height := 5
	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

	v, err := g.SetView(MergeView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Title = " Merge "
	setRoundedCorners(v)
	v.Clear()

	fmt.Fprintln(v, "")
	fmt.Fprintln(v, "  [j] Merge Down  [k] Merge Up  [Esc] Cancel")

	g.SetViewOnTop(MergeView)
	g.SetCurrentView(MergeView)

	return nil
}

// renderDialogs renders any active dialog
func (gui *Gui) renderDialogs(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || !gui.state.Dialog.Active {
		g.DeleteView(ConfirmView)
		g.DeleteView(InputView)
		g.DeleteView(HelpView)
		g.DeleteView(MergeView)
		return nil
	}

	switch gui.state.Dialog.Type {
	case "confirm":
		return gui.createConfirmDialog(g, maxX, maxY)
	case "input":
		return gui.createInputDialog(g, maxX, maxY)
	case "help":
		return gui.createHelpDialog(g, maxX, maxY)
	case "merge":
		return gui.createMergeDialog(g, maxX, maxY)
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

	// Help dialog - any key closes
	if err := gui.g.SetKeybinding(HelpView, gocui.KeyEsc, gocui.ModNone, gui.closeHelpDialog); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(HelpView, 'q', gocui.ModNone, gui.closeHelpDialog); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(HelpView, gocui.KeyEnter, gocui.ModNone, gui.closeHelpDialog); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(HelpView, gocui.KeySpace, gocui.ModNone, gui.closeHelpDialog); err != nil {
		return err
	}

	// Merge dialog
	if err := gui.g.SetKeybinding(MergeView, 'j', gocui.ModNone, gui.mergeDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(MergeView, 'k', gocui.ModNone, gui.mergeUp); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(MergeView, gocui.KeyEsc, gocui.ModNone, gui.closeMergeDialog); err != nil {
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

func (gui *Gui) closeHelpDialog(g *gocui.Gui, v *gocui.View) error {
	gui.closeDialog()
	return nil
}

func (gui *Gui) mergeDown(g *gocui.Gui, v *gocui.View) error {
	gui.closeDialog()
	return gui.executeMerge("down")
}

func (gui *Gui) mergeUp(g *gocui.Gui, v *gocui.View) error {
	gui.closeDialog()
	return gui.executeMerge("up")
}

func (gui *Gui) closeMergeDialog(g *gocui.Gui, v *gocui.View) error {
	gui.closeDialog()
	return nil
}
