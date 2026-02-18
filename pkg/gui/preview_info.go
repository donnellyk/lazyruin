package gui

import (
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/commands"

	"github.com/jesseduffield/gocui"
)

// showInfoDialog shows parent structure / TOC for the current card.
func (c *PreviewController) showInfoDialog(g *gocui.Gui, v *gocui.View) error {
	card := c.currentPreviewCard()
	if card == nil {
		return nil
	}

	var items []MenuItem
	items = append(items, MenuItem{Label: "Info: " + card.Title, IsHeader: true})

	if card.Order != nil {
		items = append(items, MenuItem{Label: fmt.Sprintf("Order: %d", *card.Order)})
	}

	// Parent tree (git-log-graph style)
	// If the note has a parent, show the tree rooted at the parent.
	// Otherwise, show the tree rooted at the current note (if it has children).
	treeRef := card.UUID
	if card.Parent != "" {
		treeRef = card.Parent
	}
	tree, err := c.gui.ruinCmd.Parent.Tree(treeRef)
	if err == nil && (card.Parent != "" || len(tree.Children) > 0) {
		items = append(items, MenuItem{})
		items = append(items, MenuItem{Label: "Parent", IsHeader: true})
		rootUUID := tree.UUID
		items = append(items, MenuItem{Label: "* " + tree.Title, OnRun: func() error {
			return c.openNoteByUUID(rootUUID)
		}})
		items = c.appendTreeItems(items, tree.Children, "", 5)
	}

	// TOC from headers (git-log-graph style with actual header titles)
	headerItems := c.buildHeaderTOC()
	if len(headerItems) > 0 {
		items = append(items, MenuItem{})
		items = append(items, MenuItem{Label: "Headers", IsHeader: true})
		items = append(items, headerItems...)
	}

	c.gui.state.Dialog = &DialogState{
		Active:        true,
		Type:          "menu",
		Title:         "Info",
		MenuItems:     items,
		MenuSelection: 0,
	}
	return nil
}

// appendTreeItems recursively adds children to the menu with indented * bullets.
// Each child item gets an OnRun that opens that note in the preview.
func (c *PreviewController) appendTreeItems(items []MenuItem, children []commands.TreeNode, indent string, maxDepth int) []MenuItem {
	if maxDepth <= 0 || len(children) == 0 {
		return items
	}
	for _, child := range children {
		childUUID := child.UUID
		items = append(items, MenuItem{
			Label: indent + "  * " + child.Title,
			OnRun: func() error {
				return c.openNoteByUUID(childUUID)
			},
		})
		items = c.appendTreeItems(items, child.Children, indent+"  ", maxDepth-1)
	}
	return items
}

// buildHeaderTOC returns menu items for the current card's headers.
// Each item gets an OnRun that moves the preview cursor to that header line.
func (c *PreviewController) buildHeaderTOC() []MenuItem {
	idx := c.gui.state.Preview.SelectedCardIndex
	if idx >= len(c.gui.state.Preview.CardLineRanges) {
		return nil
	}
	ranges := c.gui.state.Preview.CardLineRanges[idx]

	var viewLines []string
	if c.gui.views.Preview != nil {
		viewLines = c.gui.views.Preview.ViewBufferLines()
	}

	type header struct {
		level    int
		title    string
		viewLine int
	}
	var headers []header

	for _, hLine := range c.gui.state.Preview.HeaderLines {
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

	var items []MenuItem
	for _, h := range headers {
		depth := h.level - minLevel
		indent := strings.Repeat("  ", depth)
		targetLine := h.viewLine
		items = append(items, MenuItem{
			Label: indent + "* " + h.title,
			OnRun: func() error {
				c.gui.state.Preview.CursorLine = targetLine
				c.gui.renderPreview()
				return nil
			},
		})
	}
	return items
}
