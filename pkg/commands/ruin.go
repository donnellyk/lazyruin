package commands

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// unmarshalJSON is a generic helper that unmarshals JSON data into a value of type T.
func unmarshalJSON[T any](data []byte) (T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}
	return result, nil
}

// RuinCommand wraps the ruin CLI for executing commands.
type RuinCommand struct {
	vaultPath string
	executor  Executor
	Search    *SearchCommand
	Tags      *TagsCommand
	Queries   *QueriesCommand
	Parent    *ParentCommand
	Pick      *PickCommand
}

// NewRuinCommand creates a new RuinCommand with the given vault path.
func NewRuinCommand(vaultPath string) *RuinCommand {
	r := &RuinCommand{
		vaultPath: vaultPath,
	}
	r.Search = NewSearchCommand(r)
	r.Tags = NewTagsCommand(r)
	r.Queries = NewQueriesCommand(r)
	r.Parent = NewParentCommand(r)
	r.Pick = NewPickCommand(r)

	return r
}

// NewRuinCommandWithExecutor creates a new RuinCommand with a custom executor.
// Use this for testing with a mock executor.
func NewRuinCommandWithExecutor(executor Executor, vaultPath string) *RuinCommand {
	r := &RuinCommand{
		vaultPath: vaultPath,
		executor:  executor,
	}
	r.Search = NewSearchCommand(r)
	r.Tags = NewTagsCommand(r)
	r.Queries = NewQueriesCommand(r)
	r.Parent = NewParentCommand(r)
	r.Pick = NewPickCommand(r)

	return r
}

// VaultPath returns the configured vault path.
func (r *RuinCommand) VaultPath() string {
	return r.vaultPath
}

// buildCommand creates an exec.Cmd for the ruin CLI with --vault appended.
func (r *RuinCommand) buildCommand(args ...string) *exec.Cmd {
	fullArgs := append(args, "--vault", r.vaultPath)
	return exec.Command("ruin", fullArgs...)
}

// CheckVault verifies the vault is accessible, returning an error with details if not.
func (r *RuinCommand) CheckVault() error {
	cmd := r.buildCommand("today")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg != "" {
			return fmt.Errorf("vault check failed for %q: %s", r.vaultPath, msg)
		}
		return fmt.Errorf("vault check failed for %q: %w", r.vaultPath, err)
	}
	return nil
}

// Execute runs a ruin command with the given arguments.
// It automatically appends --json and --vault flags.
func (r *RuinCommand) Execute(args ...string) ([]byte, error) {
	// Use injected executor if available
	if r.executor != nil {
		return r.executor.Execute(args...)
	}

	// Default to CLI execution
	cmd := r.buildCommand(append(args, "--json")...)
	return cmd.Output()
}
