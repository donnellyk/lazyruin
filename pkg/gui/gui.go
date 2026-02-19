package gui

import (
	"strings"
	"time"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/config"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/controllers"
	helperspkg "kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"
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
	stopBg         chan struct{}
	QuickCapture   bool // when true, open capture on start and quit on save
	darkBackground bool

	// New controller/context architecture (Phase 2+)
	contexts          *context.ContextTree
	contextMgr        *ContextMgr
	notesController   *controllers.NotesController
	tagsController    *controllers.TagsController
	queriesController *controllers.QueriesController
	previewController *controllers.PreviewController
	globalController  *controllers.GlobalController

	// Shared helper/controller dependencies
	helpers          *helperspkg.Helpers
	controllerCommon *controllers.ControllerCommon
}

// NewGui creates a new Gui instance.
func NewGui(cfg *config.Config, ruinCmd *commands.RuinCommand) *Gui {
	gui := &Gui{
		config:     cfg,
		ruinCmd:    ruinCmd,
		views:      &Views{},
		state:      NewGuiState(),
		contexts:   &context.ContextTree{},
		contextMgr: NewContextMgr(),
	}
	// Wire shared helper/controller dependencies.
	helperCommon := helperspkg.NewHelperCommon(ruinCmd, gui.config, gui)
	gui.helpers = helperspkg.NewHelpers(helperCommon)
	gui.controllerCommon = controllers.NewControllerCommon(gui, ruinCmd, gui.helpers)

	gui.setupNotesContext()
	gui.setupTagsContext()
	gui.setupQueriesContext()
	gui.setupPreviewContext()
	gui.setupSearchContext()
	gui.setupCaptureContext()
	gui.setupPickContext()
	gui.setupInputPopupContext()
	gui.setupGlobalContext()
	gui.setupPaletteContext()
	gui.setupSnippetEditorContext()
	gui.setupCalendarContext()
	gui.setupContribContext()
	return gui
}

// setupNotesContext initializes the "notes" and NotesController.
func (gui *Gui) setupNotesContext() {
	notesCtx := context.NewNotesContext(gui.RenderNotes, func() { gui.helpers.Preview().UpdatePreviewForNotes() })
	gui.contexts.Notes = notesCtx
	gui.contextMgr.Register(notesCtx)

	notesCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		gui.RefreshNotes(true)
		gui.helpers.Preview().UpdatePreviewForNotes()
	})

	gui.notesController = controllers.NewNotesController(controllers.NotesControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.NotesContext { return gui.contexts.Notes },
		OnShowInfo: func(_ *models.Note) error {
			return gui.helpers.PreviewInfo().ShowInfoDialog()
		},
	})

	controllers.AttachController(gui.notesController)
}

// setupTagsContext initializes the new "tags" and TagsController.
func (gui *Gui) setupTagsContext() {
	tagsCtx := context.NewTagsContext(gui.RenderTags, func() { gui.helpers.Tags().UpdatePreviewForTags() })

	gui.contexts.Tags = tagsCtx
	gui.contextMgr.Register(tagsCtx)

	tagsCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		gui.RefreshTags(true)
		gui.helpers.Tags().UpdatePreviewForTags()
	})

	gui.tagsController = controllers.NewTagsController(controllers.TagsControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.TagsContext { return gui.contexts.Tags },
	})

	controllers.AttachController(gui.tagsController)
}

// setupQueriesContext initializes the "queries" and QueriesController.
func (gui *Gui) setupQueriesContext() {
	queriesCtx := context.NewQueriesContext(
		gui.RenderQueries, func() { gui.helpers.Queries().UpdatePreviewForQueries() },
		gui.RenderQueries, func() { gui.helpers.Queries().UpdatePreviewForParents() },
	)
	gui.contexts.Queries = queriesCtx
	gui.contextMgr.Register(queriesCtx)

	queriesCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		if gui.contexts.Queries.CurrentTab == context.QueriesTabParents {
			gui.RefreshParents(true)
			gui.helpers.Queries().UpdatePreviewForParents()
		} else {
			gui.RefreshQueries(true)
			gui.helpers.Queries().UpdatePreviewForQueries()
		}
	})

	gui.queriesController = controllers.NewQueriesController(controllers.QueriesControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.QueriesContext { return gui.contexts.Queries },
	})

	controllers.AttachController(gui.queriesController)
}

