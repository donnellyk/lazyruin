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

// lineTarget identifies a specific content line in a note file.
type lineTarget struct {
	UUID    string
	LineNum int // 1-indexed content line (after frontmatter)
	Path    string
}

// PreviewLineOpsHelper handles line-level operations: todo toggling,
// done tagging, inline tag/date toggling.
type PreviewLineOpsHelper struct {
	c *HelperCommon
}

// NewPreviewLineOpsHelper creates a new PreviewLineOpsHelper.
func NewPreviewLineOpsHelper(c *HelperCommon) *PreviewLineOpsHelper {
	return &PreviewLineOpsHelper{c: c}
}

func (self *PreviewLineOpsHelper) ctx() *context.CardListContext {
	return self.c.GuiCommon().Contexts().CardList
}

func (self *PreviewLineOpsHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// resolveTarget uses the Lines[] lookup to map the current cursor position
// to a source file line. Works for all preview contexts (cardList, pickResults,
// compose) and the pickDialog overlay.
func (self *PreviewLineOpsHelper) resolveTarget() *lineTarget {
	gui := self.c.GuiCommon()
	var ns *context.PreviewNavState
	if gui.CurrentContextKey() == "pickDialog" {
		ns = gui.Contexts().PickDialog.NavState()
	} else {
		ns = gui.Contexts().ActivePreview().NavState()
	}
	if ns.CursorLine < 0 || ns.CursorLine >= len(ns.Lines) {
		return nil
	}
	src := ns.Lines[ns.CursorLine]
	if src.UUID == "" || src.LineNum == 0 {
		return nil // separator, blank, or frontmatter display line
	}
	return &lineTarget{UUID: src.UUID, LineNum: src.LineNum, Path: src.Path}
}

// ResolveTarget exposes resolveTarget for use by other helpers (e.g. preview nav).
func (self *PreviewLineOpsHelper) ResolveTarget() *lineTarget {
	return self.resolveTarget()
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
	target := self.resolveTarget()
	if target == nil {
		return nil
	}

	err := self.c.RuinCmd().Note.ToggleTodo(target.UUID, target.LineNum)
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	if self.c.GuiCommon().Contexts().ActivePreviewKey == "cardList" {
		cl := self.ctx()
		cl.Cards[cl.SelectedCardIdx].Content = ""
	}
	self.c.Helpers().Preview().ReloadActivePreview()
	self.reloadPickDialogIfActive()
	return nil
}

// AppendDone toggles #done on the current line.
func (self *PreviewLineOpsHelper) AppendDone() error {
	target := self.resolveTarget()
	if target == nil {
		return nil
	}

	srcLine, _, _ := readSourceLine(target.Path, target.LineNum)
	hasDone := false
	for _, m := range inlineTagRe.FindAllString(srcLine, -1) {
		if strings.EqualFold(m, "#done") {
			hasDone = true
			break
		}
	}

	var err error
	if hasDone {
		err = self.c.RuinCmd().Note.RemoveTagFromLine(target.UUID, "#done", target.LineNum)
	} else {
		err = self.c.RuinCmd().Note.AddTagToLine(target.UUID, "#done", target.LineNum)
	}
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	self.c.Helpers().Preview().ReloadActivePreview()
	self.c.Helpers().Tags().RefreshTags(false)
	self.reloadPickDialogIfActive()
	return nil
}

// reloadPickDialogIfActive reloads the pick dialog overlay if it is currently
// the active context. Called after line-level mutations to keep results fresh.
func (self *PreviewLineOpsHelper) reloadPickDialogIfActive() {
	if self.c.GuiCommon().CurrentContextKey() == "pickDialog" {
		self.c.Helpers().Pick().ReloadPickDialog()
	}
}

// ToggleInlineTag opens the input popup to toggle an inline tag on the cursor line.
func (self *PreviewLineOpsHelper) ToggleInlineTag() error {
	target := self.resolveTarget()
	if target == nil {
		return nil
	}

	srcLine, _, _ := readSourceLine(target.Path, target.LineNum)
	existingTags := make(map[string]bool)
	for _, m := range inlineTagRe.FindAllString(srcLine, -1) {
		existingTags[strings.ToLower(m)] = true
	}

	uuid := target.UUID
	lineNum := target.LineNum
	gui := self.c.GuiCommon()
	self.c.Helpers().InputPopup().OpenInputPopup(&types.InputPopupConfig{
		Title:  "Toggle Inline Tag",
		Footer: " # for tags | Tab: accept | Esc: cancel ",
		Seed:   "#",
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{{Prefix: "#", Candidates: func(filter string) []types.CompletionItem {
				items := self.c.Helpers().Completion().TagCandidates(filter)
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
			self.c.Helpers().Preview().ReloadActivePreview()
			self.c.Helpers().Tags().RefreshTags(false)
			self.reloadPickDialogIfActive()
			return nil
		},
	})
	return nil
}

// ToggleInlineDate opens the input popup to toggle an inline date on the cursor line.
func (self *PreviewLineOpsHelper) ToggleInlineDate() error {
	target := self.resolveTarget()
	if target == nil {
		return nil
	}

	srcLine, _, _ := readSourceLine(target.Path, target.LineNum)
	existingDates := make(map[string]bool)
	for _, m := range inlineDateRe.FindAllString(srcLine, -1) {
		existingDates[m] = true
	}

	uuid := target.UUID
	lineNum := target.LineNum
	gui := self.c.GuiCommon()
	self.c.Helpers().InputPopup().OpenInputPopup(&types.InputPopupConfig{
		Title:  "Toggle Inline Date",
		Footer: " @ for dates | Tab: accept | Esc: cancel ",
		Seed:   "@",
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{{Prefix: "@", Candidates: func(filter string) []types.CompletionItem {
				items := AtDateCandidates(filter)
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
			self.c.Helpers().Preview().ReloadActivePreview()
			self.reloadPickDialogIfActive()
			return nil
		},
	})
	return nil
}
