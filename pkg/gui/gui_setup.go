package gui

import (
	"strings"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/controllers"
	helperspkg "kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

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

// setupPreviewContext initializes the three preview contexts (cardList,
// pickResults, compose) that share a single nav history.
func (gui *Gui) setupPreviewContext() {
	gui.contexts.ActivePreviewKey = "cardList"

	// Shared nav history across all three preview contexts
	navHistory := context.NewSharedNavHistory()

	// CardList context
	cardListCtx := context.NewCardListContext(navHistory)
	gui.contexts.CardList = cardListCtx
	gui.contextMgr.Register(cardListCtx)
	cardListCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		gui.RenderPreview()
	})
	cardListCtrl := controllers.NewCardListController(
		gui.controllerCommon,
		func() *context.CardListContext { return gui.contexts.CardList },
	)
	controllers.AttachController(cardListCtrl)

	// PickResults context
	pickResultsCtx := context.NewPickResultsContext(navHistory)
	gui.contexts.PickResults = pickResultsCtx
	gui.contextMgr.Register(pickResultsCtx)
	pickResultsCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		gui.RenderPreview()
	})
	pickResultsCtrl := controllers.NewPickResultsController(
		gui.controllerCommon,
		func() *context.PickResultsContext { return gui.contexts.PickResults },
	)
	controllers.AttachController(pickResultsCtrl)

	// Compose context
	composeCtx := context.NewComposeContext(navHistory)
	gui.contexts.Compose = composeCtx
	gui.contextMgr.Register(composeCtx)
	composeCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		gui.RenderPreview()
	})
	composeCtrl := controllers.NewComposeController(
		gui.controllerCommon,
		func() *context.ComposeContext { return gui.contexts.Compose },
	)
	controllers.AttachController(composeCtrl)

	// DatePreview context
	datePreviewCtx := context.NewDatePreviewContext(navHistory)
	gui.contexts.DatePreview = datePreviewCtx
	gui.contextMgr.Register(datePreviewCtx)
	datePreviewCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
		gui.RenderPreview()
	})
	datePreviewCtrl := controllers.NewDatePreviewController(
		gui.controllerCommon,
		func() *context.DatePreviewContext { return gui.contexts.DatePreview },
	)
	controllers.AttachController(datePreviewCtrl)
}

// setupSearchContext initializes the "search" and its popup controller.
func (gui *Gui) setupSearchContext() {
	searchCtx := context.NewSearchContext()
	gui.contexts.Search = searchCtx
	gui.contextMgr.Register(searchCtx)

	searchState := func() *types.CompletionState { return gui.contexts.Search.Completion }
	searchHelper := func() *helperspkg.SearchHelper { return gui.helpers.Search() }
	ctrl := controllers.NewPopupController(
		func() *context.SearchContext { return gui.contexts.Search },
		[]*types.Binding{
			{Key: gocui.KeyEnter, Description: "Search", Handler: func() error {
				return gui.completionEnter(searchState, gui.searchTriggers, func(g *gocui.Gui, v *gocui.View) error {
					raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
					if !searchHelper().ExecuteSearch(raw) {
						searchHelper().CancelSearch()
					}
					return nil
				})(gui.g, gui.views.Search)
			}},
			{Key: gocui.KeyEsc, Description: "Cancel", Handler: func() error {
				return gui.completionEsc(searchState, func(g *gocui.Gui, v *gocui.View) error {
					searchHelper().CancelSearch()
					return nil
				})(gui.g, gui.views.Search)
			}},
			{Key: gocui.KeyTab, Description: "Complete", Handler: func() error {
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
			{Key: gocui.KeyCtrlS, Description: "Save", Handler: func() error {
				content := strings.TrimSpace(gui.views.Capture.TextArea.GetUnwrappedContent())
				return gui.helpers.Capture().SubmitCapture(content, gui.QuickCapture)
			}},
			{Key: gocui.KeyEsc, Description: "Cancel", Handler: func() error { return gui.helpers.Capture().CancelCapture(gui.QuickCapture) }},
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
			{Key: gocui.KeyEnter, Description: "Pick", Handler: func() error {
				executePick := func(g *gocui.Gui, v *gocui.View) error {
					raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
					return gui.helpers.Pick().ExecutePick(raw)
				}
				return gui.completionEnter(pickState, gui.pickTriggers, executePick)(gui.g, gui.views.Pick)
			}},
			{Key: gocui.KeyEsc, Description: "Cancel", Handler: func() error {
				cancelPick := func(g *gocui.Gui, v *gocui.View) error {
					return gui.helpers.Pick().CancelPick()
				}
				return gui.completionEsc(pickState, cancelPick)(gui.g, gui.views.Pick)
			}},
			{Key: gocui.KeyTab, Description: "Complete", Handler: func() error {
				return gui.completionTab(pickState, gui.pickTriggers)(gui.g, gui.views.Pick)
			}},
			{Key: gocui.KeyCtrlA, Description: "--any", Handler: func() error {
				gui.helpers.Pick().TogglePickAny()
				if gui.views.Pick != nil {
					gui.views.Pick.Footer = gui.pickFooter()
				}
				return nil
			}},
			{Key: gocui.KeyCtrlT, Description: "--todo", Handler: func() error {
				gui.helpers.Pick().TogglePickTodo()
				if gui.views.Pick != nil {
					gui.views.Pick.Footer = gui.pickFooter()
				}
				return nil
			}},
			{Key: gocui.KeyCtrlL, Description: "--all-tags", Handler: func() error {
				gui.helpers.Pick().TogglePickAllTags()
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
		OnHelp:     func() error { gui.showHelp(); return nil },
		OnPalette:  func() error { return gui.openPalette(gui.g, nil) },
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
		Common:     gui.controllerCommon,
		GetContext: func() *context.CalendarContext { return gui.contexts.Calendar },
	})
	controllers.AttachController(ctrl)
}

// setupPickDialogContext initializes the pick dialog context and controller.
func (gui *Gui) setupPickDialogContext() {
	pdCtx := context.NewPickDialogContext()
	gui.contexts.PickDialog = pdCtx
	gui.contextMgr.Register(pdCtx)

	ctrl := controllers.NewPickDialogController(gui.controllerCommon, func() *context.PickResultsContext {
		return gui.contexts.PickDialog
	})
	controllers.AttachController(ctrl)
}

// setupContribContext initializes the ContribContext and ContribController.
func (gui *Gui) setupContribContext() {
	contribCtx := context.NewContribContext()
	gui.contexts.Contrib = contribCtx
	gui.contextMgr.Register(contribCtx)

	ctrl := controllers.NewContribController(controllers.ContribControllerOpts{
		Common:     gui.controllerCommon,
		GetContext: func() *context.ContribContext { return gui.contexts.Contrib },
	})
	controllers.AttachController(ctrl)
}