// setupPreviewContext initializes the "preview" and PreviewController.
func (gui *Gui) setupPreviewContext() {
	previewCtx := context.NewPreviewContext()
	gui.contexts.Preview = previewCtx
	gui.contextMgr.Register(previewCtx)

	previewCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		gui.RenderPreview()
	})

	gui.previewController = controllers.NewPreviewController(
		gui.controllerCommon,
		func() *context.PreviewContext { return gui.contexts.Preview },
	)

	controllers.AttachController(gui.previewController)
}

// setupSearchContext initializes the "search" and its popup controller.
func (gui *Gui) setupSearchContext() {
	searchCtx := context.NewSearchContext()
	gui.contexts.Search = searchCtx
	gui.contextMgr.Register(searchCtx)

	searchState := func() *types.CompletionState { return gui.state.SearchCompletion }
	searchHelper := func() *helperspkg.SearchHelper { return gui.helpers.Search() }
	ctrl := controllers.NewPopupController(
		func() *context.SearchContext { return gui.contexts.Search },
		[]*types.Binding{
			{Key: gocui.KeyEnter, Handler: func() error {
				return gui.completionEnter(searchState, gui.searchTriggers, func(g *gocui.Gui, v *gocui.View) error {
					raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
					if !searchHelper().ExecuteSearch(raw) {
						searchHelper().CancelSearch()
					}
					return nil
				})(gui.g, gui.views.Search)
			}},
			{Key: gocui.KeyEsc, Handler: func() error {
				return gui.completionEsc(searchState, func(g *gocui.Gui, v *gocui.View) error {
					searchHelper().CancelSearch()
					return nil
				})(gui.g, gui.views.Search)
			}},
			{Key: gocui.KeyTab, Handler: func() error {
				return gui.completionTab(searchState, gui.searchTriggers)(gui.g, gui.views.Search)
			}},
		},
	)
	controllers.AttachController(ctrl)
}

// setupCaptureContext initializes the "capture" and its popup controller.
func (gui *Gui) setupCaptureContext() {
	captureCtx := context.NewCaptureContext()
	gui.contexts.Capture = captureCtx
	gui.contextMgr.Register(captureCtx)

	ctrl := controllers.NewPopupController(
		func() *context.CaptureContext { return gui.contexts.Capture },
		[]*types.Binding{
			{Key: gocui.KeyCtrlS, Handler: func() error {
				content := strings.TrimSpace(gui.views.Capture.TextArea.GetUnwrappedContent())
				return gui.helpers.Capture().SubmitCapture(content, gui.QuickCapture)
			}},
			{Key: gocui.KeyEsc, Handler: func() error { return gui.helpers.Capture().CancelCapture(gui.QuickCapture) }},
			{Key: gocui.KeyTab, Handler: func() error { return gui.captureTab(gui.g, gui.views.Capture) }},
		},
	)
	controllers.AttachController(ctrl)
}

// setupPickContext initializes the "pick" and its popup controller.
func (gui *Gui) setupPickContext() {
	pickCtx := context.NewPickContext()
	gui.contexts.Pick = pickCtx
	gui.contextMgr.Register(pickCtx)

	pickState := func() *types.CompletionState { return gui.contexts.Pick.Completion }
	ctrl := controllers.NewPopupController(
		func() *context.PickContext { return gui.contexts.Pick },
		[]*types.Binding{
			{Key: gocui.KeyEnter, Handler: func() error {
				executePick := func(g *gocui.Gui, v *gocui.View) error {
					raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
					return gui.helpers.Pick().ExecutePick(raw)
				}
				return gui.completionEnter(pickState, gui.pickTriggers, executePick)(gui.g, gui.views.Pick)
			}},
			{Key: gocui.KeyEsc, Handler: func() error {
				cancelPick := func(g *gocui.Gui, v *gocui.View) error {
					return gui.helpers.Pick().CancelPick()
				}
				return gui.completionEsc(pickState, cancelPick)(gui.g, gui.views.Pick)
			}},
			{Key: gocui.KeyTab, Handler: func() error {
				return gui.completionTab(pickState, gui.pickTriggers)(gui.g, gui.views.Pick)
			}},
			{Key: gocui.KeyCtrlA, Handler: func() error {
				gui.helpers.Pick().TogglePickAny()
				if gui.views.Pick != nil {
					gui.views.Pick.Footer = gui.pickFooter()
				}
				return nil
			}},
		},
	)
	controllers.AttachController(ctrl)
}

