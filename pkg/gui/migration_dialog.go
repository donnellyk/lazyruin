package gui

import (
	"fmt"
	"strings"

	"github.com/donnellyk/lazyruin/pkg/migrations"

	"github.com/jesseduffield/gocui"
)

// MigrationView is the gocui view name for the non-interactive
// "running migration" modal.
const MigrationView = "migration"

// CloseDialog clears the active dialog. Exposed so the migrations
// helper can dismiss its own running modal from a goroutine via
// gui.Update.
func (gui *Gui) CloseDialog() { gui.closeDialog() }

// ShowMigrationPrompt presents the upgrade-required confirm dialog.
// The user can run via y/Enter or quit via q. n and Esc are no-ops
// (BlockSoftCancel) — migrations are reserved for breaking changes,
// so dismissing without running would leave the vault inconsistent.
func (gui *Gui) ShowMigrationPrompt(prev, curr migrations.VersionPair, vaultPath string, pending []migrations.Migration, onRun func() error) {
	gui.state.Dialog = &DialogState{
		Active:          true,
		Type:            "confirm",
		Title:           "Vault upgrade required",
		Message:         migrationPromptBody(prev, curr, vaultPath, pending),
		Footer:          " [y] Run now · [q] Quit ",
		QuitOnCancel:    true,
		BlockSoftCancel: true,
		OnConfirm:       onRun,
	}
}

// ShowMigrationRunning swaps the dialog to a non-dismissable progress
// modal. Doctor runs on a goroutine; the helper switches to either an
// error dialog or closes the running modal on completion via gui.g.Update.
func (gui *Gui) ShowMigrationRunning(vaultPath string) {
	gui.state.Dialog = &DialogState{
		Active:  true,
		Type:    "migration_running",
		Title:   "Migrating",
		Message: fmt.Sprintf("Running ruin doctor against %s…\n\nThis usually takes a few seconds.", vaultPath),
	}
}

// ShowMigrationError surfaces a doctor failure with a Retry/Quit choice.
// Same input semantics as the prompt: y/Enter retries, q quits, n/Esc
// no-op.
func (gui *Gui) ShowMigrationError(err error, onRetry func() error) {
	body := "ruin doctor failed:\n\n" + indentMultiline(err.Error(), "  ") +
		"\n\nRetry, or quit and run `ruin doctor` manually."
	gui.state.Dialog = &DialogState{
		Active:          true,
		Type:            "confirm",
		Title:           "Migration failed",
		Message:         body,
		Footer:          " [y] Retry · [q] Quit ",
		QuitOnCancel:    true,
		BlockSoftCancel: true,
		OnConfirm:       onRetry,
	}
}

// migrationPromptBody renders the body text for the upgrade prompt: a
// summary line per changed component, then the bulleted list of
// migrations that will run, then the vault path. Centered inside the
// dialog by createConfirmDialog's normal flow.
func migrationPromptBody(prev, curr migrations.VersionPair, vaultPath string, pending []migrations.Migration) string {
	var b strings.Builder
	if prev.Ruin != curr.Ruin {
		fmt.Fprintf(&b, "ruin-cli upgraded: %s → %s\n", displayVersion(prev.Ruin), curr.Ruin)
	}
	if prev.Lazyruin != curr.Lazyruin {
		fmt.Fprintf(&b, "lazyruin: %s → %s\n", displayVersion(prev.Lazyruin), curr.Lazyruin)
	}
	b.WriteString("\nThis upgrade requires re-indexing your vault before lazyruin can continue:\n")
	for _, m := range pending {
		fmt.Fprintf(&b, "  • %s\n", m.Description)
	}
	fmt.Fprintf(&b, "\nRuns `ruin doctor` against %s — usually a few seconds.", vaultPath)
	return b.String()
}

// displayVersion formats a version for the prompt's "X → Y" line. An
// empty value (vault never seen before) renders as "(unknown)".
func displayVersion(v string) string {
	if v == "" {
		return "(unknown)"
	}
	return v
}

func indentMultiline(s, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

// createMigrationRunningDialog renders the non-interactive progress
// modal. No keybindings are registered for the view — the only way out
// is the helper's gui.g.Update callback flipping state to a different
// dialog (or `<c-c>` killing the process).
func (gui *Gui) createMigrationRunningDialog(g *gocui.Gui, maxX, maxY int) error {
	if gui.state.Dialog == nil || gui.state.Dialog.Type != "migration_running" {
		return nil
	}

	msgLines := strings.Count(gui.state.Dialog.Message, "\n") + 1
	width, height := 64, msgLines+4
	x0, y0, x1, y1 := centerRect(maxX, maxY, width, height)

	v, err := g.SetView(MigrationView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Title = " " + gui.state.Dialog.Title + " "
	v.Footer = ""
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorYellow
	v.TitleColor = gocui.ColorYellow
	v.Clear()

	fmt.Fprintln(v, "")
	for _, line := range strings.Split(gui.state.Dialog.Message, "\n") {
		fmt.Fprintln(v, "  "+line)
	}

	g.SetViewOnTop(MigrationView)
	g.SetCurrentView(MigrationView)
	return nil
}
