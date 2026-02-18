package gui

import (
	"time"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/config"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/controllers"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
	"github.com/muesli/termenv"
)

// Gui manages the terminal user interface.
type Gui struct {
	g              *gocui.Gui
	views          *Views
	state          *GuiState
	config         *config.Config
	ruinCmd        *commands.RuinCommand
	preview        *PreviewController
	stopBg         chan struct{}
	QuickCapture   bool // when true, open capture on start and quit on save
	darkBackground bool

	// New controller/context architecture (Phase 2+)
	contexts          *context.ContextTree
	notesController   *controllers.NotesController
	tagsController    *controllers.TagsController
	queriesController *controllers.QueriesController
	previewController *controllers.PreviewController
}

// NewGui creates a new Gui instance.
func NewGui(cfg *config.Config, ruinCmd *commands.RuinCommand) *Gui {
	gui := &Gui{
		config:   cfg,
		ruinCmd:  ruinCmd,
		views:    &Views{},
		state:    NewGuiState(),
		contexts: &context.ContextTree{},
	}
	gui.preview = NewPreviewController(gui)
	gui.setupNotesContext()
	gui.setupTagsContext()
	gui.setupQueriesContext()
	gui.setupPreviewContext()
	gui.setupSearchContext()
	gui.setupCaptureContext()
	gui.setupPickContext()
	gui.setupInputPopupContext()
	return gui
}

// setupNotesContext initializes the NotesContext and NotesController.
func (gui *Gui) setupNotesContext() {
	notesCtx := context.NewNotesContext(gui.renderNotes, gui.preview.updatePreviewForNotes)
	notesCtx.CurrentTab = context.NotesTab(gui.state.Notes.CurrentTab)
	gui.contexts.Notes = notesCtx

	gui.notesController = controllers.NewNotesController(controllers.NotesControllerOpts{
		GetContext: func() *context.NotesContext { return gui.contexts.Notes },
		OnViewInPreview: func(_ *models.Note) error {
			return gui.viewNoteInPreview(nil, nil)
		},
		OnEditNote: func(_ *models.Note) error {
			return gui.editNote(nil, nil)
		},
		OnDeleteNote: func(_ *models.Note) error {
			return gui.deleteNote(nil, nil)
		},
		OnCopyPath: func(_ *models.Note) error {
			return gui.copyNotePath(nil, nil)
		},
		OnAddTag: func(_ *models.Note) error {
			return gui.addGlobalTag(nil, nil)
		},
		OnRemoveTag: func(_ *models.Note) error {
			return gui.removeTag(nil, nil)
		},
		OnSetParent: func(_ *models.Note) error {
			return gui.setParentDialog(nil, nil)
		},
		OnRemoveParent: func(_ *models.Note) error {
			return gui.removeParent(nil, nil)
		},
		OnToggleBookmark: func(_ *models.Note) error {
			return gui.toggleBookmark(nil, nil)
		},
		OnShowInfo: func(_ *models.Note) error {
			return gui.preview.showInfoDialog(nil, nil)
		},
		OnClick: gui.notesClick,
		OnWheelDown: func(g *gocui.Gui, v *gocui.View) error {
			scrollViewport(gui.views.Notes, 3)
			return nil
		},
		OnWheelUp: func(g *gocui.Gui, v *gocui.View) error {
			scrollViewport(gui.views.Notes, -3)
			return nil
		},
	})

	controllers.AttachController(gui.notesController)
}

// setupTagsContext initializes the new TagsContext and TagsController.
func (gui *Gui) setupTagsContext() {
	tagsCtx := context.NewTagsContext(gui.renderTags, gui.updatePreviewForTags)

	// Initialize from existing state defaults
	tagsCtx.CurrentTab = context.TagsTab(gui.state.Tags.CurrentTab)

	gui.contexts.Tags = tagsCtx

	gui.tagsController = controllers.NewTagsController(controllers.TagsControllerOpts{
		GetContext: func() *context.TagsContext { return gui.contexts.Tags },
		OnFilterByTag: func(tag *models.Tag) error {
			return gui.filterByTag(nil, nil)
		},
		OnRenameTag: func(tag *models.Tag) error {
			return gui.renameTag(nil, nil)
		},
		OnDeleteTag: func(tag *models.Tag) error {
			return gui.deleteTag(nil, nil)
		},
		OnClick: gui.tagsClick,
		OnWheelDown: func(g *gocui.Gui, v *gocui.View) error {
			scrollViewport(gui.views.Tags, 3)
			return nil
		},
		OnWheelUp: func(g *gocui.Gui, v *gocui.View) error {
			scrollViewport(gui.views.Tags, -3)
			return nil
		},
	})

	// Attach controller to context
	controllers.AttachController(gui.tagsController)
}

