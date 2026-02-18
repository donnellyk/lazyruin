package gui

import (
	"os"
	"regexp"
	"strings"

	"github.com/jesseduffield/gocui"
)

// inlineTagRe matches #tag patterns (hashtag followed by word characters).
var inlineTagRe = regexp.MustCompile(`#[\w-]+`)

// inlineDateRe matches @YYYY-MM-DD patterns on a content line.
var inlineDateRe = regexp.MustCompile(`@\d{4}-\d{2}-\d{2}`)

// readSourceLine reads the raw source file and returns the content line at
// the given 1-indexed content line number (after frontmatter). Returns ""
// and the file lines if the line is out of range.
func readSourceLine(path string, lineNum int) (string, []string, int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, 0
	}
	fileLines := strings.Split(string(data), "\n")
	contentStart := 0
	if len(fileLines) > 0 && strings.HasPrefix(fileLines[0], "---") {
		for i := 1; i < len(fileLines); i++ {
			if strings.TrimSpace(fileLines[i]) == "---" {
				contentStart = i + 1
				break
			}
		}
	}
	absIdx := contentStart + lineNum - 1
	if absIdx < 0 || absIdx >= len(fileLines) {
		return "", fileLines, contentStart
	}
	return fileLines[absIdx], fileLines, contentStart
}

// resolveSourceLine maps the current visual cursor position to a 1-indexed
// content line number in the raw source file (after frontmatter). This accounts
// for title and global-tag lines that may be stripped from note.Content by
// search options. Returns -1 if the cursor is not on a matchable content line.
func (c *PreviewController) resolveSourceLine(v *gocui.View) int {
	if v == nil {
		return -1
	}
	card := c.currentPreviewCard()
	if card == nil {
		return -1
	}
	idx := c.gui.state.Preview.SelectedCardIndex
	ranges := c.gui.state.Preview.CardLineRanges
	if idx >= len(ranges) {
		return -1
	}

	// Get the visible text at the cursor
	cardStart := ranges[idx][0]
	lineOffset := c.gui.state.Preview.CursorLine - cardStart - 1 // -1 for separator
	if lineOffset < 0 {
		return -1
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)
	cardLines := c.gui.buildCardContent(*card, contentWidth)
	if lineOffset >= len(cardLines) {
		return -1
	}
	visibleLine := strings.TrimSpace(stripAnsi(cardLines[lineOffset]))
	if visibleLine == "" {
		return -1
	}

	// Read the raw file and find the content start (after frontmatter)
	data, err := os.ReadFile(card.Path)
	if err != nil {
		return -1
	}
	fileLines := strings.Split(string(data), "\n")

	contentStart := 0
	if len(fileLines) > 0 && strings.HasPrefix(fileLines[0], "---") {
		for i := 1; i < len(fileLines); i++ {
			if strings.TrimSpace(fileLines[i]) == "---" {
				contentStart = i + 1
				break
			}
		}
	}
	// Do NOT skip leading blank lines here — the CLI's --line N counts
	// every line after the closing --- delimiter, including blanks.

	// Match the visible text against raw file lines
	for i := contentStart; i < len(fileLines); i++ {
		if strings.TrimSpace(fileLines[i]) == visibleLine {
			return i - contentStart + 1 // 1-indexed content line
		}
	}
	return -1
}

func (c *PreviewController) toggleTodo(g *gocui.Gui, v *gocui.View) error {
	card := c.currentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := c.resolveSourceLine(v)
	if lineNum < 1 {
		return nil
	}

	err := c.gui.ruinCmd.Note.ToggleTodo(card.UUID, lineNum)
	if err != nil {
		c.gui.showError(err)
		return nil
	}

	// Clear cached content so it re-reads from disk
	c.gui.state.Preview.Cards[c.gui.state.Preview.SelectedCardIndex].Content = ""
	c.reloadContent()
	return nil
}

// appendDone appends " #done" to the current line via note append.
func (c *PreviewController) appendDone(g *gocui.Gui, v *gocui.View) error {
	card := c.currentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := c.resolveSourceLine(v)
	if lineNum < 1 {
		return nil
	}

	// Check if #done already exists on this line
	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	hasDone := false
	for _, m := range inlineTagRe.FindAllString(srcLine, -1) {
		if strings.EqualFold(m, "#done") {
			hasDone = true
			break
		}
	}

	var err error
	if hasDone {
		err = c.gui.ruinCmd.Note.RemoveTagFromLine(card.UUID, "#done", lineNum)
	} else {
		err = c.gui.ruinCmd.Note.AddTagToLine(card.UUID, "#done", lineNum)
	}
	if err != nil {
		c.gui.showError(err)
		return nil
	}

	c.reloadContent()
	c.gui.refreshTags(false)
	return nil
}

