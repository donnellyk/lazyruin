package gui

import (
	"strings"
	"unicode"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

type statusBarEntry struct {
	key    string
	action string
}

func bindingKeyDisplay(b *types.Binding) string {
	if b.KeyDisplay != "" {
		return b.KeyDisplay
	}
	if b.Key != nil {
		return keyDisplayString(b.Key)
	}
	return ""
}

var searchFilterHints = []statusBarEntry{
	{key: "x", action: "Clear"},
	{key: "?", action: "Keys"},
}

func (gui *Gui) statusBarHints() []statusBarEntry {
	ctxKey := gui.contextMgr.Current()

	if ctxKey == "searchFilter" {
		return searchFilterHints
	}

	ctx := gui.contextMgr.ContextByKey(ctxKey)
	if ctx == nil {
		return nil
	}

	kind := ctx.GetKind()
	isPopup := kind == types.PERSISTENT_POPUP || kind == types.TEMPORARY_POPUP

	opts := types.KeybindingsOpts{}
	var entries []statusBarEntry

	for _, b := range ctx.GetKeybindings(opts) {
		if isPopup {
			if b.Description == "" {
				continue
			}
		} else if !b.DisplayOnScreen {
			continue
		}
		label := b.StatusBarLabel
		if label == "" {
			label = b.Description
		}
		if label == "" {
			continue
		}
		entries = append(entries, statusBarEntry{
			key:    bindingKeyDisplay(b),
			action: label,
		})
	}

	if !isPopup {
		globalCtx := gui.contextMgr.ContextByKey("global")
		if globalCtx != nil && string(ctxKey) != "global" {
			for _, b := range globalCtx.GetKeybindings(opts) {
				if !b.DisplayOnScreen {
					continue
				}
				label := b.StatusBarLabel
				if label == "" {
					label = b.Description
				}
				if label == "" {
					continue
				}
				entries = append(entries, statusBarEntry{
					key:    bindingKeyDisplay(b),
					action: label,
				})
			}
		}
	}

	return entries
}

var extraHints = map[types.ContextKey][]types.MenuItem{
	"capture": {
		{Key: "/", Label: "Formatting"},
		{Key: ">", Label: "Parent"},
	},
}

func (gui *Gui) contextDisplayName(key types.ContextKey) string {
	switch key {
	case "queries":
		if gui.contexts.Queries.CurrentTab == context.QueriesTabParents {
			return "Parents"
		}
		return "Queries"
	case "cardList":
		return "Preview"
	case "pickResults":
		return "Pick Results"
	case "compose":
		return "Compose"
	case "searchFilter":
		return "Search Filter"
	case "pickDialog":
		return "Pick Results"
	default:
		s := string(key)
		if len(s) == 0 {
			return s
		}
		return string(unicode.ToUpper(rune(s[0]))) + s[1:]
	}
}

type categoryGroup struct {
	name  string
	items []types.MenuItem
}

func (gui *Gui) helpDialogItems() []types.MenuItem {
	ctxKey := gui.contextMgr.Current()
	ctx := gui.contextMgr.ContextByKey(ctxKey)

	opts := types.KeybindingsOpts{}
	var contextGroups []categoryGroup
	var navGroup categoryGroup

	if ctx != nil {
		contextGroups, navGroup = collectBindingGroups(ctx.GetKeybindings(opts))
	}

	var items []types.MenuItem

	if len(contextGroups) > 0 {
		displayName := gui.contextDisplayName(ctxKey)
		if len(contextGroups) == 1 {
			items = append(items, types.MenuItem{Label: displayName, IsHeader: true})
			items = append(items, contextGroups[0].items...)
		} else {
			for _, g := range contextGroups {
				items = append(items, types.MenuItem{Label: g.name, IsHeader: true})
				items = append(items, g.items...)
			}
		}
	}

	if extra, ok := extraHints[ctxKey]; ok {
		items = append(items, extra...)
	}

	items = append(items, types.MenuItem{})

	globalCtx := gui.contextMgr.ContextByKey("global")
	if globalCtx != nil {
		globalGroups, _ := collectBindingGroups(globalCtx.GetKeybindings(opts))
		for _, g := range globalGroups {
			items = append(items, types.MenuItem{Label: g.name, IsHeader: true})
			items = append(items, g.items...)
		}
	}

	if len(navGroup.items) > 0 {
		items = append(items, types.MenuItem{})
		items = append(items, types.MenuItem{Label: "Navigation", IsHeader: true})
		items = append(items, navGroup.items...)
	}

	return items
}

func collectBindingGroups(bindings []*types.Binding) (groups []categoryGroup, nav categoryGroup) {
	categoryOrder := []string{}
	categoryMap := map[string][]types.MenuItem{}

	for _, b := range bindings {
		if b.Description == "" {
			continue
		}
		item := types.MenuItem{
			Key:   bindingKeyDisplay(b),
			Label: b.Description,
		}
		cat := b.Category
		if cat == "" {
			cat = "Other"
		}

		if strings.EqualFold(cat, "Navigation") {
			nav.items = append(nav.items, item)
			continue
		}

		if _, exists := categoryMap[cat]; !exists {
			categoryOrder = append(categoryOrder, cat)
		}
		categoryMap[cat] = append(categoryMap[cat], item)
	}

	for _, cat := range categoryOrder {
		groups = append(groups, categoryGroup{name: cat, items: categoryMap[cat]})
	}
	return groups, nav
}