// setupInputPopupContext initializes the InputPopupContext and its popup controller.
func (gui *Gui) setupInputPopupContext() {
	inputPopupCtx := context.NewInputPopupContext()
	gui.contexts.InputPopup = inputPopupCtx
	gui.contextMgr.Register(inputPopupCtx)

	ctrl := controllers.NewPopupController(
		func() *context.InputPopupContext { return gui.contexts.InputPopup },
		[]*types.Binding{
			{Key: gocui.KeyEnter, Handler: func() error {
				v, _ := gui.g.View(InputPopupView)
				raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
				state := gui.contexts.InputPopup.Completion
				var item *types.CompletionItem
				if state.Active && len(state.Items) > 0 {
					selected := state.Items[state.SelectedIndex]
					item = &selected
				}
				return gui.helpers.InputPopup().HandleEnter(raw, item)
			}},
			{Key: gocui.KeyEsc, Handler: func() error { return gui.helpers.InputPopup().HandleEsc() }},
			{Key: gocui.KeyTab, Handler: func() error {
				ctx := gui.contexts.InputPopup
				if ctx.Completion.Active && len(ctx.Completion.Items) > 0 {
					v, _ := gui.g.View(InputPopupView)
					raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
					selected := ctx.Completion.Items[ctx.Completion.SelectedIndex]
					return gui.helpers.InputPopup().HandleEnter(raw, &selected)
				}
				return nil
			}},
		},
	)
	controllers.AttachController(ctrl)
}

// setupGlobalContext initializes the GlobalContext and GlobalController.
func (gui *Gui) setupGlobalContext() {
	globalCtx := context.NewGlobalContext()
	gui.contexts.Global = globalCtx
	gui.contextMgr.Register(globalCtx)

	ctrl := controllers.NewGlobalController(controllers.GlobalControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.GlobalContext { return gui.contexts.Global },
		OnQuit:     func() error { return gui.quit(gui.g, nil) },
		OnPick:     func() error { return gui.helpers.Pick().OpenPick() },
		OnNewNote:  func() error { return gui.helpers.Capture().OpenCapture() },
		OnHelp:     func() error { gui.showHelp(); return nil },
		OnPalette:  func() error { return gui.openPalette(gui.g, nil) },
		OnCalendar: func() error { return gui.openCalendar(gui.g, nil) },
		OnContrib:  func() error { return gui.openContrib(gui.g, nil) },
	})
	gui.globalController = ctrl
	controllers.AttachController(ctrl)
}

// setupPaletteContext initializes the "palette" and PaletteController.
func (gui *Gui) setupPaletteContext() {
	paletteCtx := context.NewPaletteContext()
	gui.contexts.Palette = paletteCtx
	gui.contextMgr.Register(paletteCtx)

	ctrl := controllers.NewPaletteController(controllers.PaletteControllerOpts{
		GetContext:  func() *context.PaletteContext { return gui.contexts.Palette },
		OnEnter:     func() error { return gui.paletteEnter(gui.g, nil) },
		OnEsc:       func() error { return gui.paletteEsc(gui.g, nil) },
		OnListClick: func() error { return gui.paletteListClick(gui.g, nil) },
	})
	controllers.AttachController(ctrl)
}

