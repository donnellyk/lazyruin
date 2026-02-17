package commands

import "fmt"

// NoteCommand wraps the `ruin note` subcommands.
type NoteCommand struct {
	ruin *RuinCommand
}

func NewNoteCommand(ruin *RuinCommand) *NoteCommand {
	return &NoteCommand{ruin: ruin}
}

// SetParent sets the parent of a note via `note set --parent`.
func (n *NoteCommand) SetParent(noteRef, parentRef string) error {
	_, err := n.ruin.Execute("note", "set", noteRef, "--parent", parentRef, "-f")
	return err
}

// RemoveParent removes the parent from a note via `note set --no-parent`.
func (n *NoteCommand) RemoveParent(noteRef string) error {
	_, err := n.ruin.Execute("note", "set", noteRef, "--no-parent", "-f")
	return err
}

// AddTag adds a global tag to a note via `note set --add-tag`.
func (n *NoteCommand) AddTag(noteRef, tag string) error {
	_, err := n.ruin.Execute("note", "set", noteRef, "--add-tag", tag, "-f")
	return err
}

// RemoveTag removes a tag from a note via `note set --remove-tag`.
func (n *NoteCommand) RemoveTag(noteRef, tag string) error {
	_, err := n.ruin.Execute("note", "set", noteRef, "--remove-tag", tag, "-f")
	return err
}

// AddTagToLine adds an inline tag to a specific content line via `note set --add-tag --line`.
func (n *NoteCommand) AddTagToLine(noteRef, tag string, line int) error {
	_, err := n.ruin.Execute("note", "set", noteRef, "--add-tag", tag,
		"--line", fmt.Sprintf("%d", line), "-f")
	return err
}

// RemoveTagFromLine removes an inline tag from a specific content line via `note set --remove-tag --line`.
func (n *NoteCommand) RemoveTagFromLine(noteRef, tag string, line int) error {
	_, err := n.ruin.Execute("note", "set", noteRef, "--remove-tag", tag,
		"--line", fmt.Sprintf("%d", line), "-f")
	return err
}

// SetOrder sets the order frontmatter field on a note.
func (n *NoteCommand) SetOrder(noteRef string, order int) error {
	_, err := n.ruin.Execute("note", "set", noteRef, "--order", fmt.Sprintf("%d", order), "-f")
	return err
}

// RemoveOrder unsets the order field via `note set --no-order`.
func (n *NoteCommand) RemoveOrder(noteRef string) error {
	_, err := n.ruin.Execute("note", "set", noteRef, "--no-order", "-f")
	return err
}

// SetField sets an extra frontmatter field via `note set --field key=value`.
func (n *NoteCommand) SetField(noteRef, key, value string) error {
	_, err := n.ruin.Execute("note", "set", noteRef, "--field", key+"="+value, "-f")
	return err
}

// Append inserts or appends text in a note's content via `note append`.
// If suffix is true, appends to the end of the line; otherwise inserts before it.
// line is 1-indexed (content lines after frontmatter). Use 0 to append at end.
func (n *NoteCommand) Append(noteRef, text string, line int, suffix bool) error {
	args := []string{"note", "append", noteRef, text}
	if line > 0 {
		args = append(args, "--line", fmt.Sprintf("%d", line))
	}
	if suffix {
		args = append(args, "--suffix")
	}
	args = append(args, "-f")
	_, err := n.ruin.Execute(args...)
	return err
}

// MergeResult holds the output from a note merge operation.
type MergeResult struct {
	TargetPath    string   `json:"target_path"`
	TargetUUID    string   `json:"target_uuid"`
	SourcePath    string   `json:"source_path"`
	SourceUUID    string   `json:"source_uuid"`
	TagsMerged    []string `json:"tags_merged"`
	ChildrenMoved int      `json:"children_moved"`
	SourceDeleted bool     `json:"source_deleted"`
}

// Merge merges a source note into a target note via `note merge`.
func (n *NoteCommand) Merge(targetRef, sourceRef string, deleteSource, stripTitle bool) (*MergeResult, error) {
	args := []string{"note", "merge", targetRef, sourceRef}
	if deleteSource {
		args = append(args, "--delete-source")
	}
	if stripTitle {
		args = append(args, "--strip-title")
	}
	args = append(args, "-f")
	output, err := n.ruin.Execute(args...)
	if err != nil {
		return nil, err
	}
	return unmarshalJSON[*MergeResult](output)
}
