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

	statusHeight := 2
	contentHeight := maxY - statusHeight

	// Sidebar panel heights - Notes 50%, Queries & Tags 25%
	notesHeight := contentHeight / 2
	queriesHeight := contentHeight / 4

	// TODO: Use this if tag height seems off / weird...
	// tagsHeight := contentHeight - notesHeight - queriesHeight

	if err := gui.createNotesView(g, 0, 0, sidebarWidth-1, notesHeight-1); err != nil {
		return err
	}

	if err := gui.createQueriesView(g, 0, notesHeight, sidebarWidth-1, notesHeight+queriesHeight-1); err != nil {
		return err
	}

	if err := gui.createTagsView(g, 0, notesHeight+queriesHeight, sidebarWidth-1, contentHeight-1); err != nil {
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

	if !gui.state.Initialized {
		gui.state.Initialized = true
		g.SetCurrentView(NotesView)
		gui.renderNotes()
		gui.renderTags()
		gui.renderQueries()
		gui.renderPreview()
	}

	return nil
}

func (gui *Gui) createNotesView(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(NotesView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Notes = v
	v.Title = " [1] Notes "
	v.Highlight = true
	v.SelBgColor = gocui.ColorBlue
	v.SelFgColor = gocui.ColorWhite

	if gui.state.CurrentContext == NotesContext {
		v.FgColor = gocui.ColorGreen
	} else {
		v.FgColor = gocui.ColorDefault
	}

	return nil
}

func (gui *Gui) createQueriesView(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(QueriesView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Queries = v
	v.Title = " [2] Queries "
	v.Highlight = true
	v.SelBgColor = gocui.ColorBlue
	v.SelFgColor = gocui.ColorWhite

	if gui.state.CurrentContext == QueriesContext {
		v.FgColor = gocui.ColorGreen
	} else {
		v.FgColor = gocui.ColorDefault
	}

	return nil
}

func (gui *Gui) createTagsView(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(TagsView, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gui.views.Tags = v
	v.Title = " [3] Tags "
	v.Highlight = true
	v.SelBgColor = gocui.ColorBlue
	v.SelFgColor = gocui.ColorWhite

	if gui.state.CurrentContext == TagsContext {
		v.FgColor = gocui.ColorGreen
	} else {
		v.FgColor = gocui.ColorDefault
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

	if gui.state.CurrentContext == PreviewContext {
		v.FgColor = gocui.ColorGreen
	} else {
		v.FgColor = gocui.ColorDefault
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
		hints = "[/] Search  [n] New  [Enter] Edit  [d] Delete  [Tab] Next  [?] Help"
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
