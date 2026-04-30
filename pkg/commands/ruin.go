package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// ExecuteAndUnmarshal runs a ruin command and unmarshals the JSON output into T.
func ExecuteAndUnmarshal[T any](r *RuinCommand, args ...string) (T, error) {
	output, err := r.Execute(args...)
	if err != nil {
		var zero T
		return zero, err
	}
	return unmarshalJSON[T](output)
}

// RuinCommand wraps the ruin CLI for executing commands.
type RuinCommand struct {
	vaultPath string
	bin       string
	executor  Executor
	Search    *SearchCommand
	Tags      *TagsCommand
	Queries   *QueriesCommand
	Parent    *ParentCommand
	Pick      *PickCommand
	Note      *NoteCommand
	Link      *LinkCommand
	Embed     *EmbedCommand
}

// NewRuinCommand creates a new RuinCommand with the given vault path and binary.
func NewRuinCommand(vaultPath, bin string) *RuinCommand {
	r := &RuinCommand{
		vaultPath: vaultPath,
		bin:       bin,
	}
	r.initSubcommands()
	return r
}

// NewRuinCommandWithExecutor creates a new RuinCommand with a custom executor.
// Use this for testing with a mock executor.
func NewRuinCommandWithExecutor(executor Executor, vaultPath string) *RuinCommand {
	r := &RuinCommand{
		vaultPath: vaultPath,
		executor:  executor,
	}
	r.initSubcommands()
	return r
}

// initSubcommands wires up all sub-command instances.
func (r *RuinCommand) initSubcommands() {
	r.Search = NewSearchCommand(r)
	r.Tags = NewTagsCommand(r)
	r.Queries = NewQueriesCommand(r)
	r.Parent = NewParentCommand(r)
	r.Pick = NewPickCommand(r)
	r.Note = NewNoteCommand(r)
	r.Link = NewLinkCommand(r)
	r.Embed = NewEmbedCommand(r)
}

// VaultPath returns the configured vault path.
func (r *RuinCommand) VaultPath() string {
	return r.vaultPath
}

// buildCommand creates an exec.Cmd for the ruin CLI with --vault appended.
func (r *RuinCommand) buildCommand(args ...string) *exec.Cmd {
	fullArgs := append(args, "--vault", r.vaultPath)
	return exec.Command(r.bin, fullArgs...)
}

// IsInitialized reports whether the vault path has been initialized as a
// ruin vault. It checks for the presence of the `.ruin` directory, which
// `ruin init` creates. This is a cheap, stat-only check — no subprocess.
func (r *RuinCommand) IsInitialized() bool {
	if r.vaultPath == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(r.vaultPath, ".ruin"))
	return err == nil && info.IsDir()
}

// Init runs `ruin init <vaultPath> --force` to initialize the vault.
// `--force` is required because we invoke as a subprocess with no stdin,
// and current ruin-cli prompts for confirmation when notes already exist.
// Safe in this flow: we only reach here when `.ruin/` does not exist
// (IsInitialized() gated it), so the flag's metadata-overwrite half is
// a no-op. Returns the combined CLI output as an error on failure so
// callers can surface it to the user.
func (r *RuinCommand) Init() error {
	if r.vaultPath == "" {
		return fmt.Errorf("no vault path configured")
	}
	if r.bin == "" {
		return fmt.Errorf("ruin binary path not set")
	}
	cmd := exec.Command(r.bin, "init", r.vaultPath, "--force")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("ruin init failed: %s", msg)
		}
		return fmt.Errorf("ruin init failed: %w", err)
	}
	return nil
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

// Doctor reindexes a specific file after manual editing via `ruin doctor <path>`.
func (r *RuinCommand) Doctor(path string) error {
	cmd := r.buildCommand("doctor", path)
	_, err := cmd.CombinedOutput()
	return err
}

// DoctorFullScan runs `ruin doctor` against the configured vault with
// no path argument, performing a full vault re-index. Used by upgrade
// migrations to bring the metadata back into sync after a breaking
// change. Returns the captured stderr/stdout as the error message on
// failure so the caller can display it inline.
func (r *RuinCommand) DoctorFullScan() error {
	cmd := r.buildCommand("doctor")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("ruin doctor failed: %s", msg)
		}
		return fmt.Errorf("ruin doctor failed: %w", err)
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