// setupQueriesContext initializes the QueriesContext and QueriesController.
func (gui *Gui) setupQueriesContext() {
	queriesCtx := context.NewQueriesContext(
		gui.renderQueries, gui.updatePreviewForQueries,
		gui.renderQueries, gui.updatePreviewForParents,
	)
	queriesCtx.CurrentTab = context.QueriesTab(gui.state.Queries.CurrentTab)
	gui.contexts.Queries = queriesCtx

	gui.queriesController = controllers.NewQueriesController(controllers.QueriesControllerOpts{
		GetContext: func() *context.QueriesContext { return gui.contexts.Queries },
		OnRunQuery: func(query *models.Query) error {
			return gui.runQuery(nil, nil)
		},
		OnDeleteQuery: func(query *models.Query) error {
			return gui.deleteQuery(nil, nil)
		},
		OnViewParent: func(parent *models.ParentBookmark) error {
			return gui.viewParent(nil, nil)
		},
		OnDeleteParent: func(parent *models.ParentBookmark) error {
			return gui.deleteParent(nil, nil)
		},
		OnClick: gui.queriesClick,
		OnWheelDown: func(g *gocui.Gui, v *gocui.View) error {
			scrollViewport(gui.views.Queries, 3)
			return nil
		},
		OnWheelUp: func(g *gocui.Gui, v *gocui.View) error {
			scrollViewport(gui.views.Queries, -3)
			return nil
		},
	})

	controllers.AttachController(gui.queriesController)
}

// setupPreviewContext initializes the PreviewContext and PreviewController.
func (gui *Gui) setupPreviewContext() {
	previewCtx := context.NewPreviewContext()
	gui.contexts.Preview = previewCtx

	gui.previewController = controllers.NewPreviewController(controllers.PreviewControllerOpts{
		GetContext: func() *context.PreviewContext { return gui.contexts.Preview },

		// Navigation — delegates to existing preview_controller.go methods
		OnMoveDown:   func() error { return gui.preview.previewDown(nil, nil) },
		OnMoveUp:     func() error { return gui.preview.previewUp(nil, nil) },
		OnCardDown:   func() error { return gui.preview.previewCardDown(nil, nil) },
		OnCardUp:     func() error { return gui.preview.previewCardUp(nil, nil) },
		OnNextHeader: func() error { return gui.preview.previewNextHeader(nil, nil) },
		OnPrevHeader: func() error { return gui.preview.previewPrevHeader(nil, nil) },
		OnNextLink:   func() error { return gui.preview.highlightNextLink(nil, nil) },
		OnPrevLink:   func() error { return gui.preview.highlightPrevLink(nil, nil) },
		OnClick: func() error {
			v := gui.views.Preview
			return gui.preview.previewClick(nil, v)
		},

		// Card actions
		OnDeleteCard:        func() error { return gui.preview.deleteCardFromPreview(nil, nil) },
		OnOpenInEditor:      func() error { return gui.preview.openCardInEditor(nil, nil) },
		OnAppendDone:        func() error { return gui.preview.appendDone(nil, gui.views.Preview) },
		OnMoveCard:          func() error { return gui.preview.moveCardHandler(nil, nil) },
		OnMergeCard:         func() error { return gui.preview.mergeCardHandler(nil, nil) },
		OnToggleFrontmatter: func() error { return gui.preview.toggleFrontmatter(nil, nil) },
		OnViewOptions:       func() error { return gui.preview.viewOptionsDialog(nil, nil) },
		OnToggleInlineTag:   func() error { return gui.preview.toggleInlineTag(nil, gui.views.Preview) },
		OnToggleInlineDate:  func() error { return gui.preview.toggleInlineDate(nil, gui.views.Preview) },
		OnOpenLink:          func() error { return gui.preview.openLink(nil, nil) },
		OnToggleTodo:        func() error { return gui.preview.toggleTodo(nil, gui.views.Preview) },
		OnFocusNote:         func() error { return gui.preview.focusNoteFromPreview(nil, nil) },
		OnBack:              func() error { return gui.preview.previewBack(nil, nil) },
		OnNavBack:           func() error { return gui.preview.navBack(nil, nil) },
		OnNavForward:        func() error { return gui.preview.navForward(nil, nil) },

		// Note actions
		OnAddTag:         func() error { return gui.addGlobalTag(nil, nil) },
		OnRemoveTag:      func() error { return gui.removeTag(nil, nil) },
		OnSetParent:      func() error { return gui.setParentDialog(nil, nil) },
		OnRemoveParent:   func() error { return gui.removeParent(nil, nil) },
		OnToggleBookmark: func() error { return gui.toggleBookmark(nil, nil) },
		OnShowInfo:       func() error { return gui.preview.showInfoDialog(nil, nil) },

		// Palette-only
		OnToggleTitle:      func() error { return gui.preview.toggleTitle(nil, nil) },
		OnToggleGlobalTags: func() error { return gui.preview.toggleGlobalTags(nil, nil) },
		OnToggleMarkdown:   func() error { return gui.preview.toggleMarkdown(nil, nil) },
		OnOrderCards:       gui.preview.orderCards,
		OnShowNavHistory:   gui.preview.showNavHistory,
	})

	controllers.AttachController(gui.previewController)
}

