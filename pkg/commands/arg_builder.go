package commands

// ArgBuilder provides a fluent API for constructing command arguments
// with conditional flags, reducing repetitive if/append patterns.
type ArgBuilder struct {
	args []string
}

// NewArgBuilder creates an ArgBuilder with the given initial arguments.
func NewArgBuilder(initial ...string) *ArgBuilder {
	return &ArgBuilder{args: initial}
}

// Add appends arguments unconditionally.
func (b *ArgBuilder) Add(args ...string) *ArgBuilder {
	b.args = append(b.args, args...)
	return b
}

// AddIf appends arguments only when the condition is true.
func (b *ArgBuilder) AddIf(condition bool, args ...string) *ArgBuilder {
	if condition {
		b.args = append(b.args, args...)
	}
	return b
}

// Build returns the assembled argument slice.
func (b *ArgBuilder) Build() []string {
	return b.args
}
