package gui

import (
	"fmt"
	"strings"
	"time"

	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if maxX < 40 || maxY < 10 {
		return nil
	}

	sidebarWidth := maxX / 3
	if sidebarWidth > 40 {
		sidebarWidth = 40
	}
	if sidebarWidth < 20 {
		sidebarWidth = 20
	}

	statusHeight := 3
	contentHeight := maxY - statusHeight

	// Search filter pane height (only shown when search is active)
	searchFilterHeight := 0
	if gui.contexts.Search.Query != "" {
		searchFilterHeight = 3
	}

	// Sidebar panel heights - Notes 50%, Queries & Tags 25%
	notesHeight := contentHeight / 2
	queriesHeight := contentHeight / 4

	// Show search filter pane if there's an active search
	if gui.contexts.Search.Query != "" {
		if err := gui.createSearchFilterView(g, 0, 0, sidebarWidth-1, searchFilterHeight-1); err != nil {
			return err
		}
	} else {
		g.DeleteView(SearchFilterView)
		gui.views.SearchFilter = nil
	}

	notesStartY := searchFilterHeight
	notesEndY := notesStartY + notesHeight - 1
	if err := gui.createNotesView(g, 0, notesStartY, sidebarWidth-1, notesEndY); err != nil {
		return err
	}

	queriesStartY := notesEndY + 1
	queriesEndY := queriesStartY + queriesHeight - 1
	if err := gui.createQueriesView(g, 0, queriesStartY, sidebarWidth-1, queriesEndY); err != nil {
		return err
	}

	tagsStartY := queriesEndY + 1
	if err := gui.createTagsView(g, 0, tagsStartY, sidebarWidth-1, contentHeight-1); err != nil {
		return err
	}

	if err := gui.createPreviewView(g, sidebarWidth, 0, maxX-1, contentHeight-1); err != nil {
		return err
	}

	if err := gui.createStatusView(g, 0, contentHeight, maxX-1, maxY-1); err != nil {
		return err
	}

	// Push capture context before overlay creation so the view exists on the
	// first layout call (SetStack after the switch was too late — the capture
	// popup wouldn't be created until a second event-triggered layout).
	if !gui.state.Initialized && gui.QuickCapture {
		gui.contexts.Capture.Completion = types.NewCompletionState()
		gui.contextMgr.Push(gui.contexts.Capture.GetKey())
	}

	// Manage overlay views based on the current context
	switch gui.contextMgr.Current() {
	case "search":
		if err := gui.createSearchPopup(g, maxX, maxY); err != nil {
			return err
		}
	case "capture":
		if err := gui.createCapturePopup(g, maxX, maxY); err != nil {
			return err
		}
	case "pick":
		if err := gui.createPickPopup(g, maxX, maxY); err != nil {
			return err
		}
	case "inputPopup":
		if err := gui.createInputPopup(g, maxX, maxY); err != nil {
			return err
		}
	case "snippetName":
		if err := gui.createSnippetEditor(g, maxX, maxY); err != nil {
			return err
		}
	case "palette":
		if err := gui.createPalettePopup(g, maxX, maxY); err != nil {
			return err
		}
	case "calendarGrid":
		if err := gui.createCalendarViews(g, maxX, maxY); err != nil {
			return err
		}
	case "contribGrid":
		if err := gui.createContribViews(g, maxX, maxY); err != nil {
			return err
		}
	case "pickDialog":
		if err := gui.createPickDialog(g, maxX, maxY); err != nil {
			return err
		}
	}
	// Delete views for inactive overlays
	ctx := gui.contextMgr.Current()
	if ctx != "search" {
		g.DeleteView(SearchView)
		g.DeleteView(SearchSuggestView)
	}
	if ctx != "capture" {
		g.DeleteView(CaptureView)
		g.DeleteView(CaptureSuggestView)
		gui.views.Capture = nil
	}
	if ctx != "pick" {
		g.DeleteView(PickView)
		g.DeleteView(PickSuggestView)
		gui.views.Pick = nil
	}
	if ctx != "inputPopup" {
		g.DeleteView(InputPopupView)
		g.DeleteView(InputPopupSuggestView)
	}
	if ctx != "snippetName" {
		g.DeleteView(SnippetNameView)
		g.DeleteView(SnippetExpansionView)
		g.DeleteView(SnippetSuggestView)
	}
	if ctx != "palette" {
		g.DeleteView(PaletteView)
		g.DeleteView(PaletteListView)
		gui.views.Palette = nil
		gui.views.PaletteList = nil
	}
	if ctx != "calendarGrid" {
		g.DeleteView(CalendarGridView)
		g.DeleteView(CalendarInputView)
		g.DeleteView(CalendarNotesView)
	}
	if ctx != "contribGrid" {
		g.DeleteView(ContribGridView)
		g.DeleteView(ContribNotesView)
	}
	if ctx != "pickDialog" {
		g.DeleteView(PickDialogView)
	}

	// Render any active dialogs
	if err := gui.renderDialogs(g, maxX, maxY); err != nil {
		return err
	}

	if !gui.state.Initialized {
		gui.state.Initialized = true
		gui.state.lastWidth = maxX
		gui.state.lastHeight = maxY
		if !gui.QuickCapture {
			g.SetCurrentView(NotesView)
		}
		gui.RefreshAll()
		gui.helpers.Preview().UpdatePreviewForNotes()
	} else if maxX != gui.state.lastWidth || maxY != gui.state.lastHeight {
		gui.state.lastWidth = maxX
		gui.state.lastHeight = maxY
		ns := gui.contexts.ActivePreview().NavState()

		// Save cursor identity before re-render (Lines will be rebuilt with new width)
		var savedUUID string
		var savedLineNum int
		if ns.CursorLine >= 0 && ns.CursorLine < len(ns.Lines) {
			src := ns.Lines[ns.CursorLine]
			savedUUID = src.UUID
			savedLineNum = src.LineNum
		}

		ns.ScrollOffset = 0
		gocui.Screen.Clear()
		gui.RenderAll()

		// Restore cursor to the same source line after re-render
		if savedUUID != "" && savedLineNum > 0 {
			for i, sl := range ns.Lines {
				if sl.UUID == savedUUID && sl.LineNum == savedLineNum {
					ns.CursorLine = i
					gui.RenderPreview()
					break
				}
			}
		}
	}

	return nil
}

