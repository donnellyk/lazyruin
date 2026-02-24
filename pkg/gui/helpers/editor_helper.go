package helpers

import (
	"os"
	"os/exec"
	"strings"
)

// EditorHelper manages editor suspension and resumption.
type EditorHelper struct {
	c *HelperCommon
}

// NewEditorHelper creates a new EditorHelper.
func NewEditorHelper(c *HelperCommon) *EditorHelper {
	return &EditorHelper{c: c}
}

// OpenInEditor suspends the TUI, opens the given file in $EDITOR,
// resumes the TUI, and refreshes all data.
func (self *EditorHelper) OpenInEditor(path string) error {
	gui := self.c.GuiCommon()

	if err := gui.Suspend(); err != nil {
		return err
	}

	editor := self.c.Config().Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vim"
	}

	parts := strings.Fields(editor)
	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	if err := gui.Resume(); err != nil {
		return err
	}

	self.c.Helpers().Tags().RefreshTags(false)
	self.c.Helpers().Queries().RefreshQueries(false)
	self.c.Helpers().Queries().RefreshParents(false)
	self.c.Helpers().Preview().ReloadContent()
	gui.RenderAll()
	return nil
}