// setupSnippetEditorContext initializes the SnippetEditorContext and SnippetEditorController.
func (gui *Gui) setupSnippetEditorContext() {
	snippetCtx := context.NewSnippetEditorContext()
	gui.contexts.SnippetEditor = snippetCtx
	gui.contextMgr.Register(snippetCtx)

	acceptExpansionCompletion := func() {
		ctx := gui.contexts.SnippetEditor
		ev, _ := gui.g.View(SnippetExpansionView)
		if ev != nil {
			if isParentCompletion(ev, ctx.Completion) {
				gui.acceptSnippetParentCompletion(ev, ctx.Completion)
			} else {
				gui.acceptCompletion(ev, ctx.Completion, gui.snippetExpansionTriggers())
			}
			ev.RenderTextArea()
		}
	}

	ctrl := controllers.NewSnippetEditorController(controllers.SnippetEditorControllerOpts{
		GetContext: func() *context.SnippetEditorContext { return gui.contexts.SnippetEditor },
		OnEsc: func() error {
			ctx := gui.contexts.SnippetEditor
			if ctx.Completion.Active {
				ctx.Completion.Dismiss()
				return nil
			}
			return gui.helpers.Snippet().CloseEditor()
		},
		OnTab: func() error {
			ctx := gui.contexts.SnippetEditor
			if ctx.Completion.Active && ctx.Focus == 1 {
				acceptExpansionCompletion()
				return nil
			}
			if ctx.Focus == 0 {
				ctx.Focus = 1
			} else {
				ctx.Focus = 0
			}
			return nil
		},
		OnEnterName: func() error {
			ctx := gui.contexts.SnippetEditor
			if ctx.Completion.Active && ctx.Focus == 1 {
				acceptExpansionCompletion()
				return nil
			}
			if ctx.Focus == 0 {
				ctx.Focus = 1
			} else {
				ctx.Focus = 0
			}
			return nil
		},
		OnEnterExpansion: func() error {
			ctx := gui.contexts.SnippetEditor
			if ctx.Completion.Active {
				acceptExpansionCompletion()
				return nil
			}
			nv, _ := gui.g.View(SnippetNameView)
			ev, _ := gui.g.View(SnippetExpansionView)
			if nv == nil || ev == nil {
				return nil
			}
			name := strings.TrimLeft(strings.TrimSpace(nv.TextArea.GetUnwrappedContent()), "!")
			expansion := strings.TrimSpace(ev.TextArea.GetUnwrappedContent())
			if err := gui.helpers.Snippet().SaveSnippet(name, expansion); err != nil {
				return err
			}
			return gui.helpers.Snippet().CloseEditor()
		},
		OnClickName:      func() error { gui.contexts.SnippetEditor.Focus = 0; return nil },
		OnClickExpansion: func() error { gui.contexts.SnippetEditor.Focus = 1; return nil },
	})
	controllers.AttachController(ctrl)
}

// setupCalendarContext initializes the CalendarContext and CalendarController.
func (gui *Gui) setupCalendarContext() {
	calendarCtx := context.NewCalendarContext()
	gui.contexts.Calendar = calendarCtx
	gui.contextMgr.Register(calendarCtx)

	ctrl := controllers.NewCalendarController(controllers.CalendarControllerOpts{
		GetContext:   func() *context.CalendarContext { return gui.contexts.Calendar },
		OnGridLeft:   func() error { return gui.calendarGridLeft(nil, nil) },
		OnGridRight:  func() error { return gui.calendarGridRight(nil, nil) },
		OnGridUp:     func() error { return gui.calendarGridUp(nil, nil) },
		OnGridDown:   func() error { return gui.calendarGridDown(nil, nil) },
		OnGridEnter:  func() error { return gui.calendarGridEnter(nil, nil) },
		OnEsc:        func() error { return gui.calendarEsc(nil, nil) },
		OnTab:        func() error { return gui.calendarTab(nil, nil) },
		OnBacktab:    func() error { return gui.calendarBacktab(nil, nil) },
		OnFocusInput: func() error { return gui.calendarFocusInput(nil, nil) },
		OnGridClick: func() error {
			v, _ := gui.g.View(CalendarGridView)
			return gui.calendarGridClick(nil, v)
		},
		OnInputEnter: func() error {
			v, _ := gui.g.View(CalendarInputView)
			return gui.calendarInputEnter(nil, v)
		},
		OnInputEsc: func() error {
			v, _ := gui.g.View(CalendarInputView)
			return gui.calendarInputEsc(nil, v)
		},
		OnInputClick: func() error {
			v, _ := gui.g.View(CalendarInputView)
			return gui.calendarInputClick(nil, v)
		},
		OnNoteDown:  func() error { return gui.calendarNoteDown(nil, nil) },
		OnNoteUp:    func() error { return gui.calendarNoteUp(nil, nil) },
		OnNoteEnter: func() error { return gui.calendarNoteEnter(nil, nil) },
	})
	controllers.AttachController(ctrl)
}

