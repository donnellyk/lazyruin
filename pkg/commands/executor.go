package commands

import "os/exec"

// Executor defines the interface for executing ruin CLI commands.
type Executor interface {
	Execute(args ...string) ([]byte, error)
}

// CLIExecutor executes commands via the real ruin CLI.
type CLIExecutor struct {
	vaultPath string
}

// NewCLIExecutor creates a new CLI executor with the given vault path.
func NewCLIExecutor(vaultPath string) *CLIExecutor {
	return &CLIExecutor{vaultPath: vaultPath}
}

// Execute runs a ruin command with the given arguments.
// It automatically appends --json and --vault flags.
func (e *CLIExecutor) Execute(args ...string) ([]byte, error) {
	fullArgs := append(args, "--json", "--vault", e.vaultPath)
	cmd := exec.Command("ruin", fullArgs...)
	return cmd.Output()
}

// VaultPath returns the configured vault path.
func (e *CLIExecutor) VaultPath() string {
	return e.vaultPath
}
