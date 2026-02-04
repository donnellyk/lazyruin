package gui

import (
	"fmt"
	"strings"

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

	statusHeight := 2
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
		g.DeleteView(SearchView) // ignore error if view doesn't exist
	}

	// Render any active dialogs
	if err := gui.renderDialogs(g, maxX, maxY); err != nil {
		return err
	}

	if !gui.state.Initialized {
		gui.state.Initialized = true
		g.SetCurrentView(NotesView)
		gui.refreshAll()
		gui.renderPreview()
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
	v.Title = gui.getNotesTitle()
	v.SelBgColor = gocui.ColorBlue
	v.SelFgColor = gocui.ColorWhite
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
	v.Title = "[2]-Queries"
	v.SelBgColor = gocui.ColorBlue
	v.SelFgColor = gocui.ColorWhite
	setRoundedCorners(v)

	if gui.state.CurrentContext == QueriesContext {
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
	v.Title = " Preview "
	v.Wrap = true
	setRoundedCorners(v)

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
	height := 15

	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

	v, err := g.SetView(SearchView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Search = v
	v.Title = " Search "
	v.Editable = true
	v.Wrap = false
	setRoundedCorners(v)

	g.SetViewOnTop(SearchView)
	g.SetCurrentView(SearchView)

	return nil
}

func (gui *Gui) updateStatusBar() {
	if gui.views.Status == nil {
		return
	}

	gui.views.Status.Clear()

	var hints string
	switch gui.state.CurrentContext {
	case NotesContext:
		hints = "[1] Cycle Tab  [/] Search  [n] New  [Enter] Edit  [d] Delete  [?] Help"
	case QueriesContext:
		hints = "[Enter] Run  [e] Edit  [d] Delete  [n] New  [Tab] Next  [?] Help"
	case TagsContext:
		hints = "[Enter] Filter  [r] Rename  [d] Delete  [Tab] Next  [?] Help"
	case PreviewContext:
		hints = "[j/k] Scroll  [Enter] Focus  [f] Frontmatter  [Esc] Back  [?] Help"
	case SearchContext:
		hints = "[Enter] Search  [Esc] Cancel  [Tab] Autocomplete"
	default:
		hints = "[q] Quit  [?] Help"
	}

	fmt.Fprint(gui.views.Status, hints)
}

// getNotesTitle returns the title for the Notes view with tab indicator
// Selected tab is marked with brackets, entire title colored via TitleColor
func (gui *Gui) getNotesTitle() string {
	tabs := []struct {
		tab  NotesTab
		name string
	}{
		{NotesTabAll, "All"},
		{NotesTabToday, "Today"},
		{NotesTabRecent, "Recent"},
	}

	var parts []string
	for _, t := range tabs {
		if t.tab == gui.state.Notes.CurrentTab {
			parts = append(parts, "["+t.name+"]")
		} else {
			parts = append(parts, t.name)
		}
	}

	return "[1]-" + strings.Join(parts, "-")
}

// updateNotesTitle updates the Notes view title to reflect current tab
func (gui *Gui) updateNotesTitle() {
	if gui.views.Notes != nil {
		gui.views.Notes.Title = gui.getNotesTitle()
	}
}
