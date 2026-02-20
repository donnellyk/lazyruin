package helpers

import (
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// PreviewInfoHelper handles the info dialog: parent structure and TOC.
type PreviewInfoHelper struct {
	c *HelperCommon
}

// NewPreviewInfoHelper creates a new PreviewInfoHelper.
func NewPreviewInfoHelper(c *HelperCommon) *PreviewInfoHelper {
	return &PreviewInfoHelper{c: c}
}

func (self *PreviewInfoHelper) activeCtx() context.IPreviewContext {
	return self.c.GuiCommon().Contexts().ActivePreview()
}

func (self *PreviewInfoHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// ShowInfoDialog shows parent structure / TOC for the current card.
func (self *PreviewInfoHelper) ShowInfoDialog() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}

	var items []types.MenuItem
	items = append(items, types.MenuItem{Label: "Info: " + card.Title, IsHeader: true})

	if card.Order != nil {
		items = append(items, types.MenuItem{Label: "Order: " + fmt.Sprintf("%d", *card.Order)})
	}

	treeRef := card.UUID
	if card.Parent != "" {
		treeRef = card.Parent
	}
	tree, err := self.c.RuinCmd().Parent.Tree(treeRef)
	if err == nil && (card.Parent != "" || len(tree.Children) > 0) {
		items = append(items, types.MenuItem{})
		items = append(items, types.MenuItem{Label: "Parent", IsHeader: true})
		rootUUID := tree.UUID
		items = append(items, types.MenuItem{Label: "* " + tree.Title, OnRun: func() error {
			return self.c.Helpers().PreviewNav().OpenNoteByUUID(rootUUID)
		}})
		items = self.appendTreeItems(items, tree.Children, "", 5)
	}

	headerItems := self.buildHeaderTOC()
	if len(headerItems) > 0 {
		items = append(items, types.MenuItem{})
		items = append(items, types.MenuItem{Label: "Headers", IsHeader: true})
		items = append(items, headerItems...)
	}

	self.c.GuiCommon().ShowMenuDialog("Info", items)
	return nil
}

func (self *PreviewInfoHelper) appendTreeItems(items []types.MenuItem, children []commands.TreeNode, indent string, maxDepth int) []types.MenuItem {
	if maxDepth <= 0 || len(children) == 0 {
		return items
	}
	for _, child := range children {
		childUUID := child.UUID
		items = append(items, types.MenuItem{
			Label: indent + "  * " + child.Title,
			OnRun: func() error {
				return self.c.Helpers().PreviewNav().OpenNoteByUUID(childUUID)
			},
		})
		items = self.appendTreeItems(items, child.Children, indent+"  ", maxDepth-1)
	}
	return items
}

func (self *PreviewInfoHelper) buildHeaderTOC() []types.MenuItem {
	ctx := self.activeCtx()
	ns := ctx.NavState()
	idx := ctx.SelectedCardIndex()
	if idx >= len(ns.CardLineRanges) {
		return nil
	}
	ranges := ns.CardLineRanges[idx]

	var viewLines []string
	if v := self.view(); v != nil {
		viewLines = v.ViewBufferLines()
	}

	type header struct {
		level    int
		title    string
		viewLine int
	}
	var headers []header

	for _, hLine := range ns.HeaderLines {
		if hLine < ranges[0] || hLine >= ranges[1] {
			continue
		}
		if hLine < len(viewLines) {
			raw := strings.TrimSpace(stripAnsi(viewLines[hLine]))
			level := 0
			for _, r := range raw {
				if r == '#' {
					level++
				} else {
					break
				}
			}
			title := strings.TrimSpace(strings.TrimLeft(raw, "#"))
			headers = append(headers, header{level: level, title: title, viewLine: hLine})
		}
	}

	if len(headers) == 0 {
		return nil
	}

	minLevel := headers[0].level
	for _, h := range headers[1:] {
		if h.level < minLevel {
			minLevel = h.level
		}
	}

	var items []types.MenuItem
	for _, h := range headers {
		depth := h.level - minLevel
		indent := strings.Repeat("  ", depth)
		targetLine := h.viewLine
		items = append(items, types.MenuItem{
			Label: indent + "* " + h.title,
			OnRun: func() error {
				ns.CursorLine = targetLine
				self.c.GuiCommon().RenderPreview()
				return nil
			},
		})
	}
	return items
}
