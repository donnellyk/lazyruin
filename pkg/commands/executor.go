package commands

// Executor defines the interface for executing ruin CLI commands.
// Used by tests to inject mock executors via NewRuinCommandWithExecutor.
type Executor interface {
	Execute(args ...string) ([]byte, error)
}
