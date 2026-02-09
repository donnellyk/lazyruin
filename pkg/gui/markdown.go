package gui

import (
	"strconv"
	"strings"
)

// mdContinuation describes what should happen when Enter is pressed on a markdown line.
type mdContinuation struct {
	Prefix string // prefix for the next line (e.g. "- ", "> ", "2. ")
	Empty  bool   // true when the line is just the prefix with no content (should clear)
}

// markdownContinuation examines the current line and returns the continuation
// prefix for the next line. Returns nil if no continuation applies.
func markdownContinuation(line string) *mdContinuation {
	// Extract leading whitespace
	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]

	// Task list: - [ ] or - [x]
	for _, taskPrefix := range []string{"- [ ] ", "- [x] ", "- [X] "} {
		if strings.HasPrefix(trimmed, taskPrefix) {
			content := trimmed[len(taskPrefix):]
			return &mdContinuation{
				Prefix: indent + "- [ ] ",
				Empty:  strings.TrimSpace(content) == "",
			}
		}
	}

	// Unordered list: - or *
	for _, bullet := range []string{"- ", "* "} {
		if strings.HasPrefix(trimmed, bullet) {
			content := trimmed[len(bullet):]
			return &mdContinuation{
				Prefix: indent + bullet,
				Empty:  strings.TrimSpace(content) == "",
			}
		}
	}

	// Ordered list: 1. 2. etc.
	if dotIdx := strings.Index(trimmed, ". "); dotIdx > 0 {
		numStr := trimmed[:dotIdx]
		if num, err := strconv.Atoi(numStr); err == nil && num > 0 {
			content := trimmed[dotIdx+2:]
			return &mdContinuation{
				Prefix: indent + strconv.Itoa(num+1) + ". ",
				Empty:  strings.TrimSpace(content) == "",
			}
		}
	}

	// Blockquote: >
	if strings.HasPrefix(trimmed, "> ") {
		content := trimmed[2:]
		return &mdContinuation{
			Prefix: indent + "> ",
			Empty:  strings.TrimSpace(content) == "",
		}
	}

	return nil
}

// currentLine returns the line the cursor is on given unwrapped content and cursor Y.
func currentLine(content string, cy int) string {
	lines := strings.Split(content, "\n")
	if cy < 0 || cy >= len(lines) {
		return ""
	}
	return lines[cy]
}