// setRoundedCorners applies rounded corner frame characters to a view
func setRoundedCorners(v *gocui.View) {
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
}

func (gui *Gui) createSearchFilterView(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(SearchFilterView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.SearchFilter = v
	v.Title = "[0]-Search"
	v.Footer = fmt.Sprintf("%d results", len(gui.contexts.CardList.Cards))
	setRoundedCorners(v)

	if gui.contextMgr.Current() == "searchFilter" {
		v.FrameColor = gocui.ColorGreen
		v.TitleColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorYellow
		v.TitleColor = gocui.ColorYellow
	}

	v.Clear()
	fmt.Fprintf(v, " %s", gui.contexts.Search.Query)

	return nil
}

func (gui *Gui) createNotesView(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(NotesView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Notes = v
	v.TitlePrefix = "[1]"
	// v.Title = "[1]"
	v.Tabs = []string{"All", "Today", "Recent"}
	v.SelFgColor = gocui.ColorGreen
	gui.UpdateNotesTab()
	setRoundedCorners(v)

	// Notes uses manual multi-line highlighting in renderNotes()
	v.Highlight = false

	if gui.contextMgr.Current() == "notes" {
		v.FrameColor = gocui.ColorGreen
		v.TitleColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
		v.TitleColor = gocui.ColorDefault
	}

	return nil
}

func (gui *Gui) createQueriesView(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(QueriesView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Queries = v
	v.TitlePrefix = "[2]"
	v.Tabs = []string{"Queries", "Parents"}
	v.SelFgColor = gocui.ColorGreen
	v.Highlight = false
	gui.UpdateQueriesTab()
	setRoundedCorners(v)

	if gui.contextMgr.Current() == "queries" {
		v.FrameColor = gocui.ColorGreen
		v.TitleColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
		v.TitleColor = gocui.ColorDefault
	}

	return nil
}

func (gui *Gui) createTagsView(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(TagsView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Tags = v
	v.TitlePrefix = "[3]"
	v.Tabs = []string{"All", "Global", "Inline"}
	v.SelFgColor = gocui.ColorGreen
	v.Highlight = false
	gui.UpdateTagsTab()
	setRoundedCorners(v)

	if gui.contextMgr.Current() == "tags" {
		v.FrameColor = gocui.ColorGreen
		v.TitleColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
		v.TitleColor = gocui.ColorDefault
	}

	return nil
}

func (gui *Gui) createPreviewView(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(PreviewView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Preview = v
	v.Wrap = true
	setRoundedCorners(v)

	// Set title with card count for multi-card/pick mode
	switch gui.contexts.ActivePreviewKey {
	case "pickResults":
		pr := gui.contexts.PickResults
		query := gui.contexts.Pick.Query
		if query != "" {
			v.Title = " Pick: " + query + " "
		} else {
			v.Title = " Pick "
		}
		if len(pr.Results) > 0 {
			v.Footer = fmt.Sprintf("%d of %d", pr.SelectedCardIdx+1, len(pr.Results))
		} else {
			v.Footer = ""
		}
	case "compose":
		v.Footer = ""
	default:
		cl := gui.contexts.CardList
		if len(cl.Cards) > 0 {
			v.Title = "Preview"
			v.Footer = fmt.Sprintf("%d of %d", cl.SelectedCardIdx+1, len(cl.Cards))
		} else {
			v.Footer = ""
			v.Title = " Preview "
		}
	}

	if gui.isPreviewActive() {
		v.FrameColor = gocui.ColorGreen
		v.TitleColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
		v.TitleColor = gocui.ColorDefault
	}

	return nil
}

func (gui *Gui) createStatusView(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(StatusView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Status = v
	v.Frame = false

	gui.UpdateStatusBar()

	return nil
}

func (gui *Gui) createSearchPopup(g *gocui.Gui, maxX, maxY int) error {
	width := 60
	if width > maxX-4 {
		width = maxX - 4
	}
	height := 3

	x0 := (maxX - width) / 2
	y0 := (maxY-height)/2 - 2 // offset up to leave room for suggestions
	x1 := x0 + width
	y1 := y0 + height

	v, err := g.SetView(SearchView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Search = v
	v.Title = " Search "
	v.Footer = " / for filters | # for tags | Tab: complete | Esc: cancel "
	v.Editable = true
	v.Wrap = false
	v.Editor = &completionEditor{
		gui:        gui,
		state:      func() *types.CompletionState { return gui.contexts.Search.Completion },
		triggers:   gui.searchTriggers,
		drillFlags: 0,
	}
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen
	v.RenderTextArea() // ensure view has content so footer renders

	g.Cursor = true
	g.SetViewOnTop(SearchView)
	g.SetCurrentView(SearchView)

	// Render suggestion dropdown below the search popup
	if err := gui.renderSuggestionView(g, SearchSuggestView, gui.contexts.Search.Completion, x0, y1, width); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) createInputPopup(g *gocui.Gui, maxX, maxY int) error {
	config := gui.contexts.InputPopup.Config
	if config == nil {
		return nil
	}

	width := 60
	if width > maxX-4 {
		width = maxX - 4
	}
	height := 3

	x0 := (maxX - width) / 2
	y0 := (maxY-height)/2 - 2 // offset up to leave room for suggestions
	x1 := x0 + width
	y1 := y0 + height

	v, err := g.SetView(InputPopupView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Title = " " + config.Title + " "
	v.Footer = config.Footer
	v.Editable = true
	v.Wrap = false
	v.Editor = &completionEditor{
		gui:   gui,
		state: func() *types.CompletionState { return gui.contexts.InputPopup.Completion },
		triggers: func() []types.CompletionTrigger {
			if c := gui.contexts.InputPopup.Config; c != nil && c.Triggers != nil {
				return c.Triggers()
			}
			return nil
		},
		drillFlags: DrillParent,
	}
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen

	// Seed text on first open so completion appears immediately
	if !gui.contexts.InputPopup.SeedDone && config.Seed != "" {
		gui.contexts.InputPopup.SeedDone = true
		v.TextArea.TypeString(config.Seed)
		if config.Triggers != nil {
			gui.updateCompletion(v, config.Triggers(), gui.contexts.InputPopup.Completion)
		}
	}

	v.RenderTextArea()

	g.Cursor = true
	g.SetViewOnTop(InputPopupView)
	g.SetCurrentView(InputPopupView)

	// Render suggestion dropdown below
	if err := gui.renderSuggestionView(g, InputPopupSuggestView, gui.contexts.InputPopup.Completion, x0, y1, width); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) createCapturePopup(g *gocui.Gui, maxX, maxY int) error {
	var x0, y0, x1, y1 int
	if gui.QuickCapture {
		x0 = 0
		y0 = 0
		x1 = maxX - 1
		y1 = maxY - 1
	} else {
		width := 75
		if width > maxX-4 {
			width = maxX - 4
		}
		height := 25
		if height > maxY-4 {
			height = maxY - 4
		}
		x0 = (maxX - width) / 2
		y0 = (maxY - height) / 2
		x1 = x0 + width
		y1 = y0 + height
	}

	v, err := g.SetView(CaptureView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Capture = v
	v.Title = " New Note "
	v.Subtitle = " <c-s> to save "
	v.Editable = true
	v.TextArea.AutoWrap = true
	v.TextArea.AutoWrapWidth = v.InnerWidth() - 1
	v.Editor = &captureEditor{gui: gui}
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen
	gui.updateCaptureFooter()
	gui.renderCaptureTextArea(v) // render with syntax highlighting

	g.Cursor = true
	g.SetViewOnTop(CaptureView)
	g.SetCurrentView(CaptureView)

	// Render suggestion dropdown below the capture popup
	suggestY := y0 + 3 // position below a few lines into the popup
	if gui.contexts.Capture.Completion.Active {
		_, cy := v.TextArea.GetCursorXY()
		suggestY = y0 + cy + 2 // position relative to cursor line
		if suggestY > y1-2 {
			suggestY = y1 // below the popup if cursor is near bottom
		}
	}
	if err := gui.renderSuggestionView(g, CaptureSuggestView, gui.contexts.Capture.Completion, x0, suggestY, x1-x0); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) createPickPopup(g *gocui.Gui, maxX, maxY int) error {
	width := 60
	if width > maxX-4 {
		width = maxX - 4
	}
	height := 3

	x0 := (maxX - width) / 2
	y0 := (maxY-height)/2 - 2 // offset up to leave room for suggestions
	x1 := x0 + width
	y1 := y0 + height

	v, err := g.SetView(PickView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Pick = v
	if gui.contexts.Pick.ScopeTitle != "" {
		v.Title = " Pick from " + gui.contexts.Pick.ScopeTitle + " "
	} else {
		v.Title = " Pick "
	}
	v.Footer = gui.pickFooter()
	v.Editable = true
	v.Wrap = false
	v.Editor = &completionEditor{
		gui:        gui,
		state:      func() *types.CompletionState { return gui.contexts.Pick.Completion },
		triggers:   gui.pickTriggers,
		drillFlags: 0,
	}
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen
	// Seed "#" on first open so tag completion appears immediately
	if gui.contexts.Pick.SeedHash {
		gui.contexts.Pick.SeedHash = false
		v.TextArea.TypeString("#")
		gui.updateCompletion(v, gui.pickTriggers(), gui.contexts.Pick.Completion)
	}

	v.RenderTextArea()

	g.Cursor = true
	g.SetViewOnTop(PickView)
	g.SetCurrentView(PickView)

	if err := gui.renderSuggestionView(g, PickSuggestView, gui.contexts.Pick.Completion, x0, y1, width); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) createPalettePopup(g *gocui.Gui, maxX, maxY int) error {
	width := 60
	if width > maxX-4 {
		width = maxX - 4
	}

	// Input view (single line)
	inputHeight := 2
	// List view (up to 15 visible lines + 2 for border)
	listHeight := 17
	totalHeight := inputHeight + listHeight

	x0 := (maxX - width) / 2
	y0 := (maxY-totalHeight)/2 - 1
	if y0 < 0 {
		y0 = 0
	}
	x1 := x0 + width
	y1 := y0 + inputHeight

	// Input view
	v, err := g.SetView(PaletteView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}
	gui.views.Palette = v
	v.Editable = true
	v.Wrap = false
	v.Editor = &paletteEditor{gui: gui}
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen

	// Start in Command Palette mode; typing ":" switches to Quick Open
	if !gui.contexts.Palette.SeedDone {
		gui.contexts.Palette.SeedDone = true
		v.Title = " Command Palette "
		gui.filterPaletteCommands("")
	}

	v.RenderTextArea()

	g.Cursor = true
	g.SetViewOnTop(PaletteView)
	g.SetCurrentView(PaletteView)

	// List view below input
	ly0 := y1
	ly1 := ly0 + listHeight
	if ly1 >= maxY {
		ly1 = maxY - 1
	}

	lv, lvErr := g.SetView(PaletteListView, x0, ly0, x1, ly1, 0)
	if lvErr != nil && lvErr.Error() != "unknown view" {
		return lvErr
	}
	gui.views.PaletteList = lv
	lv.Wrap = false
	setRoundedCorners(lv)
	lv.FrameColor = gocui.ColorGreen

	g.SetViewOnTop(PaletteListView)

	// Only render on first creation; subsequent updates come from
	// the editor (filter changes) and paletteSelectMove (selection changes).
	// Re-rendering every layout cycle would clear the buffer and
	// reset the scrollable range, breaking mouse wheel scrolling.
	if lvErr != nil {
		gui.renderPaletteList()
		gui.scrollPaletteToSelection()
	}

	return nil
}

func (gui *Gui) pickFooter() string {
	anyLabel := "off"
	if gui.contexts.Pick.AnyMode {
		anyLabel = "on"
	}
	return " # for tags | --any: " + anyLabel + " | <c-a>: toggle | Tab: complete | Esc: cancel "
}

// updateCaptureFooter sets the capture popup footer to show date, tags, and parent.
func (gui *Gui) updateCaptureFooter() {
	if gui.views.Capture == nil {
		return
	}

	date := time.Now().Format("Jan 02")

	// Extract inline tags from current capture content
	content := gui.views.Capture.TextArea.GetContent()
	tagMatches := inlineTagRe.FindAllString(content, -1)
	seen := make(map[string]bool)
	var tags []string
	for _, t := range tagMatches {
		if !seen[t] {
			seen[t] = true
			tags = append(tags, t)
		}
	}

	var tagsStr string
	if len(tags) > 0 {
		tagsStr = strings.Join(tags, ", ")
	}
	var parentTitle string
	if gui.contexts.Capture.Parent != nil {
		parentTitle = gui.contexts.Capture.Parent.Title
	}

	footer := " " + models.JoinDot(date, tagsStr, parentTitle) + " "
	maxLen := gui.views.Capture.InnerWidth()
	if len([]rune(footer)) > maxLen && maxLen > 4 {
		runes := []rune(footer)
		footer = string(runes[:maxLen-1]) + "…"
	}
	gui.views.Capture.Footer = footer
}

func (gui *Gui) createSnippetEditor(g *gocui.Gui, maxX, maxY int) error {
	width := 60
	if width > maxX-4 {
		width = maxX - 4
	}

	nameHeight := 3
	expansionHeight := 8
	totalHeight := nameHeight + expansionHeight

	x0 := (maxX - width) / 2
	y0 := (maxY-totalHeight)/2 - 2
	if y0 < 0 {
		y0 = 0
	}
	x1 := x0 + width

	// Name field (top)
	ny1 := y0 + nameHeight
	nv, nErr := g.SetView(SnippetNameView, x0, y0, x1, ny1, 0)
	if nErr != nil && nErr.Error() != "unknown view" {
		return nErr
	}
	nv.Title = " Snippet name "
	nv.Editable = true
	nv.Wrap = false
	nv.Editor = gocui.EditorFunc(gocui.SimpleEditor)
	setRoundedCorners(nv)

	// Expansion field (bottom, directly below name)
	ey0 := ny1
	ey1 := ey0 + expansionHeight
	ev, eErr := g.SetView(SnippetExpansionView, x0, ey0, x1, ey1, 0)
	if eErr != nil && eErr.Error() != "unknown view" {
		return eErr
	}
	ev.Title = " Expansion "
	ev.Footer = " # > [[ / Tab: switch | Enter: save "
	ev.Editable = true
	ev.Wrap = false
	ev.Editor = &completionEditor{
		gui:        gui,
		state:      func() *types.CompletionState { return gui.contexts.SnippetEditor.Completion },
		triggers:   gui.snippetExpansionTriggers,
		drillFlags: DrillParent | DrillWikiLink,
	}
	setRoundedCorners(ev)

	// Green frame on focused view, default on other
	if gui.contexts.SnippetEditor.Focus == 0 {
		nv.FrameColor = gocui.ColorGreen
		nv.TitleColor = gocui.ColorGreen
		ev.FrameColor = gocui.ColorDefault
		ev.TitleColor = gocui.ColorDefault
	} else {
		nv.FrameColor = gocui.ColorDefault
		nv.TitleColor = gocui.ColorDefault
		ev.FrameColor = gocui.ColorGreen
		ev.TitleColor = gocui.ColorGreen
	}

	nv.RenderTextArea()
	ev.RenderTextArea()

	g.Cursor = true
	g.SetViewOnTop(SnippetNameView)
	g.SetViewOnTop(SnippetExpansionView)

	if gui.contexts.SnippetEditor.Focus == 0 {
		g.SetCurrentView(SnippetNameView)
	} else {
		g.SetCurrentView(SnippetExpansionView)
	}

	// Suggestion dropdown below expansion view
	if err := gui.renderSuggestionView(g, SnippetSuggestView, gui.contexts.SnippetEditor.Completion, x0, ey1, width); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) createPickDialog(g *gocui.Gui, maxX, maxY int) error {
	width := maxX * 85 / 100
	if width < 40 {
		width = min(maxX-4, 40)
	}
	height := maxY * 75 / 100
	if height < 10 {
		height = min(maxY-4, 10)
	}

	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

	v, err := g.SetView(PickDialogView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	pd := gui.contexts.PickDialog
	title := " Pick: " + pd.Query + " "
	if pd.ScopeTitle != "" {
		title = " Pick from " + pd.ScopeTitle + " "
	}
	v.Title = title
	v.Wrap = false
	v.Frame = true
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen

	g.SetViewOnTop(PickDialogView)
	g.SetCurrentView(PickDialogView)

	gui.RenderPickDialog()
	return nil
}
