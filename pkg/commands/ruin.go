package commands

import (
	"os/exec"
)

// RuinCommand wraps the ruin CLI for executing commands.
type RuinCommand struct {
	vaultPath string
	Search    *SearchCommand
	Tags      *TagsCommand
	Queries   *QueriesCommand
}

// NewRuinCommand creates a new RuinCommand with the given vault path.
func NewRuinCommand(vaultPath string) *RuinCommand {
	r := &RuinCommand{vaultPath: vaultPath}
	r.Search = NewSearchCommand(r)
	r.Tags = NewTagsCommand(r)
	r.Queries = NewQueriesCommand(r)

	return r
}

// VaultPath returns the configured vault path.
func (r *RuinCommand) VaultPath() string {
	return r.vaultPath
}

// VaultExists checks if the vault is accessible.
func (r *RuinCommand) VaultExists() bool {
	// Use 'ruin today' as a lightweight check that the vault is valid
	cmd := exec.Command("ruin", "today", "--vault", r.vaultPath)
	err := cmd.Run()
	return err == nil
}

// Execute runs a ruin command with the given arguments.
// It automatically appends --json and --vault flags.
func (r *RuinCommand) Execute(args ...string) ([]byte, error) {
	fullArgs := append(args, "--json", "--vault", r.vaultPath)
	cmd := exec.Command("ruin", fullArgs...)
	return cmd.Output()
}
