package gui

import (
	"fmt"
	"strings"
	"time"

	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/models"
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
	if gui.state.SearchQuery != "" {
		searchFilterHeight = 3
	}

	// Sidebar panel heights - Notes 50%, Queries & Tags 25%
	notesHeight := contentHeight / 2
	queriesHeight := contentHeight / 4

	// Show search filter pane if there's an active search
	if gui.state.SearchQuery != "" {
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

	// Manage overlay views based on ActiveOverlay
	switch gui.state.ActiveOverlay {
	case OverlaySearch:
		if err := gui.createSearchPopup(g, maxX, maxY); err != nil {
			return err
		}
	case OverlayCapture:
		if err := gui.createCapturePopup(g, maxX, maxY); err != nil {
			return err
		}
	case OverlayPick:
		if err := gui.createPickPopup(g, maxX, maxY); err != nil {
			return err
		}
	case OverlayInputPopup:
		if err := gui.createInputPopup(g, maxX, maxY); err != nil {
			return err
		}
	case OverlaySnippetEditor:
		if err := gui.createSnippetEditor(g, maxX, maxY); err != nil {
			return err
		}
	case OverlayPalette:
		if err := gui.createPalettePopup(g, maxX, maxY); err != nil {
			return err
		}
	case OverlayCalendar:
		if err := gui.createCalendarViews(g, maxX, maxY); err != nil {
			return err
		}
	case OverlayContrib:
		if err := gui.createContribViews(g, maxX, maxY); err != nil {
			return err
		}
	}
	// Delete views for inactive overlays
	if gui.state.ActiveOverlay != OverlaySearch {
		g.DeleteView(SearchView)
		g.DeleteView(SearchSuggestView)
	}
	if gui.state.ActiveOverlay != OverlayCapture {
		g.DeleteView(CaptureView)
		g.DeleteView(CaptureSuggestView)
		gui.views.Capture = nil
	}
	if gui.state.ActiveOverlay != OverlayPick {
		g.DeleteView(PickView)
		g.DeleteView(PickSuggestView)
		gui.views.Pick = nil
	}
	if gui.state.ActiveOverlay != OverlayInputPopup {
		g.DeleteView(InputPopupView)
		g.DeleteView(InputPopupSuggestView)
	}
	if gui.state.ActiveOverlay != OverlaySnippetEditor {
		g.DeleteView(SnippetNameView)
		g.DeleteView(SnippetExpansionView)
		g.DeleteView(SnippetSuggestView)
	}
	if gui.state.ActiveOverlay != OverlayPalette {
		g.DeleteView(PaletteView)
		g.DeleteView(PaletteListView)
		gui.views.Palette = nil
		gui.views.PaletteList = nil
	}
	if gui.state.ActiveOverlay != OverlayCalendar {
		g.DeleteView(CalendarGridView)
		g.DeleteView(CalendarInputView)
		g.DeleteView(CalendarNotesView)
	}
	if gui.state.ActiveOverlay != OverlayContrib {
		g.DeleteView(ContribGridView)
		g.DeleteView(ContribNotesView)
	}

	// Render any active dialogs
	if err := gui.renderDialogs(g, maxX, maxY); err != nil {
		return err
	}

	if !gui.state.Initialized {
		gui.state.Initialized = true
		gui.state.lastWidth = maxX
		gui.state.lastHeight = maxY
		g.SetCurrentView(NotesView)
		gui.refreshAll()
		gui.preview.updatePreviewForNotes()
		if gui.QuickCapture {
			gui.state.ActiveOverlay = OverlayCapture
			gui.state.CaptureCompletion = NewCompletionState()
			gui.state.ContextStack = []ContextKey{NotesContext, CaptureContext}
		}
	} else if maxX != gui.state.lastWidth || maxY != gui.state.lastHeight {
		gui.state.lastWidth = maxX
		gui.state.lastHeight = maxY
		gui.state.Preview.ScrollOffset = 0
		gocui.Screen.Clear()
		gui.renderAll()
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
	v.Footer = fmt.Sprintf("%d results", len(gui.state.Preview.Cards))
	setRoundedCorners(v)

	if gui.state.currentContext() == SearchFilterContext {
		v.FrameColor = gocui.ColorGreen
		v.TitleColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorYellow
		v.TitleColor = gocui.ColorYellow
	}

	v.Clear()
	fmt.Fprintf(v, " %s", gui.state.SearchQuery)

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
	gui.updateNotesTab()
	setRoundedCorners(v)

	// Notes uses manual multi-line highlighting in renderNotes()
	v.Highlight = false

	if gui.state.currentContext() == NotesContext {
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
	gui.updateQueriesTab()
	setRoundedCorners(v)

	if gui.state.currentContext() == QueriesContext {
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
	gui.updateTagsTab()
	setRoundedCorners(v)

	if gui.state.currentContext() == TagsContext {
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
	switch {
	case gui.state.Preview.Mode == PreviewModeCardList && len(gui.state.Preview.Cards) > 0:
		v.Title = "Preview"
		v.Footer = fmt.Sprintf("%d of %d", gui.state.Preview.SelectedCardIndex+1, len(gui.state.Preview.Cards))
	case gui.state.Preview.Mode == PreviewModePickResults && len(gui.state.Preview.PickResults) > 0:
		v.Title = " Pick: " + gui.state.PickQuery + " "
		v.Footer = fmt.Sprintf("%d of %d", gui.state.Preview.SelectedCardIndex+1, len(gui.state.Preview.PickResults))
	default:
		v.Footer = ""
		v.Title = " Preview "
	}

	if gui.state.currentContext() == PreviewContext {
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

	gui.updateStatusBar()

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
		state:      func() *CompletionState { return gui.state.SearchCompletion },
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
	if err := gui.renderSuggestionView(g, SearchSuggestView, gui.state.SearchCompletion, x0, y1, width); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) createInputPopup(g *gocui.Gui, maxX, maxY int) error {
	config := gui.state.InputPopupConfig
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
		state: func() *CompletionState { return gui.state.InputPopupCompletion },
		triggers: func() []CompletionTrigger {
			if c := gui.state.InputPopupConfig; c != nil && c.Triggers != nil {
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
	if !gui.state.InputPopupSeedDone && config.Seed != "" {
		gui.state.InputPopupSeedDone = true
		v.TextArea.TypeString(config.Seed)
		if config.Triggers != nil {
			gui.updateCompletion(v, config.Triggers(), gui.state.InputPopupCompletion)
		}
	}

	v.RenderTextArea()

	g.Cursor = true
	g.SetViewOnTop(InputPopupView)
	g.SetCurrentView(InputPopupView)

	// Render suggestion dropdown below
	if err := gui.renderSuggestionView(g, InputPopupSuggestView, gui.state.InputPopupCompletion, x0, y1, width); err != nil {
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
	if gui.state.CaptureCompletion.Active {
		_, cy := v.TextArea.GetCursorXY()
		suggestY = y0 + cy + 2 // position relative to cursor line
		if suggestY > y1-2 {
			suggestY = y1 // below the popup if cursor is near bottom
		}
	}
	if err := gui.renderSuggestionView(g, CaptureSuggestView, gui.state.CaptureCompletion, x0, suggestY, x1-x0); err != nil {
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
	v.Title = " Pick "
	v.Footer = gui.pickFooter()
	v.Editable = true
	v.Wrap = false
	v.Editor = &completionEditor{
		gui:        gui,
		state:      func() *CompletionState { return gui.state.PickCompletion },
		triggers:   gui.pickTriggers,
		drillFlags: 0,
	}
	setRoundedCorners(v)
	v.FrameColor = gocui.ColorGreen
	v.TitleColor = gocui.ColorGreen
	// Seed "#" on first open so tag completion appears immediately
	if gui.state.PickSeedHash {
		gui.state.PickSeedHash = false
		v.TextArea.TypeString("#")
		gui.updateCompletion(v, gui.pickTriggers(), gui.state.PickCompletion)
	}

	v.RenderTextArea()

	g.Cursor = true
	g.SetViewOnTop(PickView)
	g.SetCurrentView(PickView)

	if err := gui.renderSuggestionView(g, PickSuggestView, gui.state.PickCompletion, x0, y1, width); err != nil {
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
	if !gui.state.PaletteSeedDone {
		gui.state.PaletteSeedDone = true
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
	if gui.state.PickAnyMode {
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
	if gui.state.CaptureParent != nil {
		parentTitle = gui.state.CaptureParent.Title
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
		state:      func() *CompletionState { return gui.state.SnippetEditorCompletion },
		triggers:   gui.snippetExpansionTriggers,
		drillFlags: DrillParent | DrillWikiLink,
	}
	setRoundedCorners(ev)

	// Green frame on focused view, default on other
	if gui.state.SnippetEditorFocus == 0 {
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

	if gui.state.SnippetEditorFocus == 0 {
		g.SetCurrentView(SnippetNameView)
	} else {
		g.SetCurrentView(SnippetExpansionView)
	}

	// Suggestion dropdown below expansion view
	if err := gui.renderSuggestionView(g, SnippetSuggestView, gui.state.SnippetEditorCompletion, x0, ey1, width); err != nil {
		return err
	}

	return nil
}