// setupSearchContext initializes the SearchContext and SearchController.
func (gui *Gui) setupSearchContext() {
	searchCtx := context.NewSearchContext()
	gui.contexts.Search = searchCtx

	searchState := func() *CompletionState { return gui.state.SearchCompletion }
	ctrl := controllers.NewSearchController(controllers.SearchControllerOpts{
		GetContext: func() *context.SearchContext { return gui.contexts.Search },
		OnEnter: func() error {
			return gui.completionEnter(searchState, gui.searchTriggers, gui.executeSearch)(gui.g, gui.views.Search)
		},
		OnEsc: func() error {
			return gui.completionEsc(searchState, gui.cancelSearch)(gui.g, gui.views.Search)
		},
		OnTab: func() error {
			return gui.completionTab(searchState, gui.searchTriggers)(gui.g, gui.views.Search)
		},
	})
	controllers.AttachController(ctrl)
}

// setupCaptureContext initializes the CaptureContext and CaptureController.
func (gui *Gui) setupCaptureContext() {
	captureCtx := context.NewCaptureContext()
	gui.contexts.Capture = captureCtx

	ctrl := controllers.NewCaptureController(controllers.CaptureControllerOpts{
		GetContext: func() *context.CaptureContext { return gui.contexts.Capture },
		OnSubmit:   func() error { return gui.submitCapture(gui.g, gui.views.Capture) },
		OnEsc:      func() error { return gui.cancelCapture(gui.g, gui.views.Capture) },
		OnTab:      func() error { return gui.captureTab(gui.g, gui.views.Capture) },
	})
	controllers.AttachController(ctrl)
}

// setupPickContext initializes the PickContext and PickController.
func (gui *Gui) setupPickContext() {
	pickCtx := context.NewPickContext()
	gui.contexts.Pick = pickCtx

	pickState := func() *CompletionState { return gui.state.PickCompletion }
	ctrl := controllers.NewPickController(controllers.PickControllerOpts{
		GetContext: func() *context.PickContext { return gui.contexts.Pick },
		OnEnter: func() error {
			return gui.completionEnter(pickState, gui.pickTriggers, gui.executePick)(gui.g, gui.views.Pick)
		},
		OnEsc: func() error {
			return gui.completionEsc(pickState, gui.cancelPick)(gui.g, gui.views.Pick)
		},
		OnTab: func() error {
			return gui.completionTab(pickState, gui.pickTriggers)(gui.g, gui.views.Pick)
		},
		OnToggleAny: func() error { return gui.togglePickAny(gui.g, gui.views.Pick) },
	})
	controllers.AttachController(ctrl)
}

// setupInputPopupContext initializes the InputPopupContext and InputPopupController.
func (gui *Gui) setupInputPopupContext() {
	inputPopupCtx := context.NewInputPopupContext()
	gui.contexts.InputPopup = inputPopupCtx

	ctrl := controllers.NewInputPopupController(controllers.InputPopupControllerOpts{
		GetContext: func() *context.InputPopupContext { return gui.contexts.InputPopup },
		OnEnter: func() error {
			v, _ := gui.g.View(InputPopupView)
			return gui.inputPopupEnter(gui.g, v)
		},
		OnEsc: func() error { return gui.inputPopupEsc(gui.g, nil) },
		OnTab: func() error {
			v, _ := gui.g.View(InputPopupView)
			return gui.inputPopupTab(gui.g, v)
		},
	})
	controllers.AttachController(ctrl)
}