// toggleInlineTag opens the input popup to add or remove an inline tag on the cursor line.
// If the selected tag already appears on the line, it is removed; otherwise it is appended.
func (c *PreviewController) toggleInlineTag(g *gocui.Gui, v *gocui.View) error {
	card := c.currentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := c.resolveSourceLine(v)
	if lineNum < 1 {
		return nil
	}

	// Read the source line to detect existing inline tags
	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	existingTags := make(map[string]bool)
	for _, m := range inlineTagRe.FindAllString(srcLine, -1) {
		existingTags[strings.ToLower(m)] = true
	}

	uuid := card.UUID
	c.gui.openInputPopup(&InputPopupConfig{
		Title:  "Toggle Inline Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{{Prefix: "#", Candidates: func(filter string) []CompletionItem {
				items := c.gui.tagCandidates(filter)
				var onLine, rest []CompletionItem
				for _, item := range items {
					if existingTags[strings.ToLower(item.Label)] {
						item.Detail = "*"
						onLine = append(onLine, item)
					} else {
						rest = append(rest, item)
					}
				}
				return append(onLine, rest...)
			}}}
		},
		OnAccept: func(_ string, item *CompletionItem) error {
			tag := ""
			if item != nil {
				tag = item.Label
			}
			if tag == "" {
				return nil
			}
			if !strings.HasPrefix(tag, "#") {
				tag = "#" + tag
			}

			if existingTags[strings.ToLower(tag)] {
				// Remove: strip the tag from the source line via CLI
				if err := c.gui.ruinCmd.Note.RemoveTagFromLine(uuid, tag, lineNum); err != nil {
					c.gui.showError(err)
					return nil
				}
			} else {
				// Add: append tag to end of line via CLI
				if err := c.gui.ruinCmd.Note.AddTagToLine(uuid, tag, lineNum); err != nil {
					c.gui.showError(err)
					return nil
				}
			}
			c.reloadContent()
			c.gui.refreshTags(false)
			return nil
		},
	})
	return nil
}

// toggleInlineDate opens the input popup to add or remove an inline date on the cursor line.
// If the selected date already appears on the line, it is removed; otherwise it is appended.
func (c *PreviewController) toggleInlineDate(g *gocui.Gui, v *gocui.View) error {
	card := c.currentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := c.resolveSourceLine(v)
	if lineNum < 1 {
		return nil
	}

	// Read the source line to detect existing inline dates
	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	existingDates := make(map[string]bool)
	for _, m := range inlineDateRe.FindAllString(srcLine, -1) {
		existingDates[m] = true // keep the @YYYY-MM-DD form for matching
	}

	uuid := card.UUID
	c.gui.openInputPopup(&InputPopupConfig{
		Title:  "Toggle Inline Date",
		Footer: " @ for dates | Tab: accept | Esc: cancel ",
		Seed:   "@",
		Triggers: func() []CompletionTrigger {
			return []CompletionTrigger{{Prefix: "@", Candidates: func(filter string) []CompletionItem {
				items := atDateCandidates(filter)
				var onLine, rest []CompletionItem
				for _, item := range items {
					// InsertText is e.g. "@today" or "@2026-02-17"; check resolved form
					// against existing dates on the line
					if existingDates[item.InsertText] {
						item.Detail = "*"
						onLine = append(onLine, item)
					} else {
						rest = append(rest, item)
					}
				}
				return append(onLine, rest...)
			}}}
		},
		OnAccept: func(_ string, item *CompletionItem) error {
			if item == nil || item.InsertText == "" {
				return nil
			}
			// Strip the @ prefix — CLI expects "today" or "2026-02-17"
			dateArg := strings.TrimPrefix(item.InsertText, "@")

			if existingDates[item.InsertText] {
				if err := c.gui.ruinCmd.Note.RemoveDateFromLine(uuid, dateArg, lineNum); err != nil {
					c.gui.showError(err)
					return nil
				}
			} else {
				if err := c.gui.ruinCmd.Note.AddDateToLine(uuid, dateArg, lineNum); err != nil {
					c.gui.showError(err)
					return nil
				}
			}
			c.reloadContent()
			return nil
		},
	})
	return nil
}
