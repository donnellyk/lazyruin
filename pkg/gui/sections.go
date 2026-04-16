package gui

import (
	"strings"

	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/donnellyk/ruin-note-cli/pkg/notetext"
)

// isRawHeaderLine returns (level, true) for ATX headers. Delegates the
// "is this a header" check to notetext.IsHeaderLine (which requires a
// trailing space, rejecting `#tag` lines) and separately reports depth.
func isRawHeaderLine(line string) (int, bool) {
	if !notetext.IsHeaderLine(line) {
		return 0, false
	}
	trimmed := strings.TrimLeft(line, " \t")
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	return level, true
}

// sectionRange describes a markdown section by its header line index,
// depth, and the exclusive end line.
type sectionRange struct {
	headerIdx int
	level     int
	endIdx    int
}

// computeDoneSections returns a flag per content line indicating whether
// that line is part of a fully-done section (D2 recursive rule).
//
// A section is fully-done when either its header contains `#done`, or
// every non-blank, non-header, non-tag-only direct content line is
// tagged `#done` AND every nested sub-section is also fully-done.
// A section with zero non-blank direct content lines (header only) is
// NOT fully-done — avoids dimming empty scaffolding.
//
// Fenced and inline code ranges from notetext.FindCodeRanges are
// projected onto per-line flags so `# comment` inside code isn't
// misread as a header.
func computeDoneSections(contentLines []string) []bool {
	result := make([]bool, len(contentLines))
	if len(contentLines) == 0 {
		return result
	}

	inCode := codeLineFlags(contentLines)
	sections := buildSectionRanges(contentLines, inCode)

	done := make([]bool, len(sections))
	for i := len(sections) - 1; i >= 0; i-- {
		done[i] = sectionFullyDone(contentLines, sections, i, done, inCode)
	}

	for i, sec := range sections {
		if !done[i] {
			continue
		}
		for j := sec.headerIdx; j < sec.endIdx; j++ {
			result[j] = true
		}
	}
	return result
}

// buildSectionRanges scans contentLines in order and produces a
// sectionRange for every ATX header outside code fences. Each section's
// endIdx stops at the next header of equal or shallower depth.
func buildSectionRanges(contentLines []string, inCode []bool) []sectionRange {
	var sections []sectionRange
	for i, line := range contentLines {
		if inCode[i] {
			continue
		}
		level, ok := isRawHeaderLine(line)
		if !ok {
			continue
		}
		sections = append(sections, sectionRange{headerIdx: i, level: level, endIdx: len(contentLines)})
	}
	for i := range sections {
		for j := i + 1; j < len(sections); j++ {
			if sections[j].level <= sections[i].level {
				sections[i].endIdx = sections[j].headerIdx
				break
			}
		}
	}
	return sections
}

// sectionFullyDone evaluates the D2 rule for sections[idx].
func sectionFullyDone(contentLines []string, sections []sectionRange, idx int, done []bool, inCode []bool) bool {
	sec := sections[idx]

	if models.HasDoneTag(contentLines[sec.headerIdx]) {
		return true
	}

	// Track direct prose/task content separately from sub-sections so that
	// a section whose only content is sub-sections (all done) still counts
	// as populated for the empty-body guard at the bottom.
	hasDirectContent := false
	hasSubSection := false
	i := sec.headerIdx + 1
	for i < sec.endIdx {
		subIdx := findSubSectionAtLine(sections, idx, i)
		if subIdx >= 0 {
			if !done[subIdx] {
				return false
			}
			hasSubSection = true
			i = sections[subIdx].endIdx
			continue
		}

		line := contentLines[i]
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}
		// Lines inside fenced code blocks — including the fence delimiters
		// themselves — are treated like tag-only lines: neither prose to
		// complete nor a blocker. Otherwise a "Done" section containing any
		// code snippet could never satisfy the all-direct-content rule.
		if inCode[i] {
			i++
			continue
		}
		if notetext.IsTagOnlyLine(line) {
			i++
			continue
		}
		hasDirectContent = true
		if !models.HasDoneTag(line) {
			return false
		}
		i++
	}

	return hasDirectContent || hasSubSection
}

// findSubSectionAtLine returns the sections[] index of a direct child
// section of parent whose headerIdx == line, or -1. "Direct" means the
// child's header depth is strictly greater than the parent's and there is
// no intermediate sub-section header at a depth ≤ parent's between them.
// Because sections are emitted in header order and nested sections at
// deeper levels share the parent's outer range, the first section with
// headerIdx == line and level > parent.level is necessarily a direct
// child when the parent is the nearest ancestor.
func findSubSectionAtLine(sections []sectionRange, parentIdx, line int) int {
	parent := sections[parentIdx]
	for j := parentIdx + 1; j < len(sections); j++ {
		s := sections[j]
		if s.headerIdx >= parent.endIdx {
			return -1
		}
		if s.headerIdx != line {
			continue
		}
		if s.level > parent.level {
			return j
		}
	}
	return -1
}

// codeLineFlags returns a flag per line indicating "any part of this
// line lies inside a code span or fenced code block," using the CLI's
// notetext.FindCodeRanges as the source of truth. A line is flagged if
// any byte of it falls within a code range.
func codeLineFlags(contentLines []string) []bool {
	flags := make([]bool, len(contentLines))
	if len(contentLines) == 0 {
		return flags
	}
	content := strings.Join(contentLines, "\n")
	ranges := notetext.FindCodeRanges(content)
	if len(ranges) == 0 {
		return flags
	}
	offset := 0
	for i, line := range contentLines {
		lineStart := offset
		lineEnd := offset + len(line)
		for _, r := range ranges {
			if r[0] < lineEnd && r[1] > lineStart {
				flags[i] = true
				break
			}
		}
		offset = lineEnd + 1 // +1 for the joining newline
	}
	return flags
}