// setupContribContext initializes the ContribContext and ContribController.
func (gui *Gui) setupContribContext() {
	contribCtx := context.NewContribContext()
	gui.contexts.Contrib = contribCtx
	gui.contextMgr.Register(contribCtx)

	ctrl := controllers.NewContribController(controllers.ContribControllerOpts{
		GetContext:  func() *context.ContribContext { return gui.contexts.Contrib },
		OnGridLeft:  func() error { return gui.contribGridLeft(nil, nil) },
		OnGridRight: func() error { return gui.contribGridRight(nil, nil) },
		OnGridUp:    func() error { return gui.contribGridUp(nil, nil) },
		OnGridDown:  func() error { return gui.contribGridDown(nil, nil) },
		OnGridEnter: func() error { return gui.contribGridEnter(nil, nil) },
		OnEsc:       func() error { return gui.contribEsc(nil, nil) },
		OnTab:       func() error { return gui.contribTab(nil, nil) },
		OnNoteDown:  func() error { return gui.contribNoteDown(nil, nil) },
		OnNoteUp:    func() error { return gui.contribNoteUp(nil, nil) },
		OnNoteEnter: func() error { return gui.contribNoteEnter(nil, nil) },
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
	cardIdx := gui.contexts.Preview.SelectedCardIndex

	gui.RefreshNotes(true)
	gui.RefreshTags(true)
	gui.RefreshQueries(true)
	gui.RefreshParents(true)

	if gui.contexts.Preview.SelectedCardIndex != cardIdx && cardIdx < len(gui.contexts.Preview.Cards) {
		gui.contexts.Preview.SelectedCardIndex = cardIdx
	}

	gui.RenderPreview()
	gui.UpdateStatusBar()
}

// activateContext sets focus and re-renders lists for the given context.
// Per-context refresh/preview logic is handled by HandleFocus hooks.
func (gui *Gui) activateContext(ctx types.ContextKey) {
	viewName := gui.contextToView(ctx)
	gui.g.SetCurrentView(viewName)

	// Re-render lists to update highlight visibility
	gui.RenderNotes()
	gui.RenderQueries()
	gui.RenderTags()

	gui.UpdateStatusBar()
}

// pushContext pushes a new context onto the stack and activates it.
func (gui *Gui) pushContext(ctx types.Context) {
	if cur := gui.currentContextObject(); cur != nil {
		cur.HandleFocusLost(types.OnFocusLostOpts{})
	}
	gui.contextMgr.Push(ctx.GetKey())
	gui.activateContext(ctx.GetKey())
	ctx.HandleFocus(types.OnFocusOpts{})
}

// popContext pops the top context and activates the one below it.
func (gui *Gui) popContext() {
	if cur := gui.currentContextObject(); cur != nil {
		cur.HandleFocusLost(types.OnFocusLostOpts{})
	}
	gui.contextMgr.Pop()
	gui.activateContext(gui.contextMgr.Current())
	if next := gui.currentContextObject(); next != nil {
		next.HandleFocus(types.OnFocusOpts{})
	}
}

// replaceContext replaces the top of the stack (e.g., search→preview).
func (gui *Gui) replaceContext(ctx types.Context) {
	if cur := gui.currentContextObject(); cur != nil {
		cur.HandleFocusLost(types.OnFocusLostOpts{})
	}
	gui.contextMgr.Replace(ctx.GetKey())
	gui.activateContext(ctx.GetKey())
	ctx.HandleFocus(types.OnFocusOpts{})
}

// pushContextByKey looks up the context by key and pushes it.
// Falls back to a direct stack push for lightweight contexts not in the tree.
func (gui *Gui) pushContextByKey(key types.ContextKey) {
	ctx := gui.contextMgr.ContextByKey(key)
	if ctx != nil {
		gui.pushContext(ctx)
		return
	}
	// Lightweight context (e.g., "searchFilter"): push key directly.
	gui.contextMgr.Push(key)
	gui.activateContext(key)
}

// replaceContextByKey looks up the context by key and replaces the top of the stack.
// Falls back to a direct stack replace for lightweight contexts not in the tree.
func (gui *Gui) replaceContextByKey(key types.ContextKey) {
	ctx := gui.contextMgr.ContextByKey(key)
	if ctx != nil {
		gui.replaceContext(ctx)
		return
	}
	// Lightweight context: replace key directly.
	gui.contextMgr.Replace(key)
	gui.activateContext(key)
}

// currentContextObject looks up the types.Context for the top of stack.
func (gui *Gui) currentContextObject() types.Context {
	return gui.contextMgr.ContextByKey(gui.contextMgr.Current())
}

// popupActive returns true when the current context is a popup (not a main panel).
func (gui *Gui) popupActive() bool {
	ctx := gui.contextMgr.ContextByKey(gui.contextMgr.Current())
	if ctx == nil {
		return false
	}
	kind := ctx.GetKind()
	return kind != types.SIDE_CONTEXT && kind != types.MAIN_CONTEXT
}

// overlayActive returns true when any overlay or dialog is open.
func (gui *Gui) overlayActive() bool {
	return gui.popupActive() ||
		(gui.state.Dialog != nil && gui.state.Dialog.Active)
}

func (gui *Gui) contextToView(ctx types.ContextKey) string {
	return gui.contexts.ViewNameForKey(ctx)
}