// Run starts the GUI event loop.
func (gui *Gui) Run() error {
	// Detect terminal background before gocui takes over the terminal.
	gui.darkBackground = termenv.HasDarkBackground()

	err := gui.runMainLoop()
	if err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (gui *Gui) runMainLoop() error {
	g, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputTrue,
	})
	if err != nil {
		return err
	}
	defer g.Close()

	gui.g = g
	gui.views = &Views{} // Reset views for fresh layout
	g.Mouse = true
	g.Cursor = false
	g.ShowListFooter = true
	g.SetManager(gocui.ManagerFunc(gui.layout))

	if err := gui.setupKeybindings(); err != nil {
		return err
	}

	gui.stopBg = make(chan struct{})
	go gui.backgroundRefresh()

	err = g.MainLoop()
	close(gui.stopBg)
	return err
}

// backgroundRefresh polls for external changes every 30 seconds.
func (gui *Gui) backgroundRefresh() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-gui.stopBg:
			return
		case <-ticker.C:
			gui.g.Update(func(g *gocui.Gui) error {
				gui.backgroundRefreshData()
				return nil
			})
		}
	}
}

// backgroundRefreshData reloads data without resetting focus, selection, or preview mode.
func (gui *Gui) backgroundRefreshData() {
	// Preserve selections
	cardIdx := gui.state.Preview.SelectedCardIndex

	gui.refreshNotes(true)
	gui.refreshTags(true)
	gui.refreshQueries(true)
	gui.refreshParents(true)

	if gui.state.Preview.SelectedCardIndex != cardIdx && cardIdx < len(gui.state.Preview.Cards) {
		gui.state.Preview.SelectedCardIndex = cardIdx
	}

	gui.renderPreview()
	gui.updateStatusBar()
}

// renderAll re-renders all views with current state (e.g. after resize)
func (gui *Gui) renderAll() {
	gui.renderNotes()
	gui.renderQueries()
	gui.renderTags()
	gui.renderPreview()
	gui.updateStatusBar()
}

func (gui *Gui) refreshAll() {
	gui.refreshNotes(false)
	gui.refreshTags(false)
	gui.refreshQueries(false)
	gui.refreshParents(false)
}

func (gui *Gui) refreshNotes(preserve bool) {
	gui.fetchNotesForCurrentTab(preserve)
}

// syncNotesToLegacy copies NotesContext state to legacy GuiState.Notes.
func (gui *Gui) syncNotesToLegacy() {
	notesCtx := gui.contexts.Notes
	gui.state.Notes.Items = notesCtx.Items
	gui.state.Notes.SelectedIndex = notesCtx.GetSelectedLineIdx()
	gui.state.Notes.CurrentTab = NotesTab(notesCtx.CurrentTab)
}

func (gui *Gui) refreshTags(preserve bool) {
	tagsCtx := gui.contexts.Tags
	prevID := tagsCtx.GetSelectedItemId()

	tags, err := gui.ruinCmd.Tags.List()
	if err != nil {
		return
	}
	tagsCtx.Items = tags

	if preserve && prevID != "" {
		if newIdx := tagsCtx.GetList().FindIndexById(prevID); newIdx >= 0 {
			tagsCtx.SetSelectedLineIdx(newIdx)
		}
	} else {
		tagsCtx.SetSelectedLineIdx(0)
	}
	tagsCtx.ClampSelection()

	// Keep legacy state in sync during hybrid period
	gui.syncTagsToLegacy()
	gui.renderTags()
}

// syncTagsToLegacy copies TagsContext state to legacy GuiState.Tags
// so that any un-migrated code reading gui.state.Tags still works.
func (gui *Gui) syncTagsToLegacy() {
	tagsCtx := gui.contexts.Tags
	gui.state.Tags.Items = tagsCtx.Items
	gui.state.Tags.SelectedIndex = tagsCtx.GetSelectedLineIdx()
	gui.state.Tags.CurrentTab = TagsTab(tagsCtx.CurrentTab)
}

func (gui *Gui) refreshQueries(preserve bool) {
	queriesCtx := gui.contexts.Queries
	prevID := queriesCtx.GetQueriesList().GetSelectedItemId()

	queries, err := gui.ruinCmd.Queries.List()
	if err != nil {
		return
	}
	queriesCtx.Queries = queries

	if preserve && prevID != "" {
		if newIdx := queriesCtx.GetQueriesList().FindIndexById(prevID); newIdx >= 0 {
			queriesCtx.QueriesTrait().SetSelectedLineIdx(newIdx)
		}
	} else {
		queriesCtx.QueriesTrait().SetSelectedLineIdx(0)
	}
	queriesCtx.QueriesTrait().ClampSelection()

	gui.syncQueriesToLegacy()
	gui.renderQueries()
}

