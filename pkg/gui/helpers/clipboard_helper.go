package helpers

import (
	"os/exec"
	"strings"
)

// ClipboardHelper manages clipboard operations.
type ClipboardHelper struct {
	c *HelperCommon
}

// NewClipboardHelper creates a new ClipboardHelper.
func NewClipboardHelper(c *HelperCommon) *ClipboardHelper {
	return &ClipboardHelper{c: c}
}

// CopyToClipboard copies the given text to the system clipboard.
// Returns nil if no clipboard command is available.
func (self *ClipboardHelper) CopyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch {
	case isCommandAvailable("pbcopy"):
		cmd = exec.Command("pbcopy")
	case isCommandAvailable("xclip"):
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case isCommandAvailable("xsel"):
		cmd = exec.Command("xsel", "--clipboard", "--input")
	default:
		return nil
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// isCommandAvailable checks if a command exists in PATH.
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
