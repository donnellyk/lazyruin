package gui

import (
	"fmt"

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

	if gui.state.SearchMode {
		if err := gui.createSearchPopup(g, maxX, maxY); err != nil {
			return err
		}
	} else {
		g.DeleteView(SearchView)
		g.DeleteView(SearchSuggestView)
	}

	if gui.state.CaptureMode {
		if err := gui.createCapturePopup(g, maxX, maxY); err != nil {
			return err
		}
	} else {
		g.DeleteView(CaptureView)
		g.DeleteView(CaptureSuggestView)
		gui.views.Capture = nil
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
		gui.renderPreview()
		if gui.QuickCapture {
			gui.state.CaptureMode = true
			gui.state.CaptureCompletion = NewCompletionState()
			gui.state.CurrentContext = CaptureContext
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

	if gui.state.CurrentContext == SearchFilterContext {
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

	if gui.state.CurrentContext == NotesContext {
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

	if gui.state.CurrentContext == QueriesContext {
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
	v.Title = "[3]-Tags"
	v.SelBgColor = gocui.ColorBlue
	v.SelFgColor = gocui.ColorWhite
	setRoundedCorners(v)

	if gui.state.CurrentContext == TagsContext {
		v.FrameColor = gocui.ColorGreen
		v.TitleColor = gocui.ColorGreen
		v.Highlight = true
	} else {
		v.FrameColor = gocui.ColorDefault
		v.TitleColor = gocui.ColorDefault
		v.Highlight = false
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

	// Set title with card count for multi-card mode
	if gui.state.Preview.Mode == PreviewModeCardList && len(gui.state.Preview.Cards) > 0 {
		v.Title = "Preview"
		v.Footer = fmt.Sprintf("%d of %d", gui.state.Preview.SelectedCardIndex+1, len(gui.state.Preview.Cards))
	} else {
		v.Footer = ""
		v.Title = " Preview "
	}

	if gui.state.CurrentContext == PreviewContext {
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
	v.Editor = &searchEditor{gui: gui}
	setRoundedCorners(v)
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

func (gui *Gui) createCapturePopup(g *gocui.Gui, maxX, maxY int) error {
	width := 60
	if width > maxX-4 {
		width = maxX - 4
	}
	height := 15
	if height > maxY-4 {
		height = maxY - 4
	}

	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

	v, err := g.SetView(CaptureView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Capture = v
	v.Title = " New Note "
	v.Footer = " Ctrl+S: save | Esc: cancel | # for tags "
	v.Editable = true
	v.Wrap = true
	v.Editor = &captureEditor{gui: gui}
	setRoundedCorners(v)
	v.RenderTextArea() // ensure view has content so footer renders

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
	if err := gui.renderSuggestionView(g, CaptureSuggestView, gui.state.CaptureCompletion, x0, suggestY, width); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) updateStatusBar() {
	if gui.views.Status == nil {
		return
	}

	gui.views.Status.Clear()

	type hint struct {
		action string
		key    string
	}

	var hints []hint
	switch gui.state.CurrentContext {
	case NotesContext:
		hints = []hint{
			{"Edit", "e/enter"},
			{"Edit Mode", "E"},
			{"New", "n"},
			{"Delete", "d"},
			{"Search", "/"},
			{"Tab", "1"},
			{"Copy Path", "y"},
			{"Keybindings", "?"},
		}
	case QueriesContext:
		if gui.state.Queries.CurrentTab == QueriesTabParents {
			hints = []hint{
				{"View", "enter"},
				{"Delete", "d"},
				{"Tab", "2"},
				{"Keybindings", "?"},
			}
		} else {
			hints = []hint{
				{"Run", "enter"},
				{"Delete", "d"},
				{"Tab", "2"},
				{"Keybindings", "?"},
			}
		}
	case TagsContext:
		hints = []hint{
			{"Filter", "enter"},
			{"Rename", "r"},
			{"Delete", "d"},
			{"Keybindings", "?"},
		}
	case PreviewContext:
		if gui.state.Preview.EditMode {
			hints = []hint{
				{"Delete", "d"},
				{"Move Up", "K"},
				{"Move Down", "J"},
				{"Merge", "m"},
				{"Navigate", "j/k"},
				{"Back", "esc"},
			}
		} else {
			hints = []hint{
				{"Navigate", "j/k"},
				{"Focus Note", "enter"},
				{"Frontmatter", "f"},
				{"Back", "esc"},
				{"Keybindings", "?"},
			}
		}
	case SearchContext:
		hints = []hint{
			{"Search", "enter"},
			{"Complete", "tab"},
			{"Cancel", "esc"},
		}
	case CaptureContext:
		hints = []hint{
			{"Save", "ctrl+s"},
			{"Complete", "tab"},
			{"Cancel", "esc"},
		}
	case SearchFilterContext:
		hints = []hint{
			{"Clear", "x"},
			{"Keybindings", "?"},
		}
	default:
		hints = []hint{
			{"Quit", "q"},
			{"Keybindings", "?"},
		}
	}

	cyan := AnsiCyan
	reset := AnsiReset
	for i, h := range hints {
		if i > 0 {
			fmt.Fprint(gui.views.Status, " | ")
		}
		fmt.Fprintf(gui.views.Status, "%s: %s%s%s", h.action, cyan, h.key, reset)
	}
}

// notesTabIndex returns the index for the current tab
func (gui *Gui) notesTabIndex() int {
	switch gui.state.Notes.CurrentTab {
	case NotesTabToday:
		return 1
	case NotesTabRecent:
		return 2
	default:
		return 0
	}
}

// notesTabs maps tab indices to NotesTab values
var notesTabs = []NotesTab{NotesTabAll, NotesTabToday, NotesTabRecent}

// updateNotesTab syncs the gocui view's TabIndex with the current tab
func (gui *Gui) updateNotesTab() {
	if gui.views.Notes != nil {
		gui.views.Notes.TabIndex = gui.notesTabIndex()
	}
}

// queriesTabIndex returns the index for the current queries tab
func (gui *Gui) queriesTabIndex() int {
	switch gui.state.Queries.CurrentTab {
	case QueriesTabParents:
		return 1
	default:
		return 0
	}
}

// queriesTabs maps tab indices to QueriesTab values
var queriesTabs = []QueriesTab{QueriesTabQueries, QueriesTabParents}

// updateQueriesTab syncs the gocui view's TabIndex with the current queries tab
func (gui *Gui) updateQueriesTab() {
	if gui.views.Queries != nil {
		gui.views.Queries.TabIndex = gui.queriesTabIndex()
	}
}