func (gui *Gui) refreshParents(preserve bool) {
	queriesCtx := gui.contexts.Queries
	prevID := queriesCtx.GetParentsList().GetSelectedItemId()

	parents, err := gui.ruinCmd.Parent.List()
	if err != nil {
		return
	}
	queriesCtx.Parents = parents

	if preserve && prevID != "" {
		if newIdx := queriesCtx.GetParentsList().FindIndexById(prevID); newIdx >= 0 {
			queriesCtx.ParentsTrait().SetSelectedLineIdx(newIdx)
		}
	} else {
		queriesCtx.ParentsTrait().SetSelectedLineIdx(0)
	}
	queriesCtx.ParentsTrait().ClampSelection()

	gui.syncParentsToLegacy()
	gui.renderQueries()
}

// syncQueriesToLegacy copies QueriesContext queries state to legacy GuiState.Queries.
func (gui *Gui) syncQueriesToLegacy() {
	queriesCtx := gui.contexts.Queries
	gui.state.Queries.Items = queriesCtx.Queries
	gui.state.Queries.SelectedIndex = queriesCtx.QueriesTrait().GetSelectedLineIdx()
	gui.state.Queries.CurrentTab = QueriesTab(queriesCtx.CurrentTab)
}

// syncParentsToLegacy copies QueriesContext parents state to legacy GuiState.Parents.
func (gui *Gui) syncParentsToLegacy() {
	queriesCtx := gui.contexts.Queries
	gui.state.Parents.Items = queriesCtx.Parents
	gui.state.Parents.SelectedIndex = queriesCtx.ParentsTrait().GetSelectedLineIdx()
}

// activateContext sets focus and refreshes data for the given context.
func (gui *Gui) activateContext(ctx ContextKey) {
	viewName := gui.contextToView(ctx)
	gui.g.SetCurrentView(viewName)

	// Re-render lists to update highlight visibility
	gui.renderNotes()
	gui.renderQueries()
	gui.renderTags()

	// Refresh data (preserving selections) and update preview based on new context
	switch ctx {
	case NotesContext:
		gui.refreshNotes(true)
		gui.preview.updatePreviewForNotes()
	case QueriesContext:
		if gui.state.Queries.CurrentTab == QueriesTabParents {
			gui.refreshParents(true)
			gui.updatePreviewForParents()
		} else {
			gui.refreshQueries(true)
			gui.updatePreviewForQueries()
		}
	case TagsContext:
		gui.refreshTags(true)
		gui.updatePreviewForTags()
	case PreviewContext:
		gui.renderPreview()
	}

	gui.updateStatusBar()
}

// pushContext pushes a new context onto the stack and activates it.
func (gui *Gui) pushContext(ctx ContextKey) {
	gui.state.ContextStack = append(gui.state.ContextStack, ctx)
	gui.activateContext(ctx)
}

// popContext pops the top context and activates the one below it.
func (gui *Gui) popContext() {
	if len(gui.state.ContextStack) > 1 {
		gui.state.ContextStack = gui.state.ContextStack[:len(gui.state.ContextStack)-1]
	}
	gui.activateContext(gui.state.currentContext())
}

// replaceContext replaces the top of the stack (e.g., search→preview).
func (gui *Gui) replaceContext(ctx ContextKey) {
	if len(gui.state.ContextStack) > 0 {
		gui.state.ContextStack[len(gui.state.ContextStack)-1] = ctx
	} else {
		gui.state.ContextStack = []ContextKey{ctx}
	}
	gui.activateContext(ctx)
}

// setContext is a convenience that pushes a new context (legacy compatibility).
func (gui *Gui) setContext(ctx ContextKey) {
	gui.pushContext(ctx)
}

// openOverlay opens a modal overlay. Returns false if one is already active.
func (gui *Gui) openOverlay(overlay OverlayType) bool {
	if gui.state.ActiveOverlay != OverlayNone {
		return false
	}
	gui.state.ActiveOverlay = overlay
	return true
}

// closeOverlay closes the current modal overlay.
func (gui *Gui) closeOverlay() {
	gui.state.ActiveOverlay = OverlayNone
}

// overlayActive returns true when any overlay or dialog is open.
func (gui *Gui) overlayActive() bool {
	return gui.state.ActiveOverlay != OverlayNone ||
		(gui.state.Dialog != nil && gui.state.Dialog.Active)
}

func (gui *Gui) contextToView(ctx ContextKey) string {
	switch ctx {
	case NotesContext:
		return NotesView
	case QueriesContext:
		return QueriesView
	case TagsContext:
		return TagsView
	case PreviewContext:
		return PreviewView
	case SearchContext:
		return SearchView
	case SearchFilterContext:
		return SearchFilterView
	case CaptureContext:
		return CaptureView
	case PickContext:
		return PickView
	case PaletteContext:
		return PaletteView
	}
	return NotesView
}
