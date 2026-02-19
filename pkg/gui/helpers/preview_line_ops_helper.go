package helpers

import (
	"os"
	"regexp"
	"strings"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// regex patterns for line operations
var (
	inlineTagRe  = regexp.MustCompile(`#[\w-]+`)
	inlineDateRe = regexp.MustCompile(`@\d{4}-\d{2}-\d{2}`)
)

// PreviewLineOpsHelper handles line-level operations: todo toggling,
// done tagging, inline tag/date toggling.
type PreviewLineOpsHelper struct {
	c *HelperCommon
}

// NewPreviewLineOpsHelper creates a new PreviewLineOpsHelper.
func NewPreviewLineOpsHelper(c *HelperCommon) *PreviewLineOpsHelper {
	return &PreviewLineOpsHelper{c: c}
}

func (self *PreviewLineOpsHelper) ctx() *context.PreviewContext {
	return self.c.GuiCommon().Contexts().Preview
}

func (self *PreviewLineOpsHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// resolveSourceLine maps the current visual cursor position to a 1-indexed
// content line number in the raw source file (after frontmatter).
func (self *PreviewLineOpsHelper) resolveSourceLine() int {
	v := self.view()
	if v == nil {
		return -1
	}
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return -1
	}
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	ranges := pc.CardLineRanges
	if idx >= len(ranges) {
		return -1
	}

	cardStart := ranges[idx][0]
	lineOffset := pc.CursorLine - cardStart - 1
	if lineOffset < 0 {
		return -1
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)
	cardLines := self.c.GuiCommon().BuildCardContent(*card, contentWidth)
	if lineOffset >= len(cardLines) {
		return -1
	}
	visibleLine := strings.TrimSpace(stripAnsi(cardLines[lineOffset]))
	if visibleLine == "" {
		return -1
	}

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

	for i := contentStart; i < len(fileLines); i++ {
		if strings.TrimSpace(fileLines[i]) == visibleLine {
			return i - contentStart + 1
		}
	}
	return -1
}

// readSourceLine reads the raw source file and returns the content line at
// the given 1-indexed content line number.
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

// ToggleTodo toggles a todo checkbox on the current line.
func (self *PreviewLineOpsHelper) ToggleTodo() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := self.resolveSourceLine()
	if lineNum < 1 {
		return nil
	}

	err := self.c.RuinCmd().Note.ToggleTodo(card.UUID, lineNum)
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	self.ctx().Cards[self.ctx().SelectedCardIndex].Content = ""
	self.c.Helpers().Preview().ReloadContent()
	return nil
}

// AppendDone toggles #done on the current line.
func (self *PreviewLineOpsHelper) AppendDone() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := self.resolveSourceLine()
	if lineNum < 1 {
		return nil
	}

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
		err = self.c.RuinCmd().Note.RemoveTagFromLine(card.UUID, "#done", lineNum)
	} else {
		err = self.c.RuinCmd().Note.AddTagToLine(card.UUID, "#done", lineNum)
	}
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	self.c.Helpers().Preview().ReloadContent()
	self.c.GuiCommon().RefreshTags(false)
	return nil
}

// ToggleInlineTag opens the input popup to toggle an inline tag on the cursor line.
func (self *PreviewLineOpsHelper) ToggleInlineTag() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := self.resolveSourceLine()
	if lineNum < 1 {
		return nil
	}

	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	existingTags := make(map[string]bool)
	for _, m := range inlineTagRe.FindAllString(srcLine, -1) {
		existingTags[strings.ToLower(m)] = true
	}

	uuid := card.UUID
	gui := self.c.GuiCommon()
	gui.OpenInputPopup(&types.InputPopupConfig{
		Title:  "Toggle Inline Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{{Prefix: "#", Candidates: func(filter string) []types.CompletionItem {
				items := gui.TagCandidates(filter)
				var onLine, rest []types.CompletionItem
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
		OnAccept: func(_ string, item *types.CompletionItem) error {
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
				if err := self.c.RuinCmd().Note.RemoveTagFromLine(uuid, tag, lineNum); err != nil {
					gui.ShowError(err)
					return nil
				}
			} else {
				if err := self.c.RuinCmd().Note.AddTagToLine(uuid, tag, lineNum); err != nil {
					gui.ShowError(err)
					return nil
				}
			}
			self.c.Helpers().Preview().ReloadContent()
			gui.RefreshTags(false)
			return nil
		},
	})
	return nil
}

// ToggleInlineDate opens the input popup to toggle an inline date on the cursor line.
func (self *PreviewLineOpsHelper) ToggleInlineDate() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	lineNum := self.resolveSourceLine()
	if lineNum < 1 {
		return nil
	}

	srcLine, _, _ := readSourceLine(card.Path, lineNum)
	existingDates := make(map[string]bool)
	for _, m := range inlineDateRe.FindAllString(srcLine, -1) {
		existingDates[m] = true
	}

	uuid := card.UUID
	gui := self.c.GuiCommon()
	gui.OpenInputPopup(&types.InputPopupConfig{
		Title:  "Toggle Inline Date",
		Footer: " @ for dates | Tab: accept | Esc: cancel ",
		Seed:   "@",
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{{Prefix: "@", Candidates: func(filter string) []types.CompletionItem {
				items := gui.AtDateCandidates(filter)
				var onLine, rest []types.CompletionItem
				for _, item := range items {
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
		OnAccept: func(_ string, item *types.CompletionItem) error {
			if item == nil || item.InsertText == "" {
				return nil
			}
			dateArg := strings.TrimPrefix(item.InsertText, "@")

			if existingDates[item.InsertText] {
				if err := self.c.RuinCmd().Note.RemoveDateFromLine(uuid, dateArg, lineNum); err != nil {
					gui.ShowError(err)
					return nil
				}
			} else {
				if err := self.c.RuinCmd().Note.AddDateToLine(uuid, dateArg, lineNum); err != nil {
					gui.ShowError(err)
					return nil
				}
			}
			self.c.Helpers().Preview().ReloadContent()
			return nil
		},
	})
	return nil
}
