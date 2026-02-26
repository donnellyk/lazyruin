package helpers

import (
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
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

// ShowInfoDialog shows parent structure / TOC for the current card.
func (self *PreviewInfoHelper) ShowInfoDialog() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}

	var items []types.MenuItem
	items = append(items, types.MenuItem{Label: "Info: " + card.Title, Hint: "(c to copy)", IsHeader: true})

	// Frontmatter fields with copy-to-clipboard support.
	type fm struct {
		key, value string
	}
	vaultPath := self.c.RuinCmd().VaultPath()
	var fields []fm
	fields = append(fields, fm{"UUID", card.UUID})
	if !card.Created.IsZero() {
		fields = append(fields, fm{"Created", card.Created.Format("Jan 02, 2006 3:04 PM")})
	}
	if !card.Updated.IsZero() && !card.Updated.Equal(card.Created) {
		fields = append(fields, fm{"Updated", card.Updated.Format("Jan 02, 2006 3:04 PM")})
	}
	if tags := card.GlobalTagsString(); tags != "" {
		fields = append(fields, fm{"Tags", tags})
	}
	if len(card.InlineTags) > 0 {
		inline := strings.Join(card.InlineTags, ", ")
		fields = append(fields, fm{"Inline", inline})
	}
	if card.Path != "" {
		rel := strings.TrimPrefix(card.Path, vaultPath+"/")
		fields = append(fields, fm{"Path", rel})
	}
	if card.Order != nil {
		fields = append(fields, fm{"Order", fmt.Sprintf("%d", *card.Order)})
	}
	for _, f := range fields {
		val := f.value
		items = append(items, types.MenuItem{
			Label:       fmt.Sprintf("%-8s  %s", f.key, val),
			KeepOpenKey: "c",
			OnRun: func() error {
				return self.c.Helpers().Clipboard().CopyToClipboard(val)
			},
		})
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
		items = append(items, types.MenuItem{Label: tree.Title, OnRun: func() error {
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

func (self *PreviewInfoHelper) appendTreeItems(items []types.MenuItem, children []commands.TreeNode, prefix string, maxDepth int) []types.MenuItem {
	if maxDepth <= 0 || len(children) == 0 {
		return items
	}
	for i, child := range children {
		connector, childPrefix := treeConnector(prefix, i == len(children)-1)
		childUUID := child.UUID
		items = append(items, types.MenuItem{
			Label: connector + child.Title,
			OnRun: func() error {
				return self.c.Helpers().PreviewNav().OpenNoteByUUID(childUUID)
			},
		})
		items = self.appendTreeItems(items, child.Children, childPrefix, maxDepth-1)
	}
	return items
}

// treeConnector returns the box-drawing prefix for a node and the
// continuation prefix for its children.
func treeConnector(prefix string, isLast bool) (connector, childPrefix string) {
	if isLast {
		return prefix + "└─ ", prefix + "   "
	}
	return prefix + "├─ ", prefix + "│  "
}

// depthLabel pairs a depth level with a display label.
type depthLabel struct {
	depth int
	label string
}

// depthTreePrefixes converts a flat list of depth/label pairs into
// prefixed strings using box-drawing characters. Items at depth 0 get
// no prefix (they are roots); deeper items get ├─/└─ connectors with
// │ continuation lines from open ancestor levels.
func depthTreePrefixes(items []depthLabel) []string {
	if len(items) == 0 {
		return nil
	}
	maxDepth := 0
	for _, item := range items {
		if item.depth > maxDepth {
			maxDepth = item.depth
		}
	}
	open := make([]bool, maxDepth+1)
	result := make([]string, len(items))

	for i, item := range items {
		d := item.depth

		// Is this the last sibling at its depth?
		isLast := true
		for j := i + 1; j < len(items); j++ {
			if items[j].depth <= d {
				isLast = items[j].depth < d
				break
			}
		}

		open[d] = !isLast
		for k := d + 1; k <= maxDepth; k++ {
			open[k] = false
		}

		var sb strings.Builder
		// Start at 1: depth-0 items are roots with no connector,
		// so they produce no continuation column.
		for k := 1; k < d; k++ {
			if open[k] {
				sb.WriteString("│  ")
			} else {
				sb.WriteString("   ")
			}
		}
		if d > 0 {
			if isLast {
				sb.WriteString("└─ ")
			} else {
				sb.WriteString("├─ ")
			}
		}
		sb.WriteString(item.label)
		result[i] = sb.String()
	}
	return result
}

func (self *PreviewInfoHelper) buildHeaderTOC() []types.MenuItem {
	ctx := self.activeCtx()
	ns := ctx.NavState()
	idx := ctx.SelectedCardIndex()
	if idx >= len(ns.CardLineRanges) {
		return nil
	}
	ranges := ns.CardLineRanges[idx]

	// Use ns.Lines (indexed in sync with ns.HeaderLines) instead of
	// ViewBufferLines(), which returns gocui's wrapped display lines and
	// can have shifted indices when Wrap=true.
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
		if hLine < len(ns.Lines) {
			raw := strings.TrimSpace(stripAnsi(ns.Lines[hLine].Text))
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

	// Convert to depth/label pairs and compute prefixes.
	depthLabels := make([]depthLabel, len(headers))
	for i, h := range headers {
		depthLabels[i] = depthLabel{depth: h.level - minLevel, label: h.title}
	}
	prefixed := depthTreePrefixes(depthLabels)

	var items []types.MenuItem
	for i, pl := range prefixed {
		targetLine := headers[i].viewLine
		items = append(items, types.MenuItem{
			Label: pl,
			OnRun: func() error {
				ns.CursorLine = targetLine
				self.c.GuiCommon().RenderPreview()
				return nil
			},
		})
	}
	return items
}
